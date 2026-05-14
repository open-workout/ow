create table users
(
    user_id     bigserial primary key,
    email       text      not null unique,
    sport_goals text[]    not null default '{}',
    gender      text,
    birthdate   text,
    split       jsonb     not null default '{}'
);
