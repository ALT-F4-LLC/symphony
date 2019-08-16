create table if not exists "image"
(
  id   uuid         not null
    constraint image_pkey
    primary key,
  name varchar(255) not null
    constraint image_name_key
    unique
);
