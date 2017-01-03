package templates

import (
  "os"
  "html/template"
  "time"
  "path/filepath"
  "github.com/cobookman/gcp-quickstart/layout"
)

var (
  templates *template.Template
)

type Social struct {
  Context       string    `json:"@context"`
	Type          string    `json:"@type"`
	Headline      string    `json:"headline"`
	DatePublished time.Time `json:"datePublished"`
	Image         []string  `json:"iamge"`
}

type PageMetadata struct {
  Title string
  ArticleHTML template.HTML
  FilePath string
  Domain string
  Social *Social
  Layout *layout.Layout
}

func RenderHTML(html string) (template.HTML) {
  return template.HTML(html)
}

func Templates() *template.Template {
  if templates == nil {
    templates = template.Must(template.ParseGlob("templates/*"))
  }
  return templates
}

func getFile(pg *PageMetadata, buildFolder string) (*os.File, error) {
  if err := os.MkdirAll(filepath.Join(buildFolder, filepath.Dir(pg.FilePath)), os.ModePerm); err != nil {
    return nil, err
  }

  f, err := os.Create(filepath.Join(buildFolder, pg.FilePath))
  if err != nil {
    return nil, err
  }
  return f, nil
}

func RenderPage(pg *PageMetadata, buildFolder string) error {
  f, err := getFile(pg, buildFolder)
  if err != nil {
    return err
  }

  return Templates().ExecuteTemplate(f, "page", pg)
}
