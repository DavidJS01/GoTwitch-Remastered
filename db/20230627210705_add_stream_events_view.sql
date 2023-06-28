-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW stream_events_recent AS SELECT * from (
			SELECT *, row_number() over (partition by se.twitch_channel order by time_added desc) rank_number 
			FROM twitch.stream_events se 
		) t
	 WHERE rank_number=1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW stream_events_recent;
-- +goose StatementEnd
