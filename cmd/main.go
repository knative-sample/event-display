package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/google/uuid"
	"log"
	"net/http"
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
	var task EventData
	rawData, ok := event.Data.([]byte)
	if !ok {
		log.Print("get event data error, raw data: ", event.Data)
		response.RespondWith(400, &event)
		return errors.New(SysError)
	}
	if err := json.Unmarshal(rawData, &task); err != nil {
		log.Print("unmarshal event error: ", err)
		response.RespondWith(400, &event)
		return errors.New(SysError)
	}
	uid, _ := uuid.NewUUID()
	lockerVal := fmt.Sprintf("%v/%s", event.ID(), uid.String())
	processLocker.Lock()
	if processing != "" {
		processLocker.Unlock()
		log.Printf("service is running another task now, so current task[%d] failed", task.ID)
		response.RespondWith(http.StatusPreconditionFailed, &event)
	}
	processing = lockerVal
	processLocker.Unlock()

	retChan := make(chan struct{})
	go processEvent(ctx, event, response, retChan)
	timeout := time.After(time.Second * 15)
	select {
	case <-timeout:
		// TODO
	case <-ctx.Done():
		// TODO
	case <-retChan:
		// TODO
	}
	processLocker.Lock()
	if processing == lockerVal {
		processing = ""
	}
	processLocker.Unlock()
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
