package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type lgts struct {
	Apps            map[string]app     `json:"apps"`     //registered applications
	Messages        map[string]message `json:"messages"` // messages to process, removed once sent back to client
	ApprovalEmojis  []string           `json:"approval_emojis"`
	RejectionEmojis []string           `json:"reject_emojis"`
	stateFilePath   string
}

func newlgts(stateFilePath string) *lgts {

	return &lgts{
		Apps:            make(map[string]app),
		Messages:        make(map[string]message),
		ApprovalEmojis:  []string{"lgts", "lgtm"},
		RejectionEmojis: []string{"lbts", "duckno"},
		stateFilePath:   stateFilePath,
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
		err := fmt.Errorf("Invalid Parameters")
		log.Println(err)
		sendErrorResponse(w, 400, "you must supply name, callback_url, and authorized_approvers as params", err)
		return

	}

	newapp.Name = strings.ToLower(strings.TrimSpace(newapp.Name))
	newapp.CallbackURL = strings.TrimSpace(newapp.CallbackURL)
	for i := range newapp.AuthorizedApprovers {
		newapp.AuthorizedApprovers[i] = strings.ToLower(strings.TrimSpace(newapp.AuthorizedApprovers[i]))
	}

	if _, present := lgts.Apps[newapp.Name]; present {
		err := fmt.Errorf("Application %s exists", newapp.Name)
		sendErrorResponse(w, 400, "application already registered", err)
		return
	}

	lgts.Apps[newapp.Name] = *newapp

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Name                string   `json:"name"`
		CallbackURL         string   `json:"callback_url"`
		AuthorizedApprovers []string `json:"authorized_approvers"`
		Token               string   `json:"token"`
	}{newapp.Name, newapp.CallbackURL, newapp.AuthorizedApprovers, newapp.Token})

	log.Printf("Application %s registered", newapp.Name)

	lgts.writeState()

}

func (lgts *lgts) unregisterApp(w http.ResponseWriter, req *http.Request, params httprouter.Params) {

	token := req.Header.Get("Authorization")
	appName := params.ByName("name")

	if _, present := lgts.Apps[appName]; !present {
		err := fmt.Errorf("Application %s does not exist", appName)
		sendErrorResponse(w, 400, "application not registered", err)
		return
	}

	if token != lgts.Apps[appName].Token {
		err := fmt.Errorf("Incorrect token for Application %s", appName)
		sendErrorResponse(w, 401, "Unauthorized", err)
		return
	}

	delete(lgts.Apps, appName)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lgts.Apps)
}

func (lgts *lgts) updateApp(w http.ResponseWriter, req *http.Request, params httprouter.Params) {

	token := req.Header.Get("Authorization")
	appName := params.ByName("name")

	if _, present := lgts.Apps[appName]; !present {
		err := fmt.Errorf("Application %s does not exist", appName)
		sendErrorResponse(w, 400, "application not registered", err)
		return
	}

	if token != lgts.Apps[appName].Token {
		err := fmt.Errorf("Incorrect token for Application %s", appName)
		sendErrorResponse(w, 401, "Unauthorized", err)
		return
	}

	app := lgts.Apps[appName]

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&app)
	if err != nil {
		log.Println(err)
		sendErrorResponse(w, 400, "could not decode json body", err)
		return
	}

	lgts.Apps[appName] = app

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(app)

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

	if newMessage.AppName == "" {

		err := fmt.Errorf("Invalid Parameters")
		log.Println(err)
		sendErrorResponse(w, 400, "you must supply an app_name", err)
		return

	}

	if !lgts.isApplicationRegistered(newMessage.AppName) {

		err := fmt.Errorf("Application not found")
		log.Println(err)
		sendErrorResponse(w, 404, "incorrect app_name provided; you must register your application first", err)
		return

	}

	lgts.Messages[newMessage.ID] = *newMessage

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newMessage)

	log.Printf("Message %s registered from application: %s", newMessage.ID, newMessage.AppName)

}

func (lgts *lgts) isApplicationRegistered(appName string) bool {

	for app := range lgts.Apps {
		if app == appName {
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

// writeState persists current Apps map state so that the process
// can restart without losing tokens
func (lgts *lgts) writeState() {

	file, err := os.Create(lgts.stateFilePath)
	if err != nil {
		log.Printf("could not save state: %v", err)
		return
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	encoder.Encode(lgts.Apps)

}

func (lgts *lgts) loadState() {
	file, err := os.Open(lgts.stateFilePath)
	if err != nil {
		log.Printf("could not load state: %v", err)
		return
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&lgts.Apps)
	if err != nil {
		log.Printf("could not load state: %v", err)
		return
	}
}
