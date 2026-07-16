package cmd

import (
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/opsybot/opsybot/internal"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply database migrations",
	RunE: func(cmd *cobra.Command, _ []string) error {
		migrator, cleanup, err := internal.InitMigrator(cfgFile)
		if err != nil {
			return err
		}
		defer cleanup()

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return migrator.Migrate(ctx)
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
