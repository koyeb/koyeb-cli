package koyeb

import (
	"github.com/spf13/cobra"
)

func NewRegionCmd() *cobra.Command {
	h := NewRegionHandler()

	regionCmd := &cobra.Command{
		Use:     "regions ACTION",
		Aliases: []string{"region", "reg"},
		Short:   "Regions",
	}

	listRegionCmd := &cobra.Command{
		Use:   "list",
		Short: "List regions",
		RunE:  WithCLIContext(h.List),
	}
	regionCmd.AddCommand(listRegionCmd)

	getRegionCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get region",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	regionCmd.AddCommand(getRegionCmd)

	return regionCmd
}

type RegionHandler struct {
}

func NewRegionHandler() *RegionHandler {
	return &RegionHandler{}
}
