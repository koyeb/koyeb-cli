package koyeb

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	getAllCommand = &cobra.Command{
		Use:     "all [resource]",
		Aliases: []string{"a"},
		Short:   "Get all",
		RunE:    getAll,
	}
)

func getAll(cmd *cobra.Command, args []string) error {
	fmt.Printf("Store:\n")
	getStores(cmd, args)

	fmt.Printf("\nSecret::\n")
	getSecrets(cmd, args)

	return nil
}
