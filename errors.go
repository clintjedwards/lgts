package main

import "errors"

var errServiceNotFound = errors.New("service not found")
var errServiceExists = errors.New("service already exists")

var errMessageNotFound = errors.New("message not found")
var errMessageExists = errors.New("message already exists")

var errJSONParseFailure = errors.New("could not parse json")
