package main

import (
	"os"

	"github.com/koyeb/koyeb-cli/pkg/koyeb"
)

func main() {
	if err := koyeb.Run(); err != nil {
		os.Exit(1)
	}
}
