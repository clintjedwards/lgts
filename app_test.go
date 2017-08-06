package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewApp(t *testing.T) {
	newapp := newApp()

	t.Run("is not empty", func(t *testing.T) {
		if newapp.Token == "" {
			t.Fatal("token field for struct app empty")
		}
	})

	t.Run("correct length", func(t *testing.T) {
		if len(newapp.Token) != 20 {
			t.Log(newapp.Token)
			t.Fatalf("expected token length 10; got length %d", len(newapp.Token))
		}
	})
}

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

	t.Run("server recieves correct message format", func(t *testing.T) {
		var callbackInfo = make(map[string]interface{})

		callbackInfo["app_name"] = "testapp"
		callbackInfo["message_id"] = "randomstringofchars"
		callbackInfo["test_info"] = "nothingofsubstance"

		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			receivedJSON, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}

			var receivedValues map[string]interface{}
			json.Unmarshal([]byte(receivedJSON), &receivedValues)

			for key := range callbackInfo {
				if callbackInfo[key] != receivedValues[key] {
					t.Fatalf("recieved value %s for key %s different from expected value %s", callbackInfo[key], key, receivedValues[key])
				}
			}

		}))

		defer testServer.Close()

		app := &app{"testapp", testServer.URL, []string{"barack.obama@gmail.com"}, "somerandomstringofcharacters"}

		err := app.sendMessageApproval(callbackInfo, "barack.obama@gmail.com", true)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("returns error if server doesn't respond", func(t *testing.T) {
		var callbackInfo = make(map[string]interface{})
		app := &app{"testapp", "http://localhost", []string{"barack.obama@gmail.com"}, "somerandomstringofcharacters"}
		err := app.sendMessageApproval(callbackInfo, "barack.obama@gmail.com", true)
		if err != nil {
			if !strings.Contains(err.Error(), "connection refused") {
				t.Fatalf("incorrect error response; expected body contains connection refused got %s", err)
			}
		}
	})

}
