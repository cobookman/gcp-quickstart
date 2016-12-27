package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type Section struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Product struct {
	Name     string     `json:"name"`
	Acronym  string     `json:"acronym"`
	Sections []*Section `json:"sections"`
	ID       string     `json:"-"`
}

type Grouping struct {
	Name         string     `json:"name"`
	LinkInHeader bool       `json:"linkInHeader"`
	Products     []*Product `json:"products"`
	URL          string     `json:"url"`
}

type Source struct {
	GdocURL      string `json:"gdocUrl"`
	IsClaat      bool   `json:"isClaat"`
	IsFreeForm   bool   `json:"isFreeForm"`
	RenderedHTML template.HTML
}

type Social struct {
	Context       string    `json:"@context"`
	Type          string    `json:"@type"`
	Headline      string    `json:"headline"`
	DatePublished time.Time `json:"datePublished"`
	Image         []string  `json:"iamge"`
}

type Layout struct {
	ContentList []*Grouping
	Source      Source
	Title       string
	DomainRoot  string
	FilePath    string
	Social      Social
}

func main() {
	setup()

	templates := template.Must(template.ParseGlob("templates/*"))

	layout, err := GetLayout()
	if err != nil {
		log.Fatal(err)
	}

	// Render index page
	if err := RenderMiddleware(*layout, templates, "/index.html", RenderIndex); err != nil {
		log.Fatal(err)
	}
}

func GetLayout() (*Layout, error) {
	data, err := ioutil.ReadFile("layout.json")
	if err != nil {
		return nil, err
	}

	layout := &Layout{}
	if err := json.Unmarshal(data, &layout.ContentList); err != nil {
		return nil, err
	}

	// Attach defaults
	layout.Social.Context = "http://schema.org"
	layout.Social.Type = "NewsArticle"
	layout.DomainRoot = "http://localhost:8000"

	// Generate relevant URLs
	for _, grouping := range layout.ContentList {
		grouping.URL = layout.DomainRoot + "/lessons/" + grouping.Name + ".html"
		for _, product := range grouping.Products {
			if len(product.Acronym) != 0 {
				product.ID = product.Acronym
			} else {
				product.ID = product.Name
			}

			for _, section := range product.Sections {
				section.URL = layout.DomainRoot + "/lessons/" + grouping.Name + "/" + product.Name + "/" + section.Name + ".html"
			}
		}
	}

	return layout, nil
}

func setup() error {
	var err error
	log.Print("Cleaning any past build")
	err = os.RemoveAll("build")
	if err != nil {
		return err
	}

	log.Print("Creating build directory")
	err = os.Mkdir("build", 0777)
	if err != nil {
		return err
	}
	return nil
}

type RenderFn func(Layout, *template.Template, *os.File) error

func RenderMiddleware(layout Layout, templates *template.Template, filePath string, renderFn RenderFn) error {
	f, err := os.Create("build" + filePath)
	if err != nil {
		return err
	}

	layout.FilePath = filePath
	return renderFn(layout, templates, f)
}
