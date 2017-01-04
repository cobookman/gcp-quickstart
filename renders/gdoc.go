package renders

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/cobookman/gcp-quickstart/apiclients"
	"github.com/cobookman/gcp-quickstart/layout"
	"github.com/cobookman/gcp-quickstart/templates"
	"github.com/satori/go.uuid"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

const (
	commentPrefix = "#cmnt"
)

type GdocMetadata struct {
	Title   string
	Summary string
	Author  string
	Image   string
}

type GdocRender struct {
	Metadata    *GdocMetadata
	ArticleHTML string
	Source      string
	Path        string
	BuildFolder string
	Domain 		 	string
	Layout *layout.Layout
}

// Grabs the contents from a gdoc, downloads all images to the build folder.
// fixes some html issues, and pases up any errors
func RenderGdoc(layout *layout.Layout, clientSecretPath string, gdocURL string, buildFolder string, htmlPath string, domain string) (*GdocRender, error) {
	gr := &GdocRender{
		Source:      gdocURL,
		Path:        htmlPath,
		BuildFolder: buildFolder,
		Domain: domain,
		Layout: layout,
	}

	resp, err := apiclients.GetGdocHtml(clientSecretPath, gr.ID())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}
	body := doc.Find("body")

	if err = gr.parseMetadata(body); err != nil {
		return nil, err
	}

	if err = gr.renderArticleBody(body); err != nil {
		return nil, err
	}

	if err = gr.write(); err != nil {
		return nil, err
	}

	return gr, err
}

// Writes the article out
func (gr GdocRender) write() error {
	pg := &templates.PageMetadata{
		Title: gr.Metadata.Title,
		ArticleHTML: templates.RenderHTML(gr.ArticleHTML),
		FilePath: gr.Path,
		Domain: gr.Domain,
		Social: &templates.Social{
			Headline: gr.Metadata.Title,
			DatePublished: time.Now(),
			Image: []string{gr.Metadata.Image},
		},
		Layout: gr.Layout,
	}

	return templates.RenderPage(pg, gr.BuildFolder)
}

// Gives the gdoc's ID from parsing source url.
func (gr GdocRender) ID() string {
	const s = "/document/d/"
	gdocId := gr.Source
	if i := strings.Index(gdocId, s); i >= 0 {
		gdocId = gdocId[i+len(s):]
	}
	if i := strings.IndexRune(gdocId, '/'); i > 0 {
		gdocId = gdocId[:i]
	}
	return gdocId
}

// Parses the document's metadata which will be used for things like social media
// and meta tags. Metadata is attached to the GdocRender struct.
func (gr *GdocRender) parseMetadata(body *goquery.Selection) error {
	metadata := new(GdocMetadata)
	trs := body.Find("table").First().Find("tr")
	var parseError error

	trs.EachWithBreak(func(i int, tr *goquery.Selection) bool {
		columns := tr.Find("td")
		if columns.Length() != 2 {
			parseError = errors.New("row has more than 2 columns (td): " + columns.Text())
			return false
		}

		columnName := columns.First().Text()
		columnValue := columns.Next().First()

		switch strings.ToUpper(columnName) {
		case "TITLE":
			metadata.Title = columnValue.Text()

		case "SUMMARY":
			metadata.Summary = columnValue.Text()

		case "AUTHOR":
			metadata.Author = columnValue.Text()

		case "IMAGE":
			url, ok := columnValue.Find("img").First().Attr("src")
			if !ok {
				parseError = errors.New("Image does not have a source: " + columnValue.Text())
				return false
			}
			fname, _, err := gr.downloadImage(url)
			if err != nil {
				parseError = err
				return false
			}
			metadata.Image = fname
		}
		return true
	})

	if parseError == nil {
		gr.Metadata = metadata
	}
	return parseError
}

// downloads an image to the build folder. Returns the saved image's filename.
// so total path to image is downloadFolder + imageFileName
func (gr GdocRender) downloadImage(imageUrl string) (string, *image.Rectangle, error) {
	resp, err := http.Get(imageUrl)
	if err != nil {
		return "", nil, err
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	// generate a unique filename
	fname := "/img/" + uuid.NewV4().String()

	// parse the file extension for the image if possible
	mimeType := resp.Header.Get("content-type")
	if len(mimeType) != 0 {
		fname += "." + strings.Replace(mimeType, "image/", "", 1)
	}

	// Save image to build folder
	if err := os.MkdirAll(gr.BuildFolder+"/img/", os.ModePerm); err != nil {
		return "", nil, err
	}
	f, err := os.Create(gr.BuildFolder + fname)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	_, err = io.Copy(f, bytes.NewReader(b))
	if err != nil {
		return "", nil, err
	}

	// get image dimensions
	m, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return "", nil, err
	}
	bounds := m.Bounds()

	return fname, &bounds, nil
}

