package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"

	"encoding/json"

	"github.com/pkg/errors"
	"google.golang.org/api/chat/v1"
)

var (
	space string
)

const (
	url = "http://us-central1-priya-wadhwa.cloudfunctions.net/Chatbot"
)

func init() {
	flag.StringVar(&space, "space", "", "pass in the hangouts chat space name you want the container-tools-bot to respond to")
	flag.Parse()
}

func main() {
	if err := pingEndpoint(); err != nil {
		log.Printf("error pinging chatbot endpoint to respond to space %s: %v", space, err)
	}
}

func pingEndpoint() error {
	m := retrieveMessage()
	data, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "unmarshalling message")
	}

	fmt.Println(string(data))

	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return errors.Wrap(err, "generating http request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(req)
	return err
}

func retrieveMessage() chat.Message {
	m := chat.Message{
		Space: &chat.Space{
			Name: fmt.Sprintf("spaces/%s", space),
		},
	}
	return m
}
