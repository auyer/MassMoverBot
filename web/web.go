package web

import (
	"log"
	"net/http"
	"os"

	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Handler interface defines the necessary methods for a Handler
type Handler interface {
	Refresh() http.Handler
	Login(http.ResponseWriter, *http.Request)
	Callback(http.ResponseWriter, *http.Request)
	Auth(http.Handler) http.Handler
	Guilds() http.Handler
	GuildByID() http.Handler
}

// Run fun
func Run(wh Handler, port string) error {

	router := mux.NewRouter()
	staticServerHandler := http.StripPrefix("/", http.FileServer(http.Dir("./build/")))

	router.HandleFunc("/api/login", wh.Login).Methods("GET")
	router.HandleFunc("/api/callback", wh.Callback)
	router.Path("/api/refresh").Handler(wh.Auth(wh.Refresh()))
	router.Path("/api/guilds").Handler(wh.Auth(wh.Guilds())).Methods("GET")
	router.Path("/api/guild/{id}").Handler(wh.Auth(wh.GuildByID())).Methods("GET")
	router.PathPrefix("/").Handler(staticServerHandler)

	log.Println("Listening on port ", port)
	err := http.ListenAndServe(port, gh.LoggingHandler(os.Stdout, router))
	if err != nil {
		return err
	}
	return nil
}
