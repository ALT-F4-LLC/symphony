create table if not exists service
(
    hostname       text not null,
    id             uuid not null
        constraint service_pkey
            primary key,
    servicetype_id uuid not null
);

create table if not exists servicetype
(
    id   uuid         not null
        constraint servicetype_pkey
            primary key,
    name varchar(255) not null
        constraint servicetype_name_key
            unique
);

create unique index if not exists service_hostname_servicetype_id
    on service (hostname, servicetype_id);