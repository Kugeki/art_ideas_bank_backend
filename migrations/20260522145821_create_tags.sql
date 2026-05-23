-- +goose Up
create extension if not exists ltree;

create table tags
(
    id      uuid primary key default gen_random_uuid(),
    user_id integer not null references users (id) on delete cascade,
    path    ltree   not null,
    name    text    not null
);

create unique index idx_tags_user_path on tags (user_id, path);
create index idx_tags_path_gist on tags using gist (path);

create table image_tags
(
    image_id uuid not null references images (id) on delete cascade,
    tag_id   uuid not null references tags (id) on delete cascade,
    primary key (image_id, tag_id)
);

-- +goose Down
drop table if exists image_tags;
drop table if exists tags;
