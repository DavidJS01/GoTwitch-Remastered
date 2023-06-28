package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type Streamer struct {
	Name      string
	Is_Active bool
}

const (
	host     = "172.17.0.1"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

type InsertMessage func(username string, message string, channel string)

func InsertTwitchMessage(username string, message string, channel string) {
	db := connectToDB()
	sqlStatement := `
				INSERT INTO "postgres"."twitch"."stream_messages" (twitch_channel, username, message)
				VALUES ($1, $2, $3)`
	_, err := db.Exec(sqlStatement, channel, username, message)
	db.Close()
	if err != nil {
		panic(err)
	}
}


func connectToDB() *sql.DB {
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return db
}

func createUpsertStreamEventFunction() {
	db := connectToDB()
	sqlStatement := `
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
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
	db.Close()

}

func createMessageTable() {
	db := connectToDB()
	sqlStatement := `
		create schema if not exists twitch;
		create table if not exists twitch.stream_messages  (
			id serial primary key,
			twitch_channel varchar references twitch.twitch_channels(twitch_channel) not null,
			username varchar not null,
			message varchar not null,
			message_timestamp timestamp without time zone default (now() at time zone 'utc') not null 
		);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
	db.Close()

}

func createStreamerTable() {
	db := connectToDB()
	sqlStatement := `
		create schema if not exists twitch;
		create table if not exists twitch.twitch_channels  (
			twitch_channel varchar primary key
		);
	`
	_, err := db.Exec(sqlStatement)
	if err != nil {
		panic(err)
	}
	db.Close()
}

func createStreamEventsTable() {
	db := connectToDB()
	sqlStatement := `
		create schema if not exists twitch;
		create table if not exists twitch.stream_events  (
			event_id SERIAL primary KEY,
			twitch_channel varchar references twitch.twitch_channels(twitch_channel) not null,
			current_flag boolean not null,
			time_added timestamp without time zone default (now() at time zone 'utc') not null
		);
	`
	_, err := db.Exec(sqlStatement)
	db.Close()
	if err != nil {
		panic(err)
	}
}

func createStreamEventsView() {
	db := connectToDB()
	sqlStatement := `
		CREATE OR REPLACE VIEW stream_events_recent AS SELECT * from (
			SELECT *, row_number() over (partition by se.twitch_channel  order by time_added desc) rank_number 
			FROM twitch.stream_events se 
		) t
	 WHERE rank_number=1;
	`
	_, err := db.Exec(sqlStatement)
	db.Close()
	if err != nil {
		panic(err)
	}
}

func createStreamEventsStatusTable() {
	db := connectToDB()
	sqlStatement := `
		create schema if not exists twitch;
		create table if not exists twitch.stream_events_status  (
			id serial primary key,
			event_id integer references twitch.stream_events(event_id) not null,
			listening boolean not null,
			pid integer not null,
			time_added timestamp without time zone default (now() at time zone 'utc') not null
		);
	`
	_, err := db.Exec(sqlStatement)
	db.Close()
	if err != nil {
		panic(err)
	}
}

func InsertStreamer(channel string) error {
	db := connectToDB()
	log.Print(channel)
	sqlStatement := `
		INSERT INTO "twitch"."twitch_channels" (twitch_channel) 
		VALUES ($1) on conflict do nothing;
	`
	_, err := db.Exec(sqlStatement, channel)
	db.Close()
	if err != nil {
		panic(err)
	}
	return err
}

func UpdateStreamEventStatus(pid int, twitchChannel string) error {
	db := connectToDB()
	sqlStatement := `
	update twitch.stream_events_status ses set listening = false
	from public.stream_events_recent ser 
	where ser.event_id = ses.event_id
	and ser.twitch_channel = $2
	and ses.pid = $1
	and ses.time_added = (select max(time_added)from twitch.stream_events_status where pid = $1 and twitch_channel = $2)
	`
	_, err := db.Exec(sqlStatement, pid, twitchChannel)
	db.Close()
	if err != nil {
		panic(err)
	}
	return err
}

func InsertStreamEventStatus(listening bool, pid int, twitchChannel string) error {
	db := connectToDB()
	sqlStatement := `
		insert into twitch.stream_events_status (event_id, listening, pid)
		select event_id, $1, $2 from public.stream_events_recent
		where twitch_channel = $3
	`
	_, err := db.Exec(sqlStatement, listening, pid, twitchChannel)
	db.Close()
	if err != nil {
		panic(err)
	}
	return err
}

func GetStreamerData() ([]Streamer, error) {
	streamers := []Streamer{}
	db := connectToDB()
	sqlStatement := `
		SELECT  "twitch_channel", "is_active" FROM (
		SELECT *, rank() OVER (PARTITION BY "twitch_channel" ORDER BY "time_added" DESC) rank_number from twitch.streamers
		) AS t
		WHERE rank_number = 1
		`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var streamer Streamer
	for rows.Next() {
		err = rows.Scan(&streamer.Name, &streamer.Is_Active)
		streamers = append(streamers, streamer)

	}
	db.Close()
	if err != nil {
		panic(err)
	}
	return streamers, err

}

func GetLatestPID(streamer string) int {
	db := connectToDB()
	sqlStatement := `
	select pid  from public.stream_events_recent ser
	left join twitch.stream_events_status ses on ser.event_id = ses.event_id
	where twitch_channel = $1
	`
	var pid int
	rows := db.QueryRow(sqlStatement, streamer)
	err := rows.Scan(&pid)
	if err != nil {
		log.Print("L")
	}

	db.Close()
	if err != nil {
		panic(err)
	}
	return pid
}

func UpsertStreamEvent(twitchChannel string) error {
	db := connectToDB()
	upsertStatement := `SELECT upsertStreamEvent($1);`
	_, err := db.Exec(upsertStatement, twitchChannel)
	if err != nil {
		panic(err)
	}
	db.Close()
	return err
}

func SetupPostgres() {
	createStreamerTable()
	createStreamEventsTable()
	createMessageTable()
	createStreamEventsStatusTable()
	createStreamEventsView()
	createUpsertStreamEventFunction()
}
