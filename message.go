package main

import (
	"fmt"
	"math/rand"
	"time"
)

type message struct {
	ID      string `json:"id"`
	AppName string `json:"app_name"`
}

func newMessage() *message {

	rand.Seed(time.Now().UTC().UnixNano())
	id := make([]byte, 7)
	rand.Read(id)

	return &message{
		ID: fmt.Sprintf("%x", id),
	}

}
