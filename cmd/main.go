package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	SysError = "Internal Service Error"
)

type EventData struct {
	TemplateName string `json:"template_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ID           int64  `json:"id"`
}

var (
	processLocker = sync.Mutex{}
	processing    = ""
)

func processEvent(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse, ret chan struct{}) error {
	time.Sleep(time.Second * 10)
	ret <- struct{}{}
	return nil
}

func Handler(ctx context.Context, event cloudevents.Event, response *cloudevents.EventResponse) error {
	eventStr := event.String()
	log.Printf("receive cloudevents.Event: ", strings.Replace(eventStr, "\n", " ", -1))
	response.RespondWith(200, &event)
	return nil
}

func main() {
	log.Print("Hello world sample started.")
	c, err := cloudevents.NewDefaultClient()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	log.Fatal(c.StartReceiver(context.Background(), Handler))
}
