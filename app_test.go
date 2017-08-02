package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAuthorizedUser(t *testing.T) {
	app := &app{"testapp", "https://localhost", []string{"barack.obama@gmail.com"}, "somerandomstringofcharacters"}

	t.Run("is authorized", func(t *testing.T) {
		if !app.isAuthorizedUser("barack.obama@gmail.com") {
			t.Fatalf("return for isAuthorizedIUser false; should be true")
		}
	})

	t.Run("not authorized", func(t *testing.T) {
		if app.isAuthorizedUser("lewis.hamilton@gmail.com") {
			t.Fatalf("return for isAuthorizedIUser true; should be false")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if app.isAuthorizedUser("") {
			t.Fatalf("return for isAuthorizedIUser true; should be false")
		}
	})

}

func TestSendMessageApproval(t *testing.T) {

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	}))
	defer testServer.Close()

	var callbackInfo = make(map[string]interface{})

	callbackInfo["app_name"] = "testapp"
	callbackInfo["message_id"] = "randomstringofchars"
	callbackInfo["test_info"] = "nothingofsubstance"

	app := &app{"testapp", testServer.URL, []string{"barack.obama@gmail.com"}, "somerandomstringofcharacters"}

	t.Run("approval successfully sent", func(t *testing.T) {
		err := app.sendMessageApproval(callbackInfo, "barack.obama@gmail.com", true)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("correct json params sent", func(t *testing.T) {

	})

}
