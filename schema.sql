-- psql notifier
-- drop table incoming;
-- drop table notifiers;

CREATE TABLE incoming(
  id serial primary key,
  class character(256),
  received_at timestamp,
  data json
);

CREATE table notifiers(
  id serial primary key,
  notification_type character(20),
  class character(256),
  template text
);

INSERT INTO notifiers(notification_type, class, template) VALUES ('email', 'User', 'User {{.name}} created with number: {{.number}}') RETURNING id;
