create table if not exists pv
(
    device     text not null,
    id         uuid not null
        constraint pv_pkey
            primary key,
    service_id uuid not null
);

create unique index if not exists pv_device_service_id
    on pv (device, service_id);