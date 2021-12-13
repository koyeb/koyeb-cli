package koyeb

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/moby/term"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *InstanceHandler) Exec(cmd *cobra.Command, args []string) error {
	returnCode, err := h.exec(cmd, args)
	if err != nil {
		fatalApiError(err)
	}
	if returnCode != 0 {
		os.Exit(returnCode)
	}
	return nil
}

func (h *InstanceHandler) exec(cmd *cobra.Command, args []string) (int, error) {
	// Cobra options ensure we have at least 2 arguments here, but still
	if len(args) < 2 {
		return 0, errors.New("exec needs at least 2 arguments")
	}

	instanceId := h.ResolveInstanceArgs(args[0])
	userCmd := args[1:]

	stdStreams, cleanup, err := GetStdStreams()
	if err != nil {
		return 0, errors.Wrap(err, "could not get standard streams")
	}
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termResizeCh := watchTermSize(ctx, stdStreams)

	e := NewExecutor(stdStreams.Stdin, stdStreams.Stdout, stdStreams.Stderr, userCmd, instanceId, termResizeCh)
	return e.Run(ctx)
}

func watchTermSize(ctx context.Context, s *StdStreams) <-chan *TerminalSize {
	out := make(chan *TerminalSize)
	go func() {
		defer close(out)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				termSize, err := GetTermSize(s.Stdout)
				if err != nil {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- termSize:
				}
			}
		}
	}()
	return out
}

// Disclaimer: parts of this file are either taken or largey inspired from Nomad's CLI
// implementation (command/alloc_exec.go) or API implementation (api/allocations_exec.go)

// ApiExecCommandRequest is == koyeb.ApiExecCommandRequest but with public
// fields.
// Although it is unconvenient to use a custom-defined struct, we have no other
// choice. In fact, we want to use websockets; hence, we bypass grpc-gateway
// http engine.
// Thus, although the payload we're sending is == to koyeb.ApiExecCommandRequest,
// there is no public method to serialize in JSON the struct; it is supposed to
// be done internally
type ApiExecCommandRequest struct {
	Id   *string                       `json:"id,omitempty"`
	Body *koyeb.ExecCommandRequestBody `json:"body,omitempty"`
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

type TerminalSize struct {
	Height int32
	Width  int32
}

type Executor struct {
	stdin        io.Reader
	stderr       io.Writer
	stdout       io.Writer
	termResizeCh <-chan *TerminalSize

	cmd        []string
	instanceId string
}

func NewExecutor(stdin io.Reader, stdout, stderr io.Writer, cmd []string, instanceId string, termResizeCh <-chan *TerminalSize) *Executor {
	return &Executor{
		stdin:        stdin,
		stdout:       stdout,
		stderr:       stderr,
		cmd:          cmd,
		instanceId:   instanceId,
		termResizeCh: termResizeCh,
	}
}

func (e *Executor) Run(ctx context.Context) (int, error) {
	path := fmt.Sprintf("/v1/streams/instances/exec")
	c, err := e.dial(apiurl, path)
	if err != nil {
		return -1, errors.Wrapf(err, "could not dial %s", path)
	}
	go closeOn(c, syscall.SIGINT, syscall.SIGTERM)

	r := &ApiExecCommandRequest{
		Id:   &e.instanceId,
		Body: koyeb.NewExecCommandRequestBody(),
	}
	r.Body.SetCommand(e.cmd)
	err = e.pushOne(ctx, c, r)
	if err != nil {
		return -1, errors.Wrap(err, "could not intialize RPC with server")
	}

	pushErrCh := e.pushMany(ctx, c, e.stdin)
	termResizeErrCh := e.pushTermResizes(ctx, c, e.termResizeCh)
	listenErrCh, exitCh := e.report(ctx, c, e.stdout, e.stderr)

	for {
		select {
		// Context done or cancelled, let's stop there
		case <-ctx.Done():
			return -1, ctx.Err()
		// The server sent us an exit code. That means that the command has finished
		// its execution. Clean exit
		case exitCode := <-exitCh:
			return exitCode, nil
		// Something went wrong while listening for server messages or while transmitting
		// to stdout/stderr
		case err := <-listenErrCh:
			return 0, err
		// Something went wrong while sending messages to the server or while reading
		// user input
		case err := <-pushErrCh:
			return 0, err
		case err := <-termResizeErrCh:
			return 0, err
		}
	}
}

func (e *Executor) dial(address, path string) (*websocket.Conn, error) {
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

func (e *Executor) pushOne(ctx context.Context, c *websocket.Conn, r *ApiExecCommandRequest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if c == nil {
		return errors.New("fatal: need an open connection")
	}
	err := c.WriteJSON(r)
	if err != nil {
		return errors.Wrap(err, "failed sending data to remote server")
	}
	return nil
}

func (e *Executor) pushTermResizes(ctx context.Context, c *websocket.Conn, from <-chan *TerminalSize) <-chan error {
	errChan := make(chan error)
	go func() {
		if c == nil {
			errChan <- errors.New("fatal: need an open connection")
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case r := <-from:
				if r == nil {
					errChan <- errors.New("cannot resize term to nil size")
					return
				}
				resize := koyeb.NewExecCommandRequestTerminalSize()
				resize.SetHeight(r.Height)
				resize.SetWidth(r.Width)

				body := koyeb.NewExecCommandRequestBody()
				body.SetTtySize(*resize)

				err := c.WriteJSON(&ApiExecCommandRequest{
					Body: body,
				})
				if err != nil {
					errChan <- errors.Wrap(err, "failed sending term resize to remote server")
					return
				}
			}
		}
	}()
	return errChan
}

func (e *Executor) pushMany(ctx context.Context, c *websocket.Conn, from io.Reader) <-chan error {
	errChan := make(chan error)
	if c == nil {
		errChan <- errors.New("fatal: need an open connection")
		return errChan
	}

	go func() {
		data := make([]byte, 4096)
		for {
			if ctx.Err() != nil {
				return
			}

			n, err := from.Read(data)

			if err != nil && !errors.Is(err, io.EOF) {
				errChan <- errors.Wrap(err, "failed reading data from input")
				return
			}

			if n != 0 {
				io := koyeb.NewExecCommandIO()
				io.SetData(base64.StdEncoding.EncodeToString(data[:n]))

				body := koyeb.NewExecCommandRequestBody()
				body.SetCommand(e.cmd)
				body.SetStdin(*io)

				writeErr := c.WriteJSON(&ApiExecCommandRequest{
					Id:   &e.instanceId,
					Body: body,
				})
				if writeErr != nil {
					errChan <- errors.Wrap(writeErr, "failed sending data to remote server")
					return
				}
			}
			if errors.Is(err, io.EOF) {
				return
			}

		}
	}()

	return errChan
}

func (e *Executor) report(ctx context.Context, c *websocket.Conn, stdout, stderr io.Writer) (<-chan error, <-chan int) {
	errCh := make(chan error)
	exitCodeCh := make(chan int)
	if c == nil {
		errCh <- errors.New("fatal: need an open connection")
		return errCh, exitCodeCh
	}

	go func() {
		for {
			if ctx.Err() != nil {
				return
			}

			frame := &koyeb.StreamResultOfExecCommandReply{}
			err := c.ReadJSON(frame) // This blocks

			//TODO: Remove this. At some point, we should never stop by ourselves. In fact,
			// the stop conditions should be:
			// * the server sends us a frame with exited == true
			// * an unexpected error arises
			// Currently, the server is not correctly plugged-in to nomad, so it does not
			// send stop signals. Hence, we resort to CTRL+C from the client. In the future,
			// CTRL+C and CTRL+D will be sent all the way to the server which will handle them
			// and exit if needed
			if err != nil && websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				// Two cases:
				// * we closed the connection by sending the server a close control msg and
				// this is its ack answer.
				// * the server wishes to close the connection and is proactively sending a
				// close control msg. The default client close handler will take care of
				// sending back a close ack answer.
				// In both cases, that means that the connection ended normally, we can safely
				// exit
				exitCodeCh <- -1
				return
			}

			if err != nil {
				errCh <- errors.Wrap(err, "remote read failed")
				return
			}
			if frame.Error != nil {
				// This is probably bad
				errCh <- e.cast(frame.Error)
				return
			}
			err = e.forwardStdIO(frame, stdout, stderr)
			if err != nil {
				errCh <- errors.Wrap(err, "reporting to stdio failed")
				return
			}
			if frame.Result != nil && *frame.Result.Exited {
				exitCodeCh <- int(*frame.Result.ExitCode)
				return
			}
		}
	}()
	return errCh, exitCodeCh
}

