//go:build !windows
// +build !windows

package koyeb

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/moby/term"

	"github.com/pkg/errors"
)

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
				termSize, err := getTermSize(s.Stdout)
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

func getTermSize(t io.Writer) (*TerminalSize, error) {
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
