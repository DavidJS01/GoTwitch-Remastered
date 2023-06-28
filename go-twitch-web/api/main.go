package main

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"test.com/m/internal/database"
)

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
	log.Print(streamer)
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

func main() {
	// handle func is convienance method on http. registers function to a path
	// on default serve mux.
	mux := mux.NewRouter().StrictSlash(true)
	mux.HandleFunc("/about", homeHandler)
	mux.HandleFunc("/stream/upsert", func(w http.ResponseWriter, r *http.Request) {
		upsertStreamerHandler(w, r)
	}).Queries("stream", "{stream}").Methods("POST")
	mux.HandleFunc("/stream/list", listStreamersHandler).Methods("GET")
	mux.HandleFunc("/stream/add", addStreamerHandler).Queries("stream", "{stream}").Methods("POST")
	http.ListenAndServe(":8080", mux)

}
