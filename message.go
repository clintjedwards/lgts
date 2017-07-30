package main

type message struct {
	AppID string `json:"app_id"`
	Token string `json:"message_token"` //token sent by client to prevent mimics
}
