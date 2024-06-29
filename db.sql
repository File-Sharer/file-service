CREATE TABLE files
(
  id varchar(28) primary key,
  creator_id varchar(16) not null,
  is_public bool default false,
  date_added timestamp(0) without time zone default current_timestamp
);

CREATE TABLE permissions
(
  file_id varchar(28) not null,
  user_id varchar(16) not null,
  UNIQUE (file_id, user_id)
);
