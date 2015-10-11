-- psql notifier
-- drop table notifiers;

CREATE table notifiers(
  id serial primary key,
  event_name character varying(256),
  template text,
  rules json,
  notification_type character varying(20),
  target character varying(256)
);
