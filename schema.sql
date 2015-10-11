-- psql notifier
-- drop table incoming;
-- drop table notifiers;

CREATE TABLE incoming(
  id serial primary key,
  received_at timestamp,
  data json
);

CREATE table notifiers(
  id serial primary key,
  class character(256),
  template text
)
