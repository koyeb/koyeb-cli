package koyeb

import (
	"time"

	"github.com/spf13/cobra"
)

func GetDurationFlags(cmd *cobra.Command, name string) time.Duration {
	val, _ := cmd.Flags().GetDuration(name)
	return val
}

func GetBoolFlags(cmd *cobra.Command, name string) bool {
	val, _ := cmd.Flags().GetBool(name)
	return val
}

func GetStringFlags(cmd *cobra.Command, name string) string {
	val, _ := cmd.Flags().GetString(name)
	return val
}
