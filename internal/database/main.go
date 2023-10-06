package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type TwitchChannel struct {
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
	if err != nil {
		log.Errorf("Error while inserting twitch message into the DB: %s", err.Error())
	}
	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing DB connection %s", err.Error())
	}
}

func closeDBConnection(conn *sql.DB) error {
	err := conn.Close()
	if err != nil {
		return err
	}
	return nil
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
	var db = connectToDB()
	sqlStatement := `
		INSERT INTO "twitch"."twitch_channels" (twitch_channel) 
		VALUES ($1) on conflict do nothing;
	`
	var _, err = db.Exec(sqlStatement, channel)
	if err != nil {
		log.Errorf("Error while trying to insert a streamer into the DB %s", err.Error())
	}
	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing the DB connection %s", err.Error())
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
	if err != nil {
		log.Errorf("Error while trying to update stream event status %s", err.Error())
		return err
	}

	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while trying to close the DB connection %s", err.Error())
		return err
	}
	return nil
}

func InsertStreamEventStatus(listening bool, pid int, twitchChannel string) error {
	db := connectToDB()
	sqlStatement := `
		insert into twitch.stream_events_status (event_id, listening, pid)
		select event_id, $1, $2 from public.stream_events_recent
		where twitch_channel = $3
	`
	_, err := db.Exec(sqlStatement, listening, pid, twitchChannel)
	if err != nil {
		log.Errorf("Error while inserting into the stream events status table %s", err.Error())
		return err
	}
	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing the DB connection %s", err.Error())
		return err
	}
	return err
}

func GetStreamerData() ([]TwitchChannel, error) {
	db := connectToDB()
	sqlStatement := `
		SELECT  "twitch_channel", "is_active" FROM (
		SELECT *, rank() OVER (PARTITION BY "twitch_channel" ORDER BY "time_added" DESC) rank_number from twitch.twitch_channels
		) AS t
		WHERE rank_number = 1
		`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streamers []TwitchChannel
	var streamer TwitchChannel
	for rows.Next() {
		err = rows.Scan(&streamer.Name, &streamer.Is_Active)
		if err != nil {
			return nil, err
		}
		streamers = append(streamers, streamer)
	}
	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing the DB conenction %s", err.Error())
		return nil, err
	}

	return streamers, err

}

func GetLatestPID(twitchChannel string) int {
	db := connectToDB()
	sqlStatement := `
	select pid  from public.stream_events_recent ser
	left join twitch.stream_events_status ses on ser.event_id = ses.event_id
	where twitch_channel = $1
	`
	var pid int
	rows := db.QueryRow(sqlStatement, twitchChannel)
	err := rows.Scan(&pid)
	if err != nil {
		log.Errorf("Error while scanning DB rows for latest PID: %s", err.Error())
	}

	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing the DB conenction %s", err.Error())
	}

	return pid
}

func UpsertStreamEvent(twitchChannel string) error {
	db := connectToDB()
	upsertStatement := `SELECT upsertStreamEvent($1);`

	_, err := db.Exec(upsertStatement, twitchChannel)
	if err != nil {
		return err
	}
	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing the DB conenction %s", err.Error())
		return err
	}

	return nil
}

func GetTwitchChannels() ([]string, error) {
	db := connectToDB()
	sqlStatement := `
		SELECT  "twitch_channel" FROM twitch.twitch_channels
	`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var twitchChannels []string
	var channel string
	for rows.Next() {
		_ = rows.Scan(&channel)
		twitchChannels = append(twitchChannels, channel)
	}
	err = closeDBConnection(db)
	if err != nil {
		log.Errorf("Error while closing the DB conenction %s", err.Error())
		return nil, err
	}

	return twitchChannels, err

}
