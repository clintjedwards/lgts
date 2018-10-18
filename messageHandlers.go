package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/gorilla/mux"

	"github.com/clintjedwards/snark/helpers/httputil"
	validation "github.com/go-ozzo/ozzo-validation"
)

func (app *app) getMessageHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	trackedMessage, err := app.getMessage(vars["messageID"])
	if err != nil {
		sendResponse(w, http.StatusNotFound, "could not find message ID", true)
		return
	}

	responseMessage := *trackedMessage
	responseMessage.AuthToken = "[Redacted]"

	sendResponse(w, http.StatusOK, responseMessage, false)
	return
}

func (app *app) createMessageHandler(w http.ResponseWriter, req *http.Request) {

	newMessage := struct {
		CallbackURL string   `json:"callback_url"` //URL to send event stream of emoji usage
		ValidEmojis []string `json:"valid_emojis"` //List of emojis to alert on
		AuthToken   string   `json:"auth_token"`   //Auth token given by app to auth on callback
		Expire      int      `json:"expire"`       //Length of time messages can be tracked. Limited to 24h
	}{}

	err := httputil.ParseJSON(req.Body, &newMessage)
	if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusBadRequest, errJSONParseFailure.Error(), true)
		return
	}
	req.Body.Close()

	//Validate user supplied parameters
	err = validation.Errors{
		"callback_url": validation.Validate(newMessage.CallbackURL, is.URL),
		"valid_emojis": validation.Validate(newMessage.ValidEmojis, validation.Required),
		"auth_token":   validation.Validate(newMessage.AuthToken, validation.Required),
	}.Filter()
	if err != nil {
		sendResponse(w, http.StatusBadRequest, err.Error(), true)
		return
	}

	createdMessage := app.createMessage(newMessage.CallbackURL, newMessage.AuthToken, newMessage.ValidEmojis)

	response := struct {
		MessageID string `json:"message_id"`
	}{createdMessage.ID}

	sendResponse(w, http.StatusCreated, response, false)
	return
}

func (app *app) deleteMessageHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	trackedMessage, err := app.getMessage(vars["messageID"])
	if err != nil {
		sendResponse(w, http.StatusNotFound, "could not find message ID", true)
		return
	}

	authInfo := struct {
		AuthToken string `json:"auth_token"`
	}{}

	err = httputil.ParseJSON(req.Body, &authInfo)
	if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusBadRequest, errJSONParseFailure.Error(), true)
		return
	}
	req.Body.Close()

	if authInfo.AuthToken != trackedMessage.AuthToken {
		err := fmt.Errorf("incorrect auth token for messageID: %s; delete failed", trackedMessage.ID)
		log.Println(err)
		sendResponse(w, http.StatusUnauthorized, err.Error(), true)
		return
	}

	err = app.deleteMessage(vars["messageID"])
	if err != nil {
		if err == errMessageNotFound {
			sendResponse(w, http.StatusNotFound, errMessageNotFound, true)
			return
		}

		sendResponse(w, http.StatusInternalServerError, "could not delete message", true)
		return
	}

	sendResponse(w, http.StatusNoContent, "", false)
}
