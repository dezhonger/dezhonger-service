package app

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	databaseBrowserDefaultLimit = 25
	databaseBrowserMaxLimit     = 100
	databaseBrowserTextLimit    = 12_000
)

type databaseTable struct {
	Name     string `json:"name"`
	RowCount int64  `json:"row_count"`
}

type databaseColumn struct {
	Name       string  `json:"name"`
	DataType   string  `json:"data_type"`
	UDTName    string  `json:"udt_name"`
	Nullable   bool    `json:"nullable"`
	Default    *string `json:"default,omitempty"`
	PrimaryKey bool    `json:"primary_key"`
	Sensitive  bool    `json:"sensitive"`
}

func (a *App) listDatabaseTables(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	rows, err := a.db.Query(ctx, `
		select table_name
		from information_schema.tables
		where table_schema = 'public' and table_type = 'BASE TABLE'
		order by table_name
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to inspect database")
		return
	}
	defer rows.Close()

	tables := make([]databaseTable, 0)
	for rows.Next() {
		var table databaseTable
		if err := rows.Scan(&table.Name); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to inspect database")
			return
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to inspect database")
		return
	}

	for index := range tables {
		query := "select count(*) from " + pgx.Identifier{"public", tables[index].Name}.Sanitize()
		if err := a.db.QueryRow(ctx, query).Scan(&tables[index].RowCount); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to count database rows")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"schema": "public", "tables": tables})
}

func (a *App) inspectDatabaseTable(w http.ResponseWriter, r *http.Request) {
	tableName := r.PathValue("table")
	limit, offset, err := parseDatabasePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_pagination", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	exists, err := a.databaseTableExists(ctx, tableName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to inspect database")
		return
	}
	if !exists {
		writeError(w, http.StatusNotFound, "table_not_found", "database table not found")
		return
	}

	columns, err := a.databaseColumns(ctx, tableName)
	if err != nil {
		slog.Error("database browser failed to read columns", "table", tableName, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to inspect database columns")
		return
	}

	identifier := pgx.Identifier{"public", tableName}.Sanitize()
	var total int64
	if err := a.db.QueryRow(ctx, "select count(*) from "+identifier).Scan(&total); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to count database rows")
		return
	}

	query := "select * from " + identifier
	primaryKeys := make([]string, 0)
	for _, column := range columns {
		if column.PrimaryKey {
			primaryKeys = append(primaryKeys, column.Name)
		}
	}
	if len(primaryKeys) > 0 {
		parts := make([]string, 0, len(primaryKeys))
		for _, name := range primaryKeys {
			parts = append(parts, pgx.Identifier{name}.Sanitize())
		}
		query += " order by " + strings.Join(parts, ", ")
	}
	query += " limit $1 offset $2"

	rows, err := a.db.Query(ctx, query, limit, offset)
	if err != nil {
		slog.Error("database browser failed to read rows", "table", tableName, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to read database rows")
		return
	}
	defer rows.Close()

	data := make([]map[string]any, 0)
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "unable to read database rows")
			return
		}
		row := make(map[string]any, len(columns))
		for index, column := range columns {
			if column.Sensitive {
				row[column.Name] = "[hidden]"
				continue
			}
			row[column.Name] = databaseJSONValue(values[index])
		}
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "unable to read database rows")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"schema":  "public",
		"table":   tableName,
		"columns": columns,
		"rows":    data,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (a *App) databaseTableExists(ctx context.Context, tableName string) (bool, error) {
	if tableName == "" || len(tableName) > 63 {
		return false, nil
	}
	var exists bool
	err := a.db.QueryRow(ctx, `
		select exists(
			select 1 from information_schema.tables
			where table_schema = 'public' and table_type = 'BASE TABLE' and table_name = $1
		)
	`, tableName).Scan(&exists)
	return exists, err
}

func (a *App) databaseColumns(ctx context.Context, tableName string) ([]databaseColumn, error) {
	rows, err := a.db.Query(ctx, `
		select c.column_name, c.data_type, c.udt_name, c.is_nullable = 'YES', coalesce(c.column_default, ''),
		       exists(
		         select 1
		         from pg_index i
		         join pg_attribute a on a.attrelid = i.indrelid and a.attnum = any(i.indkey)
		         where i.indrelid = format('%I.%I', 'public', $1::text)::regclass
		           and i.indisprimary and a.attname = c.column_name
		       ) as primary_key
		from information_schema.columns c
		where c.table_schema = 'public' and c.table_name = $1::text
		order by c.ordinal_position
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make([]databaseColumn, 0)
	for rows.Next() {
		var column databaseColumn
		var defaultValue string
		if err := rows.Scan(
			&column.Name, &column.DataType, &column.UDTName, &column.Nullable,
			&defaultValue, &column.PrimaryKey,
		); err != nil {
			return nil, err
		}
		if defaultValue != "" {
			column.Default = &defaultValue
		}
		column.Sensitive = isSensitiveDatabaseColumn(column.Name)
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func parseDatabasePagination(r *http.Request) (int, int, error) {
	limit, offset := databaseBrowserDefaultLimit, 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > databaseBrowserMaxLimit {
			return 0, 0, fmt.Errorf("limit must be between 1 and %d", databaseBrowserMaxLimit)
		}
		limit = value
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, errors.New("offset must be zero or greater")
		}
		offset = value
	}
	return limit, offset, nil
}

func isSensitiveDatabaseColumn(name string) bool {
	normalized := strings.ToLower(name)
	if normalized == "must_change_password" {
		return false
	}
	for _, fragment := range []string{"password_hash", "token_hash", "access_token", "refresh_token", "secret", "credential", "private_key"} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func databaseJSONValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case time.Time:
		return typed.UTC().Format(time.RFC3339Nano)
	case uuid.UUID:
		return typed.String()
	case [16]byte:
		return uuid.UUID(typed).String()
	case []byte:
		encoded := hex.EncodeToString(typed)
		if len(encoded) > databaseBrowserTextLimit {
			encoded = encoded[:databaseBrowserTextLimit] + "…"
		}
		return "\\x" + encoded
	case string:
		if utf8.RuneCountInString(typed) <= databaseBrowserTextLimit {
			return typed
		}
		runes := []rune(typed)
		return string(runes[:databaseBrowserTextLimit]) + "…"
	default:
		return typed
	}
}
