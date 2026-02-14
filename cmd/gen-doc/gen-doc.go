package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/koyeb/koyeb-cli/pkg/koyeb"
	"github.com/spf13/cobra/doc"
)

func genMarkdownDocumentation(outputDir string) {
	rootCmd := koyeb.GetRootCommand()
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return "#" + strings.ReplaceAll(strings.ToLower(base), "_", "-")
	}

	filePrepender := func(filename string) string {
		if filename == outputDir+"/koyeb.md" {
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

	err := doc.GenMarkdownTreeCustom(rootCmd, outputDir, filePrepender, linkHandler)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	outputDir := "./docs"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	genMarkdownDocumentation(outputDir)
}
