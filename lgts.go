package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type lgts struct {
	Apps            map[string]app     `json:"apps"`     //registered applications
	Messages        map[string]message `json:"messages"` // messages to process, removed once sent back to client
	ApprovalEmojis  []string           `json:"approval_emojis"`
	RejectionEmojis []string           `json:"reject_emojis"`
}

func newlgts() *lgts {

	return &lgts{
		Apps:            make(map[string]app),
		Messages:        make(map[string]message),
		ApprovalEmojis:  make([]string, 0),
		RejectionEmojis: make([]string, 0),
	}

}

func (lgts *lgts) getApps(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Apps)
}

func (lgts *lgts) registerApp(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	newapp := newApp()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newapp)
	if err != nil {
		log.Println(err)
		sendErrorResponse(w, 400, "could not decode json body", err)
		return
	}

	if newapp.Name == "" || newapp.CallbackURL == "" || len(newapp.AuthorizedApprovers) == 0 {
		err := errors.New("Invalid Parameters")
		log.Println(err)
		sendErrorResponse(w, 400, "you must supply name, callback_url, and authorized_approvers as params", err)
		return

	}

	newapp.Name = strings.TrimSpace(newapp.Name)
	newapp.CallbackURL = strings.TrimSpace(newapp.CallbackURL)
	for i := range newapp.AuthorizedApprovers {
		newapp.AuthorizedApprovers[i] = strings.TrimSpace(newapp.AuthorizedApprovers[i])
	}

	lgts.Apps[newapp.ID] = *newapp

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		ID                  string   `json:"id"`
		Name                string   `json:"name"`
		CallbackURL         string   `json:"callback_url"`
		AuthorizedApprovers []string `json:"authorized_approvers"`
		Token               string   `json:"token"`
	}{newapp.ID, newapp.Name, newapp.CallbackURL, newapp.AuthorizedApprovers, newapp.token})

	log.Printf("Application %s:%s registered", newapp.Name, newapp.ID)

}

func (lgts *lgts) unregisterApp(appID string) {
	delete(lgts.Apps, appID)
}

func (lgts *lgts) registerMessage(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {

	newMessage := newMessage()

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&newMessage)
	if err != nil {
		log.Println(err)
		sendErrorResponse(w, 400, "could not decode json body", err)
		return
	}

	if newMessage.AppID == "" {

		err := errors.New("Invalid Parameters")
		log.Println(err)
		sendErrorResponse(w, 400, "you must supply app_id", err)
		return

	}

	if !lgts.isApplicationRegistered(newMessage.AppID) {

		err := errors.New("Applicaiton not found")
		log.Println(err)
		sendErrorResponse(w, 404, "incorrect app_id provided; you must register your application first", err)
		return

	}

	lgts.Messages[newMessage.ID] = *newMessage

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newMessage)

	log.Printf("Message %s registered from application: %s", newMessage.ID, newMessage.AppID)

}

func (lgts *lgts) isApplicationRegistered(appID string) bool {

	for apps := range lgts.Apps {
		if apps == appID {
			return true
		}
	}
	return false
}

func (lgts *lgts) getMessages(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Messages)
}

func (lgts *lgts) isApprovalEmoji(emoji string) bool {

	for _, approvalEmoji := range lgts.ApprovalEmojis {
		if approvalEmoji == emoji {
			return true
		}
	}
	return false

}

func (lgts *lgts) isRejectionEmoji(emoji string) bool {

	for _, rejectEmoji := range lgts.RejectionEmojis {
		if rejectEmoji == emoji {
			return true
		}
	}
	return false

}
