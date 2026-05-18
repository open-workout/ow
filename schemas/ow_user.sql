create table users
(
    user_id     text primary key,
    email       text not null unique,
    username    text not null unique,
    sport_goals text[]    not null default '{}',
    gender      text,
    birthdate   text,
    split       jsonb     not null default '{}'
);
