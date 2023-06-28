-- +goose Up
-- +goose StatementBegin
create table if not exists twitch.stream_events  (
			event_id SERIAL primary KEY,
			twitch_channel varchar references twitch.twitch_channels(twitch_channel) not null,
			current_flag boolean not null,
			time_added timestamp without time zone default (now() at time zone 'utc') not null
		);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists twitch.stream_events;
-- +goose StatementEnd
