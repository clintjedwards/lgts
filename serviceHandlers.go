package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/clintjedwards/lgts/helpers/httputil"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/gorilla/mux"
)

func (app *app) getServiceHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	service, err := app.getService(vars["name"])
	if err == errServiceNotFound {
		sendResponse(w, http.StatusNotFound, fmt.Sprintf("%s: %s", errServiceNotFound.Error(), vars["name"]), true)
		return
	} else if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not retrieve service", true)
		return
	}

	sendResponse(w, http.StatusOK, service, false)
	return
}

func (app *app) getServicesHandler(w http.ResponseWriter, req *http.Request) {
	services, err := app.getServices()
	if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not retrieve services", true)
		return
	}

	sendResponse(w, http.StatusOK, services, false)
	return
}

func (app *app) createServiceHandler(w http.ResponseWriter, req *http.Request) {

	newService := newService()
	err := httputil.ParseJSON(req.Body, &newService)
	if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusBadRequest, "could not decode json body", true)
		return
	}

	err = validation.Errors{
		"name":         validation.Validate(newService.Name, validation.Required),
		"callback_url": validation.Validate(newService.CallbackURL, validation.Required, is.URL),
	}.Filter()
	if err != nil {
		sendResponse(w, http.StatusBadRequest, err.Error(), true)
		return
	}

	err = app.createService(newService)
	if err == errServiceExists {
		sendResponse(w, http.StatusConflict, errServiceExists.Error(), true)
		return
	} else if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not create service", true)
		return
	}

	response := struct {
		Name               string   `json:"name" gorm:"unique_index"`
		CallbackURL        string   `json:"callback_url"`
		ValidEmojis        []string `json:"valid_emojis"`
		AuthorizedSlackers []string `json:"authorized_slackers"`
		Token              string   `json:"auth_token"`
	}{newService.Name, newService.CallbackURL, newService.ValidEmojis, newService.AuthorizedSlackers, newService.AuthToken}

	sendResponse(w, http.StatusOK, response, false)
	return
}

func (app *app) updateServiceHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	updatedService, err := app.getService(vars["name"])
	if err == errServiceNotFound {
		sendResponse(w, http.StatusNotFound, fmt.Sprintf("%s: %s", errServiceNotFound.Error(), vars["name"]), true)
		return
	} else if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not retrieve service", true)
		return
	}

	//Validate user supplied parameters
	err = validation.Errors{
		"callback_url": validation.Validate(updatedService.CallbackURL, is.URL),
	}.Filter()
	if err != nil {
		sendResponse(w, http.StatusBadRequest, err.Error(), true)
		return
	}

	err = app.updateService(updatedService)
	if err != nil {
		log.Println(err)
		sendResponse(w, http.StatusInternalServerError, "could not update service", true)
		return
	}

	sendResponse(w, http.StatusNoContent, "", false)
	return
}

func (app *app) deleteServiceHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	err := app.deleteService(vars["name"])
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, "could not delete service", true)
		return
	}

	sendResponse(w, http.StatusNoContent, "", false)
}
