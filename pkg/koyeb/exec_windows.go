//go:build windows
// +build windows

package koyeb

import (
	"context"
	"io"
)

func watchTermSize(ctx context.Context, s io.Writer) <-chan *TerminalSize {
	return nil
}
