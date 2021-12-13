package koyeb

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type LogMessage struct {
	Result LogMessageResult `json:"result"`
}

func (l LogMessage) String() string {
	return l.Result.Msg
}

type LogMessageResult struct {
	Msg string `json:"msg"`
}

type WatchLogQuery struct {
	LogType      *string
	ServiceID    *string
	InstanceID   *string
	DeploymentID *string
}

func WatchLog(q *WatchLogQuery, done chan struct{}) error {
	path := "/v1/streams/logs/tail?"
	if q.LogType != nil {
		path = fmt.Sprintf("%s&type=%s", path, *q.LogType)
	}
	if q.DeploymentID != nil {
		path = fmt.Sprintf("%s&deployment_id=%s", path, *q.DeploymentID)
	}
	if q.ServiceID != nil {
		path = fmt.Sprintf("%s&service_id=%s", path, *q.ServiceID)
	}
	if q.InstanceID != nil {
		path = fmt.Sprintf("%s&instance_id=%s", path, *q.InstanceID)
	}

	dest, err := url.Parse(fmt.Sprint(apiurl, path))
	if err != nil {
		return fmt.Errorf("cannot parse url for websocket: %w", err)
	}
	switch dest.Scheme {
	case "https":
		dest.Scheme = "wss"
	case "http":
		dest.Scheme = "ws"
	default:
		return fmt.Errorf("unsupported schema: %s", dest.Scheme)
	}

	h := http.Header{"Sec-Websocket-Protocol": []string{fmt.Sprintf("Bearer, %s", token)}}
	c, _, err := websocket.DefaultDialer.Dial(dest.String(), h)
	if err != nil {
		return fmt.Errorf("cannot create websocket: %w", err)
	}
	defer c.Close()

	readDone := make(chan struct{})

	go func() {
		defer close(done)
		for {
			msg := LogMessage{}
			err := c.ReadJSON(&msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				return
			}
			fmt.Println(msg.String())
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-readDone:
			return nil
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return err
			}
		}
	}
}
