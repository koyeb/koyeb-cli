//go:build windows
// +build windows

package koyeb

import (
	"context"
)

func watchTermSize(ctx context.Context, s *StdStreams) <-chan *TerminalSize {
	return nil
}
