package main

import (
	"fmt"
	"math/rand"
	"time"
)

type message struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`
}

func newMessage() *message {

	rand.Seed(time.Now().UTC().UnixNano())
	id := make([]byte, 7)
	rand.Read(id)

	return &message{
		ID: fmt.Sprintf("%x", id),
	}

}
