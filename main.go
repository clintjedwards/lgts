// lgts (looks good to slack) is a web service which
// allows applications to register messages to look out for,
// listens to the slack event stream and then reports back
// when certain emojis are used by certain users on those
// predetermined messages

// Provides a service much like interactive messages.
// Useful for instances where public callbacks are impossible

package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/clintjedwards/snark/helpers/httputil"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	app := *newApp()
	go app.runrtm()

	router.Handle("/track", handlers.MethodHandler{
		"POST": http.HandlerFunc(app.createMessageHandler),
	})

	router.Handle("/track/{messageID}", handlers.MethodHandler{
		"DELETE": http.HandlerFunc(app.deleteMessageHandler),
	})

	server := http.Server{
		Addr:         app.config.ServerURL,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	server.Handler = httputil.DefaultHeaders(router)
	if app.config.Debug {
		server.Handler = handlers.LoggingHandler(os.Stdout, server.Handler)
	}

	log.Printf("Starting lgts server on %s\n", app.config.ServerURL)
	log.Fatal(http.ListenAndServe(app.config.ServerURL, router))
}
