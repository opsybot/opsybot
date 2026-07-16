package cmd

import (
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/opsybot/opsybot/internal"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the opsybot server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		app, cleanup, err := internal.InitApp(cfgFile)
		if err != nil {
			return err
		}
		defer cleanup()

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return app.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
