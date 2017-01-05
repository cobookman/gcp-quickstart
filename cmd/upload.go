package cmd

import (
  "log"
  "os"
  "os/exec"
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
  log.Print("Uploading")
  uploadCmd := exec.Command("firebase", "deploy")
	uploadCmd.Stdout = os.Stdout
	uploadCmd.Stderr = os.Stderr
	if err := uploadCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := uploadCmd.Wait(); err != nil {
		log.Fatal(err)
	}

}
