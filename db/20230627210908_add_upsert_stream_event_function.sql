-- +goose Up
-- +goose StatementBegin
create or replace function upsertStreamEvent(streamEventChannelName varchar) returns void
	language plpgsql
		AS $$
			BEGIN
			IF streamEventChannelName in (SELECT twitch_channel FROM twitch.stream_events) then
				update twitch.stream_events se
        set current_flag = false
        where se.event_id = (
					select event_id from stream_events_recent
					WHERE rank_number=1
					and twitch_channel = streamEventChannelName
				);
			
				insert into twitch.stream_events (twitch_channel, current_flag)
					select twitch_channel, true as current_flag  from stream_events_recent
					WHERE rank_number=1
					and twitch_channel = streamEventChannelName;
			ELSE
				INSERT INTO twitch.stream_events values (default, streamEventChannelName, true);
			END IF;
			END;
		$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION upsertStreamEvent;
-- +goose StatementEnd
