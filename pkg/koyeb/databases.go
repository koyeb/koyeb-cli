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

	getDbCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get database",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	databaseCmd.AddCommand(getDbCmd)

	return databaseCmd
}

func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{}
}

type DatabaseHandler struct {
}

func (h *DatabaseHandler) ResolveDatabaseArgs(ctx *CLIContext, val string) (string, error) {
	databaseMapper := ctx.Mapper.Database()
	id, err := databaseMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
