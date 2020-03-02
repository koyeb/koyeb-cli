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
	fmt.Printf("Stacks:\n")
	getStacks(cmd, args)

	fmt.Printf("\nManaged stores:\n")
	getManagedStores(cmd, args)

	fmt.Printf("\nDeliveries:\n")
	getDeliveries(cmd, args)

	return nil
}
