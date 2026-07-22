package cmd

import (
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/opsybot/opsybot/internal"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Run the opsybot background jobs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		worker, cleanup, err := internal.InitWorker(cfgFile)
		if err != nil {
			return err
		}
		defer cleanup()

		ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return worker.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(cronCmd)
}
