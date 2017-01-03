package renders
import (
  "strings"
  "bytes"
  "io"
  "strconv"
  "io/ioutil"
  "net/http"
  "net/url"

  "image"
  _ "image/gif"
  _ "image/jpeg"
  _ "image/png"
  "os"
  "errors"
  "golang.org/x/net/html"
  "golang.org/x/net/html/atom"
  "github.com/cobookman/gcp-quickstart/apiclients"
  "github.com/PuerkitoBio/goquery"
  "github.com/satori/go.uuid"
)

const (
  commentPrefix = "#cmnt"
)

type GdocMetadata struct {
  Title string
  Summary string
  Author string
  Image string
}

// Grabs the contents from a gdoc, downloads all images to the build folder.
// fixes some html issues, and pases up any errors
func RenderGdoc(clientSecretPath string, gdocURL string, buildFolder string, htmlPath string) (metadata *GdocMetadata, htmlBody string, err error) {
  resp, err := apiclients.GetGdocHtml(clientSecretPath, gdocID(gdocURL))
  if err != nil {
    return nil, "", err
  }

  defer resp.Body.Close()
  doc, err := goquery.NewDocumentFromResponse(resp)
  if err != nil {
    return nil, "", err
  }
  body := doc.Find("body")


  metadata, err = parseMetadata(body, buildFolder)
  if err != nil {
    return nil, "", err
  }

  htmlBody, err = cleanHtml(body, buildFolder)
  if err != nil {
    return nil, "", err
  }

  return metadata, htmlBody, err
}

func gdocID(gdocURL string) string {
  const s = "/document/d/"
  if i := strings.Index(gdocURL, s); i >= 0 {
    gdocURL = gdocURL[i+len(s):]
  }
  if i := strings.IndexRune(gdocURL, '/'); i > 0 {
    gdocURL = gdocURL[:i]
  }
  return gdocURL
}

// Parses the document's metadata which will be used for things like social media
// and meta tags
func parseMetadata(body *goquery.Selection, buildFolder string) (*GdocMetadata, error) {
  metadata := new(GdocMetadata)
  trs := body.Find("table").First().Find("tr")
  var parseError error

  trs.EachWithBreak(func(i int, tr *goquery.Selection) bool {
    columns := tr.Find("td")
    if columns.Length() != 2 {
      parseError  = errors.New("row has more than 2 columns (td): " + columns.Text())
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
      fname, _, err  := downloadImage(url, buildFolder)
      if err != nil {
        parseError = err
        return false
      }
      metadata.Image = fname
    }
    return true
  })

  return metadata, parseError
}

// downloads an image to the build folder. Returns the saved image's filename.
// so total path to image is downloadFolder + imageFileName
func downloadImage(imageUrl string, buildFolder string) (string, *image.Rectangle, error) {
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
  os.MkdirAll(buildFolder + "/img/", 0777)
  f, err := os.Create(buildFolder + fname)
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
func cleanHtml(body *goquery.Selection, buildFolder string) (string, error) {
  seenMetadataTable := false

  // Remove Comment Links From Doc
  body.Find("a").Each(func(i int, node *goquery.Selection) {
    a := node.Get(0)
    if strings.HasPrefix(nodeAttr(a, "href"), commentPrefix) {
      node.Remove()
    }
  });

  // Go through root elements one by one and re-style them to correct DOM
  outputHtml := ""
  var cleaningError error
  body.Children().EachWithBreak(func(i int, node *goquery.Selection) bool {
    parentNode := node.Get(0)

    // Ignore all nodes until after the first table which stores the metadata
    if !seenMetadataTable {
      if parentNode.DataAtom == atom.Table{
        seenMetadataTable = true
      }
      return true
    }

    parentNode, err := cleanNode(parentNode, buildFolder)
    if err != nil {
      cleaningError = err
      return false
    }

    if parentNode != nil {
      buf := new(bytes.Buffer)
      html.Render(buf, parentNode)
      outputHtml += buf.String()
    }

    return true
  })
  return outputHtml, cleaningError
}

// Cleans up a given node
func cleanNode(n *html.Node, buildFolder string) (*html.Node, error) {
  if n == nil {
    return nil, nil
  }

  // clean up children first
  c := n.FirstChild
  for c != nil {
    next := c.NextSibling
    if _, err := cleanNode(c, buildFolder); err != nil {
      return nil, err
    }
    c = next
  }

  // remove if empty element or a comment element
  isEmptyElement := (n.DataAtom == atom.Span || n.DataAtom == atom.P || n.DataAtom == atom.Div) && n.FirstChild == nil
  isCommentElement := n.DataAtom == atom.Div && nodeAttr(n, "style") == "border:1px solid black;margin:5px"
  if isEmptyElement || isCommentElement {
    n.Parent.RemoveChild(n)
    return nil, nil
  }

  // fix html
  switch n.DataAtom {
  case atom.Img:
    // TODO(bookman): Finish conversion to an amp-img
    n.DataAtom = 0x0
    n.Data = "amp-img"

    fname, bounds, err := downloadImage(nodeAttr(n, "src"), buildFolder)
    if err != nil {
      return nil, err
    }

    setNodeAttr(n, "src", fname)
    setNodeAttr(n, "width", strconv.Itoa(bounds.Dx()))
    setNodeAttr(n, "height", strconv.Itoa(bounds.Dy()))
    setNodeAttr(n, "layout", "responsive")

  case atom.A:
    // fix href url to be that of actual domain and not a google redirect
    u, err := url.Parse(nodeAttr(n, "href"))
    if err != nil {
      return nil, err
    }
    if q, ok := u.Query()["q"]; ok {
      setNodeAttr(n, "href", q[0])
    }
  }

  // remove unwanted attributes
  delNodeAttr(n, "style")
  delNodeAttr(n, "id")
  delNodeAttr(n, "class")
  delNodeAttr(n, "title")

  return n, nil
}

func setNodeAttr(node *html.Node, key string, val string) {
  key = strings.ToLower(key)
  for i := 0; i < len(node.Attr); i++ {
    if strings.ToLower(node.Attr[i].Key) == key {
      node.Attr[i].Val = val
      return
    }
  }
  // attribute does not exist already, append it
  node.Attr = append(node.Attr, html.Attribute{
    Key: key,
    Val: val,
  })
}

func delNodeAttr(node *html.Node, name string) {
  name = strings.ToLower(name)
  for i := 0; i < len(node.Attr); i++ {
    if strings.ToLower(node.Attr[i].Key) == name {
      node.Attr = append(node.Attr[:i], node.Attr[i+1:]...)
      return
    }
  }
}


// Get an html node's attribute matching to the string name. If no matching
// attributes found, then an empty string is returned.
func nodeAttr(node *html.Node, name string) (string) {
  name = strings.ToLower(name)
  for _, attr := range node.Attr {
    if strings.ToLower(attr.Key) == name {
      return attr.Val
    }
  }
  return ""
}
