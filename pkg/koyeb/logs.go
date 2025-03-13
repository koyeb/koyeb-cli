package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
)

const (
	tailPath  = "/v1/streams/logs/tail"
	queryPath = "/v1/streams/logs/query"
)

type LogsAPIClient struct {
	client *koyeb.APIClient
	url    *url.URL
	token  string
}

func NewLogsAPIClient(client *koyeb.APIClient, apiUrl string, token string) (*LogsAPIClient, error) {
	url, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	return &LogsAPIClient{
		client: client,
		url:    url,
		token:  token,
	}, nil
}

type WatchLogsQuery struct {
	url          *url.URL
	header       http.Header
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
		serviceId:    serviceId,
		deploymentId: deploymentId,
		instanceId:   instanceId,
		since:        since,
		full:         full,
	}

	endpoint, err := url.JoinPath(client.url.String(), tailPath)
	if err != nil {
		return nil, err
	}
	query.url, err = url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	switch query.url.Scheme {
	case "https":
		query.url.Scheme = "wss"
	case "http":
		query.url.Scheme = "ws"
	default:
		return nil, fmt.Errorf("unsupported schema: %s", query.url.Scheme)
	}

	query.header = http.Header{
		"Sec-Websocket-Protocol": []string{fmt.Sprintf("Bearer, %s", token)},
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

func (client *LogsAPIClient) ExecuteQueryLogsQuery(ctx context.Context,
	logType string, serviceId string, deploymentId string, instanceId string, start time.Time, end time.Time, regex string, text string, order string, full bool,
) ([]koyeb.LogEntry, error) {
	switch logType {
	case "build", "runtime", "":
		break
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

	hasMore := true
	logs := make([]koyeb.LogEntry, 0)

	req := client.client.LogsApi.QueryLogs(ctx).
		Type_(logType).
		ServiceId(serviceId).
		DeploymentId(deploymentId).
		InstanceId(instanceId).
		Regex(regex).
		Text(text).
		Limit(fmt.Sprintf("%d", 100)).
		Order(order)

	for hasMore {
		req = req.Start(start).End(end)

		resp, _, err := req.Execute()
		if err != nil {
			return nil, &errors.CLIError{
				What: "Error while fetching logs",
				Why:  "could not fetch query results",
				Orig: err,
			}
		}
		logs = append(logs, resp.Data...)

		if resp.Pagination.HasMore != nil {
			hasMore = *resp.Pagination.HasMore
			if hasMore {
				start = *resp.Pagination.NextStart
				end = *resp.Pagination.NextEnd
			}
		}
	}

	return logs, nil
}

// LogLine represents a line returned by /v1/streams/logs/tail or /v1/streams/logs/query
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
	conn, _, err := websocket.DefaultDialer.Dial(query.url.String(), query.header)
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
	query.url.RawQuery = queryParams.Encode()

	conn, err := query.reconnect(true)
	if err != nil {
		return nil, err
	}

	logs := make(chan WatchLogsEntry)

	go func() {
		var lastLogReceived *LogLine

		logsTimeout := 6 * time.Hour
		timer := time.NewTimer(logsTimeout)

		for {
			readCh := make(chan LogLine)
			errCh := make(chan error)

			go func() {
				defer close(readCh)
				defer close(errCh)

				var msg LogLine
				err := conn.Conn.ReadJSON(&msg)
				if err != nil {
					errCh <- err
				} else {
					readCh <- msg
				}
			}()

			select {
			case <-timer.C:
				// Stop sending ping messages to the websocket connection
				conn.Stop()

				newErr := &errors.CLIError{
					Icon: "⏰",
					What: "Disconnected from the logs API",
					Why:  fmt.Sprintf("forced disconnection after %s", logsTimeout),
					Additional: []string{
						fmt.Sprintf("To avoid keeping the connection to the logs API open indefinitely, the CLI disconnects after %s.", logsTimeout),
						"This timeout value is hardcoded in the CLI and cannot be changed.",
						"If you need to make the timeout configurable, please create an issue on GitHub:",
						"https://github.com/koyeb/koyeb-cli/issues/new",
					},
					Orig:     nil,
					Solution: "Run the command again to reconnect",
				}
				log.Errorf("%s", newErr)
				close(logs)
			case msg := <-readCh:
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
			case err := <-errCh:
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
					query.url.RawQuery = queryParams.Encode()
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
	for logl := range logs {
		if logl.Err != nil {
			return logl.Err
		}
		PrintLogLine(logl, query.full, logl.Date, logl.Msg, logl.Stream, logl.Labels.InstanceID)
	}
	return nil
}

func PrintLogLine(logl any, full bool, ts time.Time, msg string, stream string, instanceID string) error {
	layout := "2006-01-02 15:04:05"
	date := ts.Format(layout)
	zone, _ := ts.Zone()

	switch outputFormat {
	case "json", "yaml":
		data, err := json.Marshal(logl)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", data)
	default:
		fmt.Printf("[%s %s] %s %6s - %s\n", date, zone, renderer.FormatID(instanceID, full), stream, msg)
	}

	return nil
}
