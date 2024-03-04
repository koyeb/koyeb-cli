package koyeb

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
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
	conn         *websocket.Conn
	ticker       *time.Ticker
}

func (client *LogsAPIClient) NewWatchLogsQuery(
	logType string, serviceId string, deploymentId string, instanceId string,
) (*WatchLogsQuery, error) {
	query := &WatchLogsQuery{
		client:       client,
		serviceId:    serviceId,
		deploymentId: deploymentId,
		instanceId:   instanceId,
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
	Date   time.Time
	Stream string
	Msg    string
	Err    error
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
	query.client.url.RawQuery = queryParams.Encode()

	conn, _, err := websocket.DefaultDialer.Dial(query.client.url.String(), query.client.header)
	if err != nil {
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
	query.conn = conn

	// Read logs from the websocket connection
	logs := make(chan WatchLogsEntry)
	go func() {
		for {
			msg := LogLine{}
			err := conn.ReadJSON(&msg)
			if err != nil {
				logs <- WatchLogsEntry{Err: &errors.CLIError{
					What: "Error while fetching the logs",
					Why:  "unable to read the logs from the websocket connection",
					Additional: []string{
						"Unfortunately, we couldn't read the logs from the websocket connection",
						"If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
					},
					Orig:     err,
					Solution: "Try again in a few seconds",
				}}
			} else {
				logs <- WatchLogsEntry{
					Stream: msg.Result.Labels.Stream,
					Msg:    msg.Result.Msg,
					Date:   query.ParseTime(msg.Result.CreatedAt),
				}
			}
		}
	}()
	// Consume the logs channel, forward them to the caller. Also send a ping every 10 seconds to keep the connection alive.
	ret := make(chan WatchLogsEntry)
	query.ticker = time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case line := <-logs:
				ret <- line
				if line.Err != nil {
					close(ret)
					return
				}
			case tick := <-query.ticker.C:
				err := conn.WriteMessage(websocket.PingMessage, []byte(tick.String()))
				if err != nil {
					ret <- WatchLogsEntry{Err: err}
					close(ret)
					return
				}
			}
		}
	}()
	return ret, nil
}

func (query *WatchLogsQuery) Close() {
	if query.conn != nil {
		query.conn.Close()
	}
	if query.ticker != nil {
		query.ticker.Stop()
	}
}

// PrintAll prints all the logs returned by WatchLogsQuery.Execute(). It returns
// an error if the query failed, or if there was an error while printing the
// logs.
func (query *WatchLogsQuery) PrintAll() error {
	logs, err := query.Execute()
	if err != nil {
		return err
	}
	defer query.Close()
	for log := range logs {
		if log.Err != nil {
			return log.Err
		}
		layout := "2006-01-02 15:04:05"
		date := log.Date.Format(layout)
		zone, _ := log.Date.Zone()
		fmt.Printf("[%s %s] %6s - %s\n", date, zone, log.Stream, log.Msg)
	}
	return nil
}
