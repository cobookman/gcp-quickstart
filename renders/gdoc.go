package renders
import (
  "strings"
  "fmt"
  "io"
  "net/http"
  "os"
  "errors"
  "golang.org/x/net/html/atom"
  "github.com/cobookman/gcp-quickstart/apiclients"
  "github.com/PuerkitoBio/goquery"
  "github.com/satori/go.uuid"
)

type GdocMetadata struct {
  Title string
  Summary string
  Author string
  Image string
}

func RenderGdoc(clientSecretPath string, gdocURL string, buildFolder string, htmlPath string) error {
  resp, err := apiclients.GetGdocHtml(clientSecretPath, gdocID(gdocURL))
  if err != nil {
    return err
  }

  defer resp.Body.Close()
  doc, err := goquery.NewDocumentFromResponse(resp)
  if err != nil {
    return err
  }
  body := doc.Find("body")


  metadata, err := parseMetadata(body, buildFolder)
  if err != nil {
    return err
  }

  htmlBody, err := cleanHtml(body, buildFolder)
  if err != nil {
    return err
  }
  fmt.Printf("Metdata: %v, Body: %v\n", metadata, htmlBody)

  return nil
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
      fmt.Println("Image...")
      url, ok := columnValue.Find("img").First().Attr("src")
      if !ok {
        parseError = errors.New("Image does not have a source: " + columnValue.Text())
        return false
      }
      fmt.Println("URL...: " + url)
      fname, err  := downloadImage(url, buildFolder)
      if err != nil {
        fmt.Printf("Error: %v\n", err)
        parseError = err
        return false
      }
      fmt.Println("fname...: " + fname)
      metadata.Image = buildFolder + "/" + fname
    }
    return true
  })

  return metadata, nil
}

// downloads an image to the build folder. Returns the saved image's filename.
// so total path to image is downloadFolder + imageFileName
func downloadImage(imageUrl string, buildFolder string) (string, error) {
  resp, err := http.Get(imageUrl)
  if err != nil {
    return "", err
  }
  defer resp.Body.Close()

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
    return "", err
  }
  defer f.Close()
  _, err = io.Copy(f, resp.Body)
  if err != nil {
    return "", err
  }

  return fname, nil
}

// Cleans up the document's html to only include the relavent styling
func cleanHtml(body *goquery.Selection, buildFolder string) (string, error) {
  seenMetadataTable := false

  body.Children().EachWithBreak(func(i int, node *goquery.Selection) bool {
    parentNode := node.Get(0)
    fmt.Printf("Parent: %v\n", parentNode)
    if !seenMetadataTable {
      if parentNode.DataAtom == atom.Table{
        seenMetadataTable = true
      }
      return true
    }

    fmt.Printf("HTML: %v\n", node.Text())
    return true
  })
  return "", nil
}
