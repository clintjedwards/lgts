package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/clintjedwards/lgts/config"
	"github.com/clintjedwards/lgts/storage"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/nlopes/slack"
)

//We define two slackClients because each token has different permissions
// that do not and cannot overlap
type app struct {
	config         *config.Config
	db             *pg.DB
	slackAppClient *slack.Client
	slackBotClient *slack.Client
	messages       map[string]map[string]message //Make a map of service names that has a map of message ids
}

func newApp() *app {

	config, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db := storage.NewPostgresDB(config.Database.User, config.Database.Password, config.Database.URL, config.Database.Name)
	err = storage.InitDB(db, []interface{}{&service{}})
	if err != nil {
		log.Println(err)
		log.Fatalf("Cannot connect to database %s with user %s", config.Database.URL, config.Database.User)
	}

	if config.Debug {
		db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
			query, err := event.FormattedQuery()
			if err != nil {
				panic(err)
			}

			log.Printf("%s %s", time.Since(event.StartTime), query)
		})
	}

	log.Printf("Connected to Database: %s@%s/%s",
		config.Database.User, config.Database.URL, config.Database.Name)

	slackAppClient := slack.New(config.Slack.AppToken)
	slackBotClient := slack.New(config.Slack.BotToken)

	slackAppClient.SetDebug(config.Debug)
	slackBotClient.SetDebug(config.Debug)

	return &app{
		config:         config,
		db:             db,
		slackAppClient: slackAppClient,
		slackBotClient: slackBotClient,
		messages:       make(map[string]map[string]message),
	}

}

func (app *app) checkAuthorizationHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		token := request.Header.Get("Authorization")

		service, err := app.getService(vars["name"])
		if err == errServiceNotFound {
			sendResponse(writer, http.StatusNotFound, fmt.Sprintf("%s: %s", errServiceNotFound.Error(), vars["name"]), true)
			return
		} else if err != nil {
			log.Println(err)
			sendResponse(writer, http.StatusInternalServerError, "could not retrieve service", true)
			return
		}

		if token != service.AuthToken {
			err := fmt.Errorf("Incorrect token for service %s", service.Name)
			sendResponse(writer, http.StatusUnauthorized, err.Error(), true)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
