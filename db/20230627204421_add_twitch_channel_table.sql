-- +goose Up
-- +goose StatementBegin
create schema if not exists twitch;
		create table if not exists twitch.twitch_channels  (
			twitch_channel varchar primary key
		);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists twitch.twitch_channels;

drop schema if exists twitch;
-- +goose StatementEnd
