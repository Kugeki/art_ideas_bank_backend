-- +goose Up
create table images
(
    id          uuid primary key     default gen_random_uuid(),
    user_id     integer     not null references users (id) on delete cascade ,
    ext         text        not null,
    s3_key      text        not null,
    uploaded_at timestamptz not null default now()
);

-- +goose Down
drop table images;
