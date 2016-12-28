package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"time"
)


type Claat struct {
	Gdoc      string `json:"gdoc"`
}

type Section struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Claat *Claat `json:"claat"`
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
	Path          string     `json:"url"`
}

type Social struct {
	Context       string    `json:"@context"`
	Type          string    `json:"@type"`
	Headline      string    `json:"headline"`
	DatePublished time.Time `json:"datePublished"`
	Image         []string  `json:"iamge"`
}

type Page struct {
	ContentList []*Grouping
	Title       string
	DomainRoot  string
	PagePath    string
	Social      Social
	RenderedHTML template.HTML
}

var (
	GoogleAnalytics = "UA-88560603-1"
)

func main() {
	setup()

	templates := template.Must(template.ParseGlob("templates/*"))

	page, err := GetBasePage()
	if err != nil {
		log.Fatal(err)
	}


	page.RenderSupportingContent()

	// Render a grouping page for nav & SEO
	for _, grouping := range page.ContentList {
		if err := os.MkdirAll("build/" + grouping.Path, 0777); err != nil {
			log.Fatal(err)
		}

		log.Printf("Rending grouping page: %s", grouping.Path)
		if err := RenderMiddleware(*page, templates, grouping.Path + "/index.html", RenderGrouping); err != nil {
			log.Fatal(err)
		}
	}

	// Render index page
	if err := RenderMiddleware(*page, templates, "/index.html", RenderIndex); err != nil {
		log.Fatal(err)
	}
}

func (page *Page) RenderSupportingContent() {
	// Render any claats
	for _, grouping := range page.ContentList {
		for _, product := range grouping.Products {
			for _, section := range product.Sections {
				if len(section.URL) == 0 {
					if section.Claat != nil {
						log.Printf("Rendering claat for: %s -> %s -> %s\n",
							grouping.Name, product.Name, section.Name)
						buildPath := "build/" + grouping.Name + "/" + product.Name + "/" + section.Name
						section.URL = page.DomainRoot + "/" + buildPath + "/index.html"
						if err := renderClaat(section.Claat.Gdoc, buildPath, GoogleAnalytics); err != nil {
							log.Fatal(err)
						}
					} else {
						log.Fatalf("%s -> %s -> %s section does not have a url or claat\n",
							grouping.Name, product.Name, section.Name)
					}
				}
			}
		}
	}
}


func GetBasePage() (*Page, error) {
	data, err := ioutil.ReadFile("layout.json")
	if err != nil {
		return nil, err
	}

	basePage := &Page{}
	if err := json.Unmarshal(data, &basePage.ContentList); err != nil {
		return nil, err
	}

	// Attach defaults
	basePage.Social.Context = "http://schema.org"
	basePage.Social.Type = "NewsArticle"
	basePage.DomainRoot = "http://localhost:8000"


	// Generate unique URLs for each grouping, and unique IDs for products
	for _, grouping := range basePage.ContentList {
		if len(grouping.Path) == 0 {
			grouping.Path = "/" + grouping.Name
		}
		for _, product := range grouping.Products {
			if len(product.Acronym) != 0 {
				product.ID = product.Acronym
			} else {
				product.ID = product.Name
			}
		}
	}


	return basePage, nil
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

type RenderFn func(Page, *template.Template, *os.File) error

func RenderMiddleware(basePage Page, templates *template.Template, filePath string, renderFn RenderFn) error {
	f, err := os.Create("build/" + filePath)
	if err != nil {
		return err
	}

	basePage.PagePath = filePath
	return renderFn(basePage, templates, f)
}
