package main

import (
	"fmt"
	"math/rand"
	"time"
)

type app struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	CallbackURL         string   `json:"callback_url"`
	AuthorizedApprovers []string `json:"authorized_approvers"` //Approvers email addresses
}

func newApp() *app {

	rand.Seed(time.Now().UTC().UnixNano())
	id := make([]byte, 10)
	rand.Read(id)

	return &app{
		ID: fmt.Sprintf("%x", id),
	}

}

func (app *app) sendMessageApproval() {

}
