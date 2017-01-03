package cmd

import (
  "log"
  "github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use: "clean",
	Short: "Cleans all built assets",
	Long: "Removes the build/ folder",
	Run: func(cmd *cobra.Command, args []string) {
    Clean()
  },
}

func Clean() {
  log.Print("Clean...")
}
