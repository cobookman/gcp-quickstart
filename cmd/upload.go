package cmd

import (
  "log"
  "github.com/spf13/cobra"
  "os/exec"
  "os"
  "github.com/fatih/color"
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
  color.Red("Uploading Build")
  deployCmd := exec.Command("firebase", "deploy")
  deployCmd.Stdout = os.Stdout
  deployCmd.Stderr = os.Stderr
  if err := deployCmd.Start(); err != nil {
    log.Fatal(err)
  }
  if err := deployCmd.Wait(); err != nil {
    log.Fatal(err)
  }
}
