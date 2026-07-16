package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "opsybot",
	Short: "Self-hosted on-call, incident response, and status pages",
	Long: `Opsybot is a self-hosted incident management platform: on-call
schedules, alerting, incident response, and status pages.

Configuration comes from OPSYBOT_* environment variables, an optional
--config file, and built-in defaults — in that order of precedence.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"path to config file (optional; OPSYBOT_* env vars take precedence)")
}
