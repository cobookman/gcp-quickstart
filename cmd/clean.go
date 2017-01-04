package cmd

import (
  "log"
  "os"
  "github.com/fatih/color"

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
  color.Red("Cleaning Build")
  if err := os.RemoveAll("build/"); err != nil {
    log.Fatal(err)
  }
}
