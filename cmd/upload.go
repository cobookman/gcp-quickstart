package cmd

import (
  "log"
  "github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use: "upload",
	Short: "Uploads the GCP Quickstart webpage",
	Long: "Uploads the GCP Quickstart webpage's static assets in the build/ folder",
	Run: func(cmd *cobra.Command, args []string) {
    Upload()
  },
}



func Upload() {
  log.Print("Upload...")
}
