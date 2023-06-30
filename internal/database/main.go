package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

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

func InsertStreamer(channel string) error {
	db := connectToDB()
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
	
	streamers := []Streamer{}
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
		log.Errorf("Error while scanning DB rows for latest PID: %s", err.Error())
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
