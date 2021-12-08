package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *ServiceHandler) Log(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	serviceDetail, _, err := client.ServicesApi.GetService(ctx, ResolveServiceShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	serviceID := serviceDetail.Service.GetId()
	instanceID, _ := cmd.Flags().GetString("instance")

	done := make(chan struct{})
	query := &watchLogQuery{serviceID: koyeb.PtrString(serviceID)}
	if instanceID != "" {
		query.instanceID = koyeb.PtrString(instanceID)
	}
	return watchLog(query, done)
}

type LogMessage struct {
	Result LogMessageResult `json:"result"`
}

func (l LogMessage) String() string {
	return l.Result.Msg
}

type LogMessageResult struct {
	Msg string `json:"msg"`
}

type watchLogQuery struct {
	serviceID    *string
	instanceID   *string
	deploymentID *string
	logType      *string
}

func watchLog(q *watchLogQuery, done chan struct{}) error {
	path := "/v1/streams/logs/tail?"
	if q.logType != nil {
		path = fmt.Sprintf("%s&type=%s", path, *q.logType)
	}
	if q.deploymentID != nil {
		path = fmt.Sprintf("%s&deployment_id=%s", path, *q.deploymentID)
	}
	if q.serviceID != nil {
		path = fmt.Sprintf("%s&service_id=%s", path, *q.serviceID)
	}
	if q.instanceID != nil {
		path = fmt.Sprintf("%s&instance_id=%s", path, *q.instanceID)
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
				log.Println("error:", err)
				return
			}
			log.Printf("%s", msg)
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
