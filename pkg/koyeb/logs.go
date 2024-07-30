package koyeb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
)

type LogsAPIClient struct {
	url    *url.URL
	header http.Header
}

func NewLogsAPIClient(apiUrl string, token string) (*LogsAPIClient, error) {
	endpoint, err := url.JoinPath(apiUrl, "/v1/streams/logs/tail")
	if err != nil {
		return nil, err
	}
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	switch url.Scheme {
	case "https":
		url.Scheme = "wss"
	case "http":
		url.Scheme = "ws"
	default:
		return nil, fmt.Errorf("unsupported schema: %s", url.Scheme)
	}
	return &LogsAPIClient{
		url: url,
		header: http.Header{
			"Sec-Websocket-Protocol": []string{fmt.Sprintf("Bearer, %s", token)},
		},
	}, nil
}

type WatchLogsQuery struct {
	client       *LogsAPIClient
	logType      string
	serviceId    string
	deploymentId string
	instanceId   string
	since        time.Time
	full         bool // Whether to display full IDs
}

func (client *LogsAPIClient) NewWatchLogsQuery(
	logType string, serviceId string, deploymentId string, instanceId string, since time.Time, full bool,
) (*WatchLogsQuery, error) {
	query := &WatchLogsQuery{
		client:       client,
		serviceId:    serviceId,
		deploymentId: deploymentId,
		instanceId:   instanceId,
		since:        since,
		full:         full,
	}
	switch logType {
	case "build", "runtime", "":
		query.logType = logType
	default:
		return nil, &errors.CLIError{
			What: "Error while fetching the logs",
			Why:  "the log type you provided is invalid",
			Additional: []string{
				fmt.Sprintf("The log type should be either `build` or `runtime`, not `%s`", logType),
			},
			Orig:     nil,
			Solution: "Fix the log type and try again",
		}
	}
	return query, nil
}

// LogLine represents a line returned by /v1/streams/logs/tail
type LogLine struct {
	Result LogLineResult `json:"result"`
}

type LogLineResult struct {
	CreatedAt string              `json:"created_at"`
	Msg       string              `json:"msg"`
	Labels    LogLineResultLabels `json:"labels"`
}

type LogLineResultLabels struct {
	Type           string `json:"type"`
	Stream         string `json:"stream"`
	OrganizationID string `json:"organization_id"`
	AppID          string `json:"app_id"`
	ServiceID      string `json:"service_id"`
	InstanceID     string `json:"instance_id"`
}

// WatchLogsEntry is an entry returned by WatchLogsQuery.Execute()
type WatchLogsEntry struct {
	Date   time.Time           `json:"date"`
	Stream string              `json:"stream"`
	Msg    string              `json:"msg"`
	Err    error               `json:"error"`
	Labels LogLineResultLabels `json:"labels"`
}

