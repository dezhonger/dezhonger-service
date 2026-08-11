create extension if not exists pgcrypto;

create table if not exists users (
  id uuid primary key default gen_random_uuid(),
  username text not null unique,
  password_hash text not null,
  role text not null default 'user' check (role in ('admin', 'user')),
  is_active boolean not null default true,
  must_change_password boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  check (username ~ '^[a-z0-9][a-z0-9._-]{2,31}$')
);

create table if not exists sessions (
  token_hash bytea primary key,
  user_id uuid not null references users(id) on delete cascade,
  expires_at timestamptz not null,
  created_at timestamptz not null default now()
);

create index if not exists sessions_user_id_idx on sessions(user_id);
create index if not exists sessions_expires_at_idx on sessions(expires_at);

create table if not exists notes (
  id uuid primary key,
  user_id uuid not null references users(id) on delete cascade,
  title text not null default '' check (char_length(title) <= 200),
  content text not null default '' check (char_length(content) <= 2000000),
  tags text[] not null default '{}'::text[] check (cardinality(tags) <= 20),
  category text not null default '' check (char_length(category) <= 50),
  status text not null default 'inbox' check (status in ('inbox', 'todo', 'doing', 'done', 'archived')),
  is_pinned boolean not null default false,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists notes_user_updated_at_idx
  on notes(user_id, is_pinned desc, updated_at desc);
