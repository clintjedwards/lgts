package httputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

//DefaultHeaders is a wrapper function setting the reponse headers
func DefaultHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

//ParseJSON parses the given json request into interface
func ParseJSON(rc io.Reader, object interface{}) error {
	decoder := json.NewDecoder(rc)
	err := decoder.Decode(object)
	if err != nil {
		log.Println(err)
		return errors.New("could not parse json")
	}
	return nil
}

//SendHTTPGETRequest allows easy send of http GET requests. Automatically takes care of retrying
func SendHTTPGETRequest(url string, headers map[string]string, debug bool) (*http.Response, error) {
	client := retryablehttp.NewClient()
	if !debug {
		client.Logger.SetOutput(ioutil.Discard)
	}
	request, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if len(headers) != 0 {
		for headerValue := range headers {
			request.Header.Add(headerValue, headers[headerValue])
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

//SendHTTPPOSTRequest allows easy send of http POST requests. Automatically takes care of retrying
func SendHTTPPOSTRequest(url string, headers map[string]string, messageBody []byte, debug bool) (*http.Response, error) {

	client := retryablehttp.NewClient()
	if !debug {
		client.Logger.SetOutput(ioutil.Discard)
	}

	request, err := retryablehttp.NewRequest("POST", url, bytes.NewReader(messageBody))
	if err != nil {
		return nil, err
	}

	if len(headers) != 0 {
		for headerValue := range headers {
			request.Header.Add(headerValue, headers[headerValue])
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}
