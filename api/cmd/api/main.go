package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/dezhonger/dezhonger_service/api/internal/app"
	"golang.org/x/term"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := app.LoadConfig()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer application.Close()

	command := "serve"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	switch command {
	case "serve":
		return application.Run(ctx)
	case "create-admin":
		return createAdmin(ctx, application, os.Args[2:])
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func createAdmin(ctx context.Context, application *app.App, arguments []string) error {
	flags := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	username := flags.String("username", "admin", "administrator username")
	if err := flags.Parse(arguments); err != nil {
		return err
	}

	fmt.Fprint(os.Stderr, "Initial password (12-72 bytes): ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		if !errors.Is(err, syscall.ENOTTY) {
			return err
		}
		line, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
		if readErr != nil {
			return readErr
		}
		passwordBytes = []byte(strings.TrimSpace(line))
	}
	fmt.Fprintln(os.Stderr)
	user, err := application.CreateAdmin(ctx, *username, string(passwordBytes))
	for index := range passwordBytes {
		passwordBytes[index] = 0
	}
	if err != nil {
		return err
	}
	fmt.Printf("created administrator %s (%s); password change required on first login\n", user.Username, user.ID)
	return nil
}
