create table if not exists image
(
    filename text not null,
    id       uuid not null
        constraint image_pkey
            primary key,
    name     text not null
);