// Cleans up the document's html to only include the relavent styling
func (gr *GdocRender) renderArticleBody(body *goquery.Selection) error {
	seenMetadataTable := false

	// Go through root elements one by one and re-style them to correct DOM
	var cleaningError error
	articleHTML := ""
	body.Children().EachWithBreak(func(i int, ns *goquery.Selection) bool {
		n := ns.Get(0)

		// Ignore all nodes until after the first table which stores the metadata
		if !seenMetadataTable {
			if n.DataAtom == atom.Table {
				seenMetadataTable = true
			}
			return true
		}

		// Clean children of this node
		n, err := gr.cleanNode(n)
		if err != nil {
			cleaningError = err
			return false
		}

		// // if resulting node is not empty after cleaning, render it out
		if n != nil {
			buf := new(bytes.Buffer)
			html.Render(buf, n)
			articleHTML += buf.String()
		}

		return true
	})
	gr.ArticleHTML = articleHTML
	return cleaningError
}

// Cleans up a given node
func (gr GdocRender) cleanNode(n *html.Node) (*html.Node, error) {
	if n == nil {
		return nil, nil
	}

	// clean up children first
	c := n.FirstChild
	for c != nil {
		next := c.NextSibling
		if _, err := gr.cleanNode(c); err != nil {
			return nil, err
		}
		c = next
	}

	// remove if empty element or a comment element
	isEmptyElement := (n.DataAtom == atom.Span || n.DataAtom == atom.P ||
		n.DataAtom == atom.Div || n.DataAtom == atom.Sup) && n.FirstChild == nil
	isCommentElement := n.DataAtom == atom.Div && nodeAttr(n, "style") == "border:1px solid black;margin:5px"
	if isEmptyElement || isCommentElement {
		n.Parent.RemoveChild(n)
		return nil, nil
	}

	// fix html
	var switchErr error
	switch n.DataAtom {
	case atom.Img:
		n, switchErr = gr.cleanAtomImg(n)
	case atom.A:
		n, switchErr = gr.cleanAtomA(n)
	case atom.Span:
		n, switchErr = gr.cleanAtomSpan(n)
	}

	// handle errors occuring in the switch
	if switchErr != nil {
		return nil, switchErr
	}
	if n != nil {
		gr.cleanAttributes(n)
	}
	return n, nil
}

// cleans up a <span> node
func (gr GdocRender) cleanAtomSpan(n *html.Node) (*html.Node, error) {
	// if span only contians text or a single child, replace the span with
	// the contents of its singular child.
	if n.FirstChild == n.LastChild {
		newN := n.FirstChild
		n.FirstChild = newN.FirstChild
		n.LastChild = newN.LastChild
		n.Type = newN.Type
		n.DataAtom = newN.DataAtom
		n.Data = newN.Data
		n.Namespace = newN.Namespace
		n.Attr = newN.Attr
	}
	return n, nil
}

// cleans up a <img> node
func (gr GdocRender) cleanAtomImg(n *html.Node) (*html.Node, error) {
	n.DataAtom = 0x0
	n.Data = "amp-img"

	fname, bounds, err := gr.downloadImage(nodeAttr(n, "src"))
	if err != nil {
		return nil, err
	}

	setNodeAttr(n, "src", fname)
	setNodeAttr(n, "width", strconv.Itoa(bounds.Dx()))
	setNodeAttr(n, "height", strconv.Itoa(bounds.Dy()))
	setNodeAttr(n, "layout", "responsive")
	return n, nil
}

// cleans up an <a> node.
func (gr GdocRender) cleanAtomA(n *html.Node) (*html.Node, error) {
	// delete elment if a link to a comment
	if strings.HasPrefix(nodeAttr(n, "href"), commentPrefix) {
		n.Parent.RemoveChild(n)
		return nil, nil
	}

	// fix href url to be that of actual domain and not a google redirect
	u, err := url.Parse(nodeAttr(n, "href"))
	if err != nil {
		return nil, err
	}
	if q, ok := u.Query()["q"]; ok {
		setNodeAttr(n, "href", q[0])
	}
	return n, nil
}

// cleans all stlyes, id, classes, and titles from the node.
func (gr GdocRender) cleanAttributes(n *html.Node) {
	delNodeAttr(n, "style")
	delNodeAttr(n, "id")
	delNodeAttr(n, "class")
	delNodeAttr(n, "title")
}

func setNodeAttr(n *html.Node, key string, val string) {
	key = strings.ToLower(key)
	for i := 0; i < len(n.Attr); i++ {
		if strings.ToLower(n.Attr[i].Key) == key {
			n.Attr[i].Val = val
			return
		}
	}
	// attribute does not exist already, append it
	n.Attr = append(n.Attr, html.Attribute{
		Key: key,
		Val: val,
	})
}

func delNodeAttr(n *html.Node, name string) {
	name = strings.ToLower(name)
	for i := 0; i < len(n.Attr); i++ {
		if strings.ToLower(n.Attr[i].Key) == name {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			return
		}
	}
}

// Get an html node's attribute matching to the string name. If no matching
// attributes found, then an empty string is returned.
func nodeAttr(node *html.Node, name string) string {
	name = strings.ToLower(name)
	for _, attr := range node.Attr {
		if strings.ToLower(attr.Key) == name {
			return attr.Val
		}
	}
	return ""
}
