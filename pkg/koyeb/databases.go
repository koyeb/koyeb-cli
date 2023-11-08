package koyeb

import "github.com/spf13/cobra"

func NewDatabaseCmd() *cobra.Command {
	h := NewDatabaseHandler()

	databaseCmd := &cobra.Command{
		Use:     "databases ACTION",
		Aliases: []string{"db", "database"},
		Short:   "Databases",
	}

	listDbCmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		RunE:  WithCLIContext(h.List),
	}
	databaseCmd.AddCommand(listDbCmd)

	return databaseCmd
}

func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{}
}

type DatabaseHandler struct {
}

func (h *DatabaseHandler) ResolveDatabaseArgs(ctx *CLIContext, val string) (string, error) {
	return "", nil
}
