-- psql notifier
-- drop table incoming;
-- drop table notifiers;

CREATE TABLE incoming(
  id serial primary key,
  class character varying(256),
  received_at timestamp,
  data json
);

CREATE table notifiers(
  id serial primary key,
  notification_type character varying(20),
  class character varying(256),
  template text,
  rules json
);

INSERT INTO notifiers(notification_type, class, template) VALUES ('email', 'User', 'User {{.name}} created with number: {{.number}}') RETURNING id;
INSERT INTO notifiers(notification_type, class, template) VALUES ('slack', 'User', 'User {{.name}} created with number: {{.number}} http://www.springest.nl/u/{{.number}}') RETURNING id;
