package layout

import (
  "testing"
)


func TestGetLayout(t *testing.T) {
  layout, err := GetLayout("../client_secret.json", "1-Nj5UkRGfD-9N6zj3B7mYXrJFvOgxmzm3RXv2cLeAh4")
  if err != nil {
    t.Fatal(err)
  }

  if len(layout.Categories) == 0 {
    t.Fatal("No categories")
  }

  if len(layout.Products) == 0 {
    t.Fatal("No products")
  }

  if len(layout.Lessons) == 0 {
    t.Fatal("No Lessons")
  }

  if len(layout.Others) == 0 {
    t.Fatal("No Others")
  }
}
