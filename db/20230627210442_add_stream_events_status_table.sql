-- +goose Up
-- +goose StatementBegin
create table if not exists twitch.stream_events_status  (
			id serial primary key,
			event_id integer references twitch.stream_events(event_id) not null,
			listening boolean not null,
			pid integer not null,
			time_added timestamp without time zone default (now() at time zone 'utc') not null
		);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists twitch.stream_events_status;
-- +goose StatementEnd
