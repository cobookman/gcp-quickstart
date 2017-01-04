package layout

import (
  "log"
  "strings"
  "fmt"
  "github.com/cobookman/gcp-quickstart/apiclients"
)

type Lesson struct {
  Product *Product
  Name string
  Summary string
  SourceURL string
  SourceClaat string
  SourceGDoc string
  Href string
}

type Product struct {
  Category *Category
  ID string
  Name string
  Acronym string
  Summary string
  Icon string
  Lessons []*Lesson
}

type Category struct {
  ID string
  Name string
  Summary string
  InHeader bool
  Products []*Product
}

type Other struct {
  URL string
  SourceGDoc string
}

type Layout struct {
  Categories []*Category
  Products []*Product
  Lessons []*Lesson
  Others []*Other
}

func GetLayout(clientSecretPath string, spreadsheetId string) (*Layout, error) {
  log.Print("Getting Layout")
  srv, err := apiclients.NewSheetsClient(clientSecretPath)
  if err != nil {
    return nil, err
  }
  layout := new(Layout)

  // keep a reference to the tree structure
  categories := make(map[string]*Category)
  products := make(map[string]*Product)

  // Populate categories
  categoryColumns := "Categories!A2:D"
  categoryResp, err := srv.Spreadsheets.Values.Get(spreadsheetId, categoryColumns).Do()
  if err != nil {
    return nil, err
  }
  for _, row := range categoryResp.Values {
    log.Printf("Parsing - Category: %v\n", row)
    row = deNullifyRow(row)

    category := &Category{
      ID: assert(row[0].(string)),
      Name: assert(row[1].(string)),
      Summary: row[2].(string),
      InHeader: false,
    }

    if (strings.ToUpper(row[3].(string)) == "TRUE") {
      category.InHeader = true
    }

    categories[category.ID] = category
    layout.Categories = append(layout.Categories, category)
  }

  // Populate products
  productColumns := "Products!A2:F"
  productResp, err := srv.Spreadsheets.Values.Get(spreadsheetId, productColumns).Do()
  if err != nil {
    return nil, err
  }

  for _, row := range productResp.Values {
    log.Printf("Parsing - Product: %v\n", row)
    row = deNullifyRow(row)

    product := &Product{
      ID: assert(row[1].(string)),
      Name: assert(row[2].(string)),
      Acronym: row[3].(string),
      Summary: row[4].(string),
      Icon: row[5].(string),
    }

    // Attach Product's category
    productCategory := categories[row[0].(string)]
    product.Category = productCategory

    // Store lesson
    productCategory.Products = append(productCategory.Products, product)
    layout.Products = append(layout.Products, product)

    // Cache product in map
    products[product.ID] = product
  }

  lessonColumns := "Lessons!A2:F"
  lessonResp, err := srv.Spreadsheets.Values.Get(spreadsheetId, lessonColumns).Do()
  if err != nil {
    return nil, err
  }
  for _, row := range lessonResp.Values {
    log.Printf("Parsing - Lesson: %v\n", row)
    row = deNullifyRow(row)

    lesson := &Lesson{
      Name: assert(row[1].(string)),
      Summary: row[2].(string),
      SourceURL: row[3].(string),
      SourceClaat: row[4].(string),
      SourceGDoc: row[5].(string),
    }

    // Attach lesson's product
    lessonProduct := products[row[0].(string)]
    lesson.Product = lessonProduct

    // populate lesson's href
    if len(lesson.SourceURL) != 0 {
      lesson.Href = lesson.SourceURL
    } else if len(lesson.SourceClaat) != 0 || len(lesson.SourceGDoc) != 0 {
      lesson.Href = fmt.Sprintf("/%s/%s/%s/index.html",
      lesson.Product.Category.ID, lesson.Product.ID, lesson.Name)
    }

    // Store lesson
    lessonProduct.Lessons = append(lessonProduct.Lessons, lesson)
    layout.Lessons = append(layout.Lessons, lesson)
  }

  otherColumns := "Others!A2:B"
  otherResp, err := srv.Spreadsheets.Values.Get(spreadsheetId, otherColumns).Do()
  if err != nil {
    return nil, err
  }
  for _, row := range otherResp.Values {
    log.Printf("Parsing - Others: %v\n", row)
    row = deNullifyRow(row)

    layout.Others = append(layout.Others, &Other{
      URL: assert(row[0].(string)),
      SourceGDoc: assert(row[1].(string)),
    })
  }

  return layout, nil
}

// If value is empty a fatal error is thrown, else it returns the string
func assert(value string) string {
  if len(value) == 0 {
    log.Fatal("required string is empty")
  }
  return value
}

// Converts a string of "NULL" to an empty string, else returns the string
func deNullifyRow(columns []interface{}) []interface{} {
  for i := 0; i < len(columns); i++ {
    switch columns[i].(type) {
    case string:
      if strings.ToUpper(columns[i].(string)) == "NULL" {
        columns[i] = ""
      }
    }
  }
  return columns
}
