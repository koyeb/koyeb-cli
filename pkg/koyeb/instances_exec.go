package koyeb

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"fmt"
	"syscall"
)

// ApiExecCommandRequest is == koyeb.ApiExecCommandRequest but with public
// fields.
// Although it is unconvenient to use a custom-defined struct, we have no other
// choice. In fact, we want to use websockets; hence, we bypass grpc-gateway
// http engine.
// Thus, although the payload we're sending is == to koyeb.ApiExecCommandRequest,
// there is no public method to serialize in JSON the struct; it is supposed to
// be done internally
type ApiExecCommandRequest struct {
	Id   *string  `json:"id,omitempty"`
	Body *ReqBody `json:"body,omitempty"`
}

type ReqBody struct {
	Command *[]string `json:"command,omitempty"`
	TTYSize *TTYSize  `json:"ttysize,omitempty"`
	Stdin   *IO       `json:"stdin,omitempty"`
}

type TTYSize struct {
	Height *int32 `json:"height,omitempty"`
	Width  *int32 `json:"width,omitempty"`
}

type IO struct {
	Data *[]byte `json:"data,omitempty"`
}

func (h *InstanceHandler) exec(instanceId string, cmd, input []string) error {
	path := fmt.Sprintf("/v1/instances/exec")
	c, err := dial(apiurl, path)
	if err != nil {
		return errors.Wrapf(err, "could not dial %s", path)
	}
	go closeOn(c, syscall.SIGINT, syscall.SIGTERM)
	// Safeguard: closeOn() should handle this but in case something panics,
	// let's keep this
	defer c.Close()

	// Write input
	for _, txt := range input {
		data := bytes.NewBufferString(txt).Bytes()
		r := &ApiExecCommandRequest{
			Id: &instanceId,
			Body: &ReqBody{
				Command: &cmd,
				Stdin: &IO{
					Data: &data,
				},
			},
		}
		c.WriteJSON(r)
	}

	// Reader prints everything it receives to stdout and returns when:
	// * an unexpected error arises
	// * it receives a close control message from the server
	err = read(c)
	if err != nil {
		return err
	}

	return nil
}

func dial(address, path string) (*websocket.Conn, error) {
	u, err := url.Parse(address)
	if err != nil {
		er(err)
	}
	u.Path = path
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	headers := http.Header{
		"Sec-Websocket-Protocol": []string{
			fmt.Sprintf("Bearer, %s", token),
		},
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	return c, err
}

func closeOn(c *websocket.Conn, s ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, s...)
	<-sigs
	// gorilla.websocket does not support implicit graceful shutdowns.
	// We need to manually send a control message to the server. Once
	// done, the client websocket (our) will go to closing mode. When
	// the server sends back a close control message, read() should
	// receive an error with the code  websocket.CloseNormalClosure.
	// See https://github.com/gorilla/websocket/issues/448
	err := sendCloseMsg(c)
	if err != nil {
		log.Warnf("exec: could not cleanly close websocket (%s)", err)
	}
}

func sendCloseMsg(c *websocket.Conn) error {
	code := websocket.CloseNormalClosure
	msg := websocket.FormatCloseMessage(code, "close from client")
	deadline := time.Now().Add(time.Second * 300)

	err := c.WriteControl(websocket.CloseMessage, msg, deadline)
	if err != nil && err != websocket.ErrCloseSent {
		// Close message could not be sent. Let's close without the handshake.
		return c.Close()
	}
	return nil
}

func read(c *websocket.Conn) error {
	for {
		t, r, err := c.NextReader()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				// Two cases:
				// * we closed the connection by sending the server a close control msg and
				// this is its ack answer.
				// * the server wishes to close the connection and is proactively sending a
				// close control msg. The default client close handler will take care of
				// sending back a close ack answer.
				// In both cases, that means that the connection ended normally, we can safely
				// exit
				return nil
			}
			return errors.Wrap(err, "read failed")
		}
		if t != websocket.TextMessage {
			return fmt.Errorf("read failed: expected receiving TextMessages (%d), got %d", websocket.TextMessage, t)
		}
		err = readFrom(r)
		if err != nil {
			return errors.Wrap(err, "read failed")
		}
	}
}

func readFrom(r io.Reader) error {
	for {
		buff := make([]byte, 8192)
		_, err := r.Read(buff)
		if err == nil {
			fmt.Printf(string(buff) + "\n")
		} else if err == io.EOF {
			return nil
		} else {
			return err
		}
	}
}
