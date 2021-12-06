package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
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
	logType, _ := cmd.Flags().GetString("type")

	done := make(chan struct{})

	return watchLog(serviceID, instanceID, logType, done)
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

func watchLog(serviceID string, instanceID string, logType string, done chan struct{}) error {
	path := fmt.Sprintf("/v1/streams/logs/tail?service_id=%s", serviceID)
	if logType == "" {
		path = fmt.Sprintf("%s&type=%s", path, "runtime")
	}
	if instanceID != "" {
		path = fmt.Sprintf("%s&instance_id=%s", path, instanceID)
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
