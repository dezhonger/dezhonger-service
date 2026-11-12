create table blog_posts (
  id uuid primary key default gen_random_uuid(),
  author_id uuid not null references users(id),
  title text not null default '' check (char_length(title) <= 200),
  blocks jsonb not null default '[]'::jsonb check (jsonb_typeof(blocks) = 'array' and jsonb_array_length(blocks) < 200),
  status text not null default 'draft' check (status in ('draft', 'published')),
  password_hash text not null default '',
  access_version bigint not null default 1,
  version bigint not null default 1,
  published_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  deleted_at timestamptz,
  check (status <> 'published' or (title <> '' and jsonb_array_length(blocks) > 0 and published_at is not null))
);
create index blog_posts_public_idx on blog_posts(published_at desc, id desc)
  where status = 'published' and deleted_at is null;

create table blog_media (
  id uuid primary key,
  post_id uuid not null references blog_posts(id),
  storage_name text not null unique,
  original_name text not null,
  mime_type text not null,
  size bigint not null check (size > 0 and size <= 209715200),
  created_at timestamptz not null default now()
);
create index blog_media_post_idx on blog_media(post_id);

create table blog_reader_grants (
  token_hash bytea not null,
  post_id uuid not null references blog_posts(id) on delete cascade,
  access_version bigint not null,
  expires_at timestamptz not null,
  primary key(token_hash, post_id)
);
create index blog_reader_grants_expiry_idx on blog_reader_grants(expires_at);
