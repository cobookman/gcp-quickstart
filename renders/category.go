package renders

import (
  "time"
  "fmt"
  "bytes"
  "github.com/cobookman/gcp-quickstart/templates"
  "github.com/cobookman/gcp-quickstart/layout"
)

func RenderCategory(layout *layout.Layout, category *layout.Category, buildFolder string, domain string) error {
  buf := new(bytes.Buffer)
  templates.Templates().ExecuteTemplate(buf, "category", category)
  categoryHtml := buf.String()

  pg := &templates.PageMetadata{
    Title: "GCP Quickstarts - " + category.Name,
    FilePath: "/" + category.ID + "/index.html",
    Domain: domain,
    ArticleHTML: templates.RenderHTML(categoryHtml),
    Social: &templates.Social{
      Headline: fmt.Sprintf("Getting started with Google Cloud's %s offerings", category.Name),
      DatePublished: time.Now(),
      Image: []string{},
    },
    Layout: layout,
  }

  return templates.RenderPage(pg, buildFolder)
}