func (e *Executor) forwardStdIO(f *koyeb.StreamResultOfExecCommandReply, stdout, stderr io.Writer) error {
	forwardFn := func(src *koyeb.ExecCommandIO, dst io.Writer) error {
		if src == nil {
			return nil
		}
		if src.Data == nil {
			return nil
		}
		buf, err := base64.StdEncoding.DecodeString(*src.Data)
		if err != nil {
			log.Errorf("could not base64decode %s", *src.Data)
			return errors.New("could not decode received data")
		}
		_, err = dst.Write(buf)
		return err
	}

	if f == nil {
		return nil
	}
	if f.Result == nil {
		return nil
	}
	err := forwardFn(f.Result.Stderr, stderr)
	if err != nil {
		return errors.Wrap(err, "reporting to stderr failed")
	}
	err = forwardFn(f.Result.Stdout, stdout)
	if err != nil {
		return errors.Wrap(err, "reporting to stdout failed")
	}
	return nil

}

func (e *Executor) cast(s *koyeb.GoogleRpcStatus) error {
	code := s.GetCode()
	msg := s.GetMessage()
	return fmt.Errorf("server failure: %s (code %d)", msg, code)
}

type StdStreams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func GetStdStreams() (*StdStreams, func() error, error) {
	stdin, stdout, stderr := term.StdStreams()
	fd, isTerm := term.GetFdInfo(stdin)
	if !isTerm {
		return nil, nil, errors.New("stdin is not a terminal")
	}

	termState, err := term.SetRawTerminal(fd)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not make stdin a raw terminal")
	}

	resetTermState := func() error {
		return term.RestoreTerminal(fd, termState)
	}

	stdStreams := &StdStreams{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	return stdStreams, resetTermState, nil
}

func GetTermSize(t io.Writer) (*TerminalSize, error) {
	fd, isTerm := term.GetFdInfo(t)
	if !isTerm {
		return nil, errors.New("not a terminal")
	}
	ws, err := term.GetWinsize(fd)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get winsize")
	}
	return &TerminalSize{Height: int32(ws.Height), Width: int32(ws.Width)}, nil
}
