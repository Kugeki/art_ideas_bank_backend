-- +goose Up
create table passwords
(
    id             serial
        constraint passwords_pk
            primary key,
    hash_base64    varchar not null,
    argon2_version integer not null,
    argon2_type    integer not null,
    salt_base64    varchar not null,
    argon2_time    integer not null,
    argon2_memory  integer not null,
    argon2_threads integer not null,
    argon2_keylen  integer not null
);

create table users
(
    id          serial
        constraint users_pk
            primary key,
    email       varchar not null
        constraint users_pk2
            unique,
    password_id integer not null
        constraint users_passwords_id_fk
            references passwords
);


-- +goose Down
drop table users;
drop table password;
