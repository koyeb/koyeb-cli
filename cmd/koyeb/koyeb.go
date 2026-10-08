package main

import (
	"os"

	"github.com/koyeb/koyeb-cli/pkg/koyeb"
)

func main() {
	os.Exit(koyeb.Run(os.Args[1:]))
}