// ParseTime parses a time string contained in the field result.created_at of
// the endpoint /v1/streams/logs/tail. In case of error, a zero time is returned.
func (query *WatchLogsQuery) ParseTime(date string) time.Time {
	layout := "2006-01-02T15:04:05.999999999Z"
	parsed, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func (query *WatchLogsQuery) reconnect(isFirstconnection bool) (*WebsocketPingConnection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(query.client.url.String(), query.client.header)
	if err != nil {
		if isFirstconnection {
			return nil, &errors.CLIError{
				What: "Error while fetching the logs",
				Why:  "unable to create the websocket connection",
				Additional: []string{
					"It usually happens because the API URL in your configuration is incorrect",
				},
				Orig:     err,
				Solution: "Fix the error and try again",
			}
		}
		return nil, &errors.CLIError{
			What: "Error while fetching the logs",
			Why:  "we failed to reconnect to the websocket connection",
			Additional: []string{
				"The websocket to the logs API was closed and we couldn't reconnect.",
				"If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
			},
			Orig:     err,
			Solution: "Try again in a few seconds",
		}
	}
	ret := NewWebsocketPingConnection(conn)
	return &ret, nil
}

func (query *WatchLogsQuery) Execute() (chan WatchLogsEntry, error) {
	queryParams := url.Values{}
	if query.logType != "" {
		queryParams.Add("type", query.logType)
	}
	if query.serviceId != "" {
		queryParams.Add("service_id", query.serviceId)
	}
	if query.deploymentId != "" {
		queryParams.Add("deployment_id", query.deploymentId)
	}
	if query.instanceId != "" {
		queryParams.Add("instance_id", query.instanceId)
	}
	if !query.since.IsZero() {
		queryParams.Add("start", query.since.Format(time.RFC3339))
	}
	query.client.url.RawQuery = queryParams.Encode()

	conn, err := query.reconnect(true)
	if err != nil {
		return nil, err
	}

	logs := make(chan WatchLogsEntry)

	go func() {
		var lastLogReceived *LogLine

		for {
			msg := LogLine{}
			err := conn.Conn.ReadJSON(&msg)

			if err != nil {
				// Stop sending ping messages to the websocket connection
				conn.Stop()

				// Normal closure, close the channel and return
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					close(logs)
					return
				}

				// Abnormal closure, try to reconnect. Add a delay to avoid flooding the API with reconnections
				delay := 10 * time.Second
				log.Debugf("Error while fetching the logs: %v, reconnecting in %s...", err, delay)
				time.Sleep(delay)

				// Update the querystring to set the ?start parameter to the
				// date of the last log received, to avoid receiving the same
				// logs again
				if lastLogReceived != nil && lastLogReceived.Result.CreatedAt != "" {
					queryParams.Del("start")
					queryParams.Add("start", lastLogReceived.Result.CreatedAt)
					query.client.url.RawQuery = queryParams.Encode()
				}

				conn, err = query.reconnect(false)
				if err == nil {
					log.Debugf("Reconnection successful")
					continue
				}
				// Reconnection failed, return the error to the caller and close the channel
				log.Debugf("Unable to reconnect")
				logs <- WatchLogsEntry{Err: err}
				close(logs)
				return
			}

			// Sometimes, for example when passing a future date in --since, the
			// first log message is empty.
			var emptyLogLine LogLine
			if msg == emptyLogLine {
				continue
			}

			// If the last log received is the same as the current one, ignore
			// it. This can happens when there is a connection error: we
			// reconnect and set the ?start parameter to the last log received,
			// which is then sent again.
			if lastLogReceived != nil && msg == *lastLogReceived {
				continue
			}

			lastLogReceived = &msg
			logs <- WatchLogsEntry{
				Stream: msg.Result.Labels.Stream,
				Msg:    msg.Result.Msg,
				Date:   query.ParseTime(msg.Result.CreatedAt),
				Labels: msg.Result.Labels,
			}
		}
	}()
	return logs, nil
}

// WebsocketPingConnection is a wrapper around a websocket connection that sends
// a ping message every few seconds. The Stop() method should be called to stop
// sending ping messages.
type WebsocketPingConnection struct {
	Conn     *websocket.Conn
	stopChan chan (struct{})
}

func NewWebsocketPingConnection(conn *websocket.Conn) WebsocketPingConnection {
	ret := WebsocketPingConnection{Conn: conn, stopChan: make(chan struct{})}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case tick := <-ticker.C:
				err := conn.WriteMessage(websocket.PingMessage, []byte(tick.String()))
				if err != nil {
					log.Debugf("Unable to send a ping message to the websocket connection: %v", err)
					return
				}
			case <-ret.stopChan:
				close(ret.stopChan)
				return
			}
		}
	}()
	return ret
}

// Stop sendings ping messages to the websocket connection.
func (conn *WebsocketPingConnection) Stop() {
	conn.stopChan <- struct{}{}
}

// PrintAll prints all the logs returned by WatchLogsQuery.Execute(). It returns
// an error if the query failed, or if there was an error while printing the
// logs.
func (query *WatchLogsQuery) PrintAll() error {
	logs, err := query.Execute()
	if err != nil {
		return err
	}
	for log := range logs {
		if log.Err != nil {
			return log.Err
		}
		layout := "2006-01-02 15:04:05"
		date := log.Date.Format(layout)
		zone, _ := log.Date.Zone()

		switch outputFormat {
		case "json", "yaml":
			data, err := json.Marshal(log)
			if err != nil {
				return err
			}
			fmt.Printf("%s\n", data)
		default:
			fmt.Printf("[%s %s] %s %6s - %s\n", date, zone, renderer.FormatID(log.Labels.InstanceID, query.full), log.Stream, log.Msg)
		}
	}
	return nil
}
