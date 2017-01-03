package apiclients

import (
  "fmt"
  "io/ioutil"
  "testing"
)


func TestNewDriveFilesService(t *testing.T) {
  resp, err := GetGdocHtml("../client_secret.json", "1AfHN84meZVkywhN59bNJsvCD1FbtO3Srbh5tJmG_tX8")
  if err != nil {
    t.Fatal(err)
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    t.Fatal(err)
  }
  fmt.Println(string(body))

  if len(string(body)) == 0 {
      t.Fatal("No body from gdoc")
  }
}
