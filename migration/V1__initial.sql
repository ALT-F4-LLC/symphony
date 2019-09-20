create table if not exists "image"
(
  filename    varchar(255) not null,
  id   uuid         not null
    constraint image_pkey
    primary key,
  name        varchar(255) not null
);
