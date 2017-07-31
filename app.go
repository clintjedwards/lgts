package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

type app struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	CallbackURL         string   `json:"callback_url"`
	AuthorizedApprovers []string `json:"authorized_approvers"` //Approvers email addresses
	token               string
}

func newApp() *app {

	rand.Seed(time.Now().UTC().UnixNano())
	id := make([]byte, 10)
	token := make([]byte, 10)
	rand.Read(id)
	rand.Read(token)

	return &app{
		ID:    fmt.Sprintf("%x", id),
		token: fmt.Sprintf("%x", token),
	}

}

func (app *app) isAuthorizedUser(email string) bool {

	for _, approvedEmail := range app.AuthorizedApprovers {
		if approvedEmail == email {
			return true
		}
	}
	return false

}

func (app *app) sendMessageApproval(callbackInfo map[string]interface{}, approved bool) error {
	jsonString, _ := json.Marshal(callbackInfo)

	_, err := http.Post(app.CallbackURL, "application/json", bytes.NewBuffer(jsonString))
	if err != nil {
		return err
	}

	return nil
}
