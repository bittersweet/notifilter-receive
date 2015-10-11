CREATE TABLE incoming(
  id serial primary key,
  received_at timestamp,
  data json
);
