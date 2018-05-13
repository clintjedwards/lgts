// snark is a web service which
// allows applications to register messages to look out for,
// listens to the slack event stream and then reports back
// when certain emojis are used by certain users on those
// predetermined messages

// Provides a service much like native slack interactive messages.
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
		"GET":    http.HandlerFunc(app.getMessageHandler),
		"DELETE": http.HandlerFunc(app.deleteMessageHandler),
	})

	router.Handle("/version", handlers.MethodHandler{
		"GET": http.HandlerFunc(app.getVersionHandler),
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

	log.Printf("Starting snark server on %s\n", app.config.ServerURL)
	log.Fatal(http.ListenAndServe(app.config.ServerURL, server.Handler))
}
