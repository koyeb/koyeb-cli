package main

import (
	"log"
	"path"
	"strings"

	"github.com/koyeb/koyeb-cli/pkg/koyeb"
	"github.com/spf13/cobra/doc"
)

func genMarkdownDocumentation() {
	rootCmd := koyeb.GetRootCmd()
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return "#" + strings.Replace(strings.ToLower(base), "_", "-", -1)
	}

	filePrepender := func(filename string) string {
		if filename == "docs/koyeb.md" {
			return `---
title: "Koyeb CLI Reference"
shortTitle: Reference
description: "Discover all the commands available via the Koyeb CLI and how to use them to interact with the Koyeb serverless platform directly from the terminal."
---

# Koyeb CLI Reference

The Koyeb CLI allows you to interact with Koyeb directly from the terminal. This documentation references all commands and options available in the CLI.

If you have not installed the Koyeb CLI yet, please read the [installation guide](/build-and-deploy/cli/installation).
`
		}
		return ""
	}

	err := doc.GenMarkdownTreeCustom(rootCmd, "./docs", filePrepender, linkHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	genMarkdownDocumentation()
}
