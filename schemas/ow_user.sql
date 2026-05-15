create table users
(
    user_id       bigserial primary key,
    email         text      not null unique,
    username      text      not null unique,
    password_hash text      not null default '',
    sport_goals   text[]    not null default '{}',
    gender        text,
    birthdate     text,
    split         jsonb     not null default '{}'
);

create table refresh_tokens
(
    id         bigserial primary key,
    user_id    bigint      not null references users (user_id) on delete cascade,
    token_hash text        not null unique,
    expires_at timestamptz not null,
    created_at timestamptz not null default now()
);
