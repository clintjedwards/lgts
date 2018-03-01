package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (app *app) getMessageHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	message, err := app.getMessage(vars["name"], vars["id"])
	if err == errMessageNotFound {
		sendResponse(w, http.StatusNotFound, fmt.Sprintf("%s: %s", errMessageNotFound.Error(), vars["id"]), true)
		return
	} else if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not retrieve message", true)
		return
	}

	sendResponse(w, http.StatusOK, message, false)
	return
}

func (app *app) getMessagesHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	messages, err := app.getMessages(vars["name"])
	if err != nil {
		if err == errServiceNotFound {
			sendResponse(w, http.StatusNotFound, fmt.Sprintf("%s: %v", errServiceNotFound, vars["name"]), true)
			return
		}
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not retrieve messages", true)
		return
	}

	sendResponse(w, http.StatusOK, messages, false)
	return
}

func (app *app) createMessageHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	messageID, err := app.createMessage(vars["name"])
	if err == errServiceNotFound {
		sendResponse(w, http.StatusNotFound, errServiceNotFound, true)
		return
	} else if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not create message", true)
		return
	}

	response := struct {
		MessageID string `json:"message_id"`
	}{messageID}

	sendResponse(w, http.StatusOK, response, false)
	return
}

func (app *app) deleteMessageHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	err := app.deleteMessage(vars["name"], vars["id"])
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "could not delete message", true)
		return
	}

	sendResponse(w, http.StatusNoContent, "", false)
}
