-- +goose Up
-- +goose StatementBegin
		create table if not exists twitch.stream_messages  (
			id serial primary key,
			twitch_channel varchar references twitch.twitch_channels(twitch_channel) not null,
			username varchar not null,
			message varchar not null,
			message_timestamp timestamp without time zone default (now() at time zone 'utc') not null 
		);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
		drop table if exists twitch.stream_messages;
		
-- +goose StatementEnd
