package cmd

import (
  "log"
  "os/exec"
  "os"
  "fmt"
  "github.com/spf13/cobra"
  "github.com/cobookman/gcp-quickstart/layout"
  "github.com/cobookman/gcp-quickstart/renders"
  "github.com/fatih/color"
)

var layoutSheetId string
var clientSecretPath string
var isClean bool
var domain string
var gaID string

var buildCmd = &cobra.Command{
	Use: "build",
	Short: "Build the GCP Quickstart webpage",
	Long: "Builds the GCP Quickstart webpage's static assets under the build/ folder",
	Run: func(cmd *cobra.Command, args []string) {
    Build()
  },
}

func init() {
  buildCmd.Flags().BoolVarP(&isClean, "clean", "c", false, "Clean before buliding")
  buildCmd.Flags().StringVarP(&layoutSheetId, "layout-sheet-id", "l", "1-Nj5UkRGfD-9N6zj3B7mYXrJFvOgxmzm3RXv2cLeAh4", "Id of the google sheet containing the layout")
  buildCmd.Flags().StringVarP(&clientSecretPath, "client_secret", "s", "client_secret.json", "path to the oauth client secret")
  buildCmd.Flags().StringVarP(&domain, "domain", "d", "https://example.com", "root domain url. used in cononical metadata")
  buildCmd.Flags().StringVarP(&gaID, "ga", "g", "", "Google Analytics ID")
}

func Build() {
  if isClean {
    Clean()
  }

  color.Red("Copying Statics")
  os.MkdirAll("build/statics/", os.ModePerm)
  if err := exec.Command("cp", "-r", "statics/", "build/").Run(); err != nil {
    log.Fatal(err)
  }

  color.Red("Getting Layout")
  layout, err := layout.GetLayout(clientSecretPath, layoutSheetId)
  if err != nil {
    log.Fatal(err)
  }

  color.Red("Building Webpages")
  if err := buildCategories(layout); err != nil {
    log.Fatal(err)
  }

  if err := buildLessons(layout); err != nil {
    log.Fatal(err)
  }

  if err := buildOthers(clientSecretPath, layout); err != nil {
    log.Fatal(err)
  }
}

func buildOthers(clientSecretPath string, layout *layout.Layout) error {
  color.Blue("Building Other Pages")
  for _, other := range layout.Others {
    color.Magenta("\tBuilding Other: " + other.URL)
    gr , err := renders.RenderGdoc(layout, clientSecretPath, other.SourceGDoc, "build", other.URL, domain)
    if err != nil {
      return err
    }

    fmt.Printf("\t\tTitle: %s\n\t\tSummary: %s\n\t\tAuthor: %s\n\t\tImage: %s\n",
      gr.Metadata.Title, gr.Metadata.Summary, gr.Metadata.Author, gr.Metadata.Image)
  }
  return nil
}

func buildCategories(layout *layout.Layout) error {
  color.Blue("Building Category Pages")
  for _, category := range layout.Categories {
    color.Magenta("\tBuilding Category: " + category.Name)
    if err := renders.RenderCategory(layout, category, "build", domain); err != nil {
      return err
    }
    fmt.Println("\t\tBuilt")
  }
  return nil
}

func buildLessons(layout *layout.Layout) error {
  color.Blue("Building Lesson Pages")
  for _, lesson := range layout.Lessons {
    color.Magenta("\tBuilding lesson: " + lesson.Name)

    if len(lesson.SourceClaat) != 0 {
      color.Cyan("\t\tBuilding claat")
      if err := renders.RenderClaat(lesson, gaID, "build"); err != nil {
        return err
      }
    } else if len(lesson.SourceGDoc) != 0 {
      fmt.Println("\t\tBuilding Gdoc")

    } else if len(lesson.SourceURL) != 0 {
      fmt.Println("\t\tUsing Source URL")

    } else {
      color.Red("\t\tLesson has no source, skipping")
    }
  }
  return nil
}
