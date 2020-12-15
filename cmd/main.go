package main

import (
	"context"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
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

func processEvent(ctx context.Context, event cloudevents.Event, ret chan struct{}) error {
	time.Sleep(time.Second * 10)
	ret <- struct{}{}
	return nil
}

func Handler(ctx context.Context, event cloudevents.Event) (error) {
	eventStr := event.String()
	log.Printf("receive cloudevents.Event: ", strings.Replace(eventStr, "\n", " ", -1))
	var task EventData
	if err := event.DataAs(&task); err != nil {
		return cloudevents.NewHTTPResult(400, "this is a mod 7 server error message")
	}

	uid, _ := uuid.NewUUID()
	lockerVal := fmt.Sprintf("%v/%s", event.ID(), uid.String())
	processLocker.Lock()
	if processing != "" {
		processLocker.Unlock()
		log.Printf("service is running another task now, so current task[%d] failed", task.ID)
		return cloudevents.NewHTTPResult(http.StatusPreconditionFailed, "service is running another task now")
	}
	processing = lockerVal
	processLocker.Unlock()

	retChan := make(chan struct{})
	go processEvent(ctx, event, retChan)
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
	return nil
}

func main() {

	adminServer := buildAdminServer()

	// Don't forward ErrServerClosed as that indicates we're already shutting down.
	log.Printf("will listen on :8080\n")
	if err := adminServer.ListenAndServe(); err != nil {
		log.Fatalf("unable to start http server, %s", err)
	}

}
func buildAdminServer() *http.Server {
	ctx := context.Background()
	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	h, err := cloudevents.NewHTTPReceiveHandler(ctx, p, Handler)
	if err != nil {
		log.Fatalf("failed to create handler: %s", err.Error())
	}
	return &http.Server{
		Addr:    ":8080",
		Handler: h,
	}
}