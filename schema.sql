-- psql notifier
-- drop table notifiers;

CREATE table notifiers(
  id serial primary key,
  application character varying(256),
  event_name character varying(256),
  template text,
  rules json,
  notification_type character varying(20),
  target character varying(256)
);

CREATE INDEX index_application_event_name ON notifiers (application, event_name)
