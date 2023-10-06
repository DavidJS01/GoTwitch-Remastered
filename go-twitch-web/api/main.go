package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"test.com/m/internal/database"
)

// leaving this here for reference on templates, will be implemented in another dir
const doc = `
<!DOCTYPE html>
<html>
	<head>
		<title>Streamers</title>
	</head>
	<body>
		<h3>Streamers:</h3>
		{{range .}}
			<li>{{.}}</li>
		{{end}}
	</body>
</html>
`

func upsertStreamerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	streamer := mux.Vars(r)["stream"]
	err := database.UpsertStreamEvent(streamer)
	log.Print(err)
	if err == nil {
		w.WriteHeader(200)
		return
	}
	if err != nil {
		http.Error(w, "Error while adding a stream event", http.StatusInternalServerError)
	}
}

func addStreamerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	streamer := mux.Vars(r)["stream"]
	err := database.InsertStreamer(streamer)
	if err == nil {
		w.WriteHeader(200)
		return
	}
	if err != nil {
		http.Error(w, "Error while adding a streamer", http.StatusInternalServerError)
	}
}

func listStreamersHandler(w http.ResponseWriter, r *http.Request) {
	streamerData, err := database.GetStreamerData()
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(streamerData)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "assets/index.html")
}

func renderTwitchChannelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content Type", "text/html")
	templates := template.New("template")
	// "doc" is the constant that holds the HTML content
	templates.New("doc").Parse(doc)
	var channels, _ = database.GetTwitchChannels()
	templates.Lookup("doc").Execute(w, channels)

}

func getTwitchChannelsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var channels, _ = database.GetTwitchChannels()
	json.NewEncoder(w).Encode(channels)
}

func main() {
	// handle func is convienance method on http. registers function to a path
	// on default serve mux.
	mux := mux.NewRouter().StrictSlash(true)
	mux.HandleFunc("/about", homeHandler)
	mux.HandleFunc("/stream/add", addStreamerHandler).Queries("stream", "{stream}").Methods("POST")
	mux.HandleFunc("/stream/list", getTwitchChannelsHandler).Methods("GET")
	mux.HandleFunc("/stream/list/render", renderTwitchChannelsHandler).Methods("GET")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Print(err)
	}

}
