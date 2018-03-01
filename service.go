package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/clintjedwards/lgts/helpers/httputil"
)

type service struct {
	ID                 int       `json:"-"`
	Name               string    `json:"name" sql:",unique"`         //Name of the application
	CallbackURL        string    `json:"callback_url" sql:",unique"` //URL of callback endpoint where lgts will post a response
	ValidEmojis        []string  `json:"valid_emojis" pg:",array"`
	AuthorizedSlackers []string  `json:"authorized_slackers" pg:",array"`
	CreatedAt          time.Time `json:"created_at" sql:"default:now()"`
	AuthToken          string    `json:"-"` //Pre-shared randomly generated token that must be included on message
}

func newService() *service {

	rand.Seed(time.Now().UTC().UnixNano())
	token := make([]byte, 10)
	rand.Read(token)

	return &service{
		AuthToken: fmt.Sprintf("%x", token),
	}

}

func (app *app) getService(name string) (*service, error) {
	service := service{}

	err := app.db.Model(&service).Where("name = ?", name).First()
	if service.ID == 0 {
		return nil, errServiceNotFound
	} else if err != nil {
		return nil, err
	}

	return &service, nil
}

func (app *app) getServices() ([]*service, error) {
	services := []*service{}

	err := app.db.Model(&services).Select()
	if err != nil {
		return nil, err
	}

	return services, nil
}

func (app *app) createService(newService *service) error {
	_, err := app.getService(newService.Name)
	if err != nil {
		if err != errServiceNotFound {
			return errServiceExists
		}
	}

	response, err := app.db.Model(newService).OnConflict("DO NOTHING").Insert()
	if err != nil {
		return err
	}

	if response.RowsAffected() == 0 {
		return errServiceExists
	}

	return nil
}

func (app *app) updateService(updatedService *service) error {
	storedService, err := app.getService(updatedService.Name)
	if err != nil {
		return err
	}

	updatedService.Name = storedService.Name
	updatedService.AuthToken = storedService.AuthToken

	err = app.db.Update(updatedService)
	if err != nil {
		return err
	}

	return nil
}

func (app *app) deleteService(name string) error {

	currentService, err := app.getService(name)
	if err != nil {
		return err
	}

	app.messages[currentService.Name] = nil

	err = app.db.Delete(currentService)
	if err != nil {
		return err
	}

	return nil
}

func (service *service) isAuthorizedSlacker(email string) bool {

	for _, approvedEmail := range service.AuthorizedSlackers {
		if approvedEmail == email {
			return true
		}
	}
	return false

}

func (service *service) isValidEmoji(emoji string) bool {

	for _, validEmoji := range service.ValidEmojis {
		if validEmoji == emoji {
			return true
		}
	}
	return false

}

func (service *service) sendCallbackMessage(messageID, slackerEmail, emojiUsed string) error {

	type callbackInfo struct {
		MessageID    string `json:"message_id"`
		Token        string `json:"token"`
		SlackerEmail string `json:"slacker_email"`
		EmojiUsed    string `json:"emoji_used"`
	}

	info := &callbackInfo{
		MessageID:    messageID,
		Token:        service.AuthToken,
		SlackerEmail: slackerEmail,
		EmojiUsed:    emojiUsed,
	}

	jsonString, _ := json.Marshal(&info)

	headers := map[string]string{
		"content-type": "application/json",
	}

	response, err := httputil.SendHTTPPOSTRequest(service.CallbackURL, headers, jsonString)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNoContent {
		err := fmt.Errorf("callback URL did not return correct status code; recieved %d", response.StatusCode)
		return err
	}

	return nil
}
