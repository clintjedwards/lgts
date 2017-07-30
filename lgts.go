package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type lgts struct {
	Apps           map[string]app  `json:"apps"`     //registered applications
	Messages       map[int]message `json:"messages"` // messages to process, removed once sent back to client
	ApprovalEmojis []string        `json:"approval_emojis"`
	RejectEmojis   []string        `json:"reject_emojis"`
}

func newlgts() *lgts {

	return &lgts{
		Apps:     make(map[string]app),
		Messages: make(map[int]message),
	}

}

func (lgts *lgts) getApps(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Apps)
}

func (lgts *lgts) registerApp(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	newapp := newApp()

	var appinfo app

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&appinfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}{false, "could not decode json body", fmt.Sprintf("%v", err)})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if appinfo.Name == "" || appinfo.CallbackURL == "" || len(appinfo.AuthorizedApprovers) == 0 {

		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Ok      bool   `json:"ok"`
			Message string `json:"message"`
		}{
			false,
			"Invalid Parameters: You must supply name, callbackurl, authorizedapprovers as params",
		})

	} else {

		newapp.Name = strings.TrimSpace(appinfo.Name)
		newapp.CallbackURL = strings.TrimSpace(appinfo.CallbackURL)
		for i := range newapp.AuthorizedApprovers {
			newapp.AuthorizedApprovers[i] = strings.TrimSpace(appinfo.AuthorizedApprovers[i])
		}

		lgts.Apps[newapp.ID] = *newapp

		json.NewEncoder(w).Encode(struct {
			ID   string `json:"app_id"`
			Name string `json:"app_name"`
		}{newapp.ID, newapp.Name})

		log.Printf("Application %s:%s registered", newapp.Name, newapp.ID)
	}

}

func (lgts *lgts) unregisterApp(appID string) {
	delete(lgts.Apps, appID)
}

func (lgts *lgts) registerMessage(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	var messageinfo message

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&messageinfo)
	if err != nil {
		log.Println(err)
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}{"invalid parameters", "could not decode json body"})
		return
	}

	if messageinfo.AppID == "" || messageinfo.Token == "" {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}{"Invalid Parameters", "You must supply both message_id and message_token params"})
	} else {

		lgts.Messages[messageinfo.AppID] = messageinfo

		json.NewEncoder(w).Encode(struct {
			SuccessCode string `json:"success_code"`
			Message     string `json:"message"`
		}{"200", "Message registered"})

		log.Printf("Message %d registered", messageinfo.ID)
	}
}

func (lgts *lgts) getMessages(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Messages)
}
