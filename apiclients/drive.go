package apiclients
import (
  "net/http"
  "golang.org/x/net/context"
  "google.golang.org/api/drive/v3"
)

func init() {
  scopes = append(scopes, drive.DriveReadonlyScope);
}

func NewDriveClient(clientSecretPath string) (*drive.Service, error) {
  ctx := context.Background()

  client, err := NewClient(ctx, clientSecretPath)
  if err != nil {
    return nil, err
  }

  return drive.New(client)
}

func NewDriveFilesService(clientSecretPath string) (*drive.FilesService, error) {
  driveClient, err := NewDriveClient(clientSecretPath)
  if err != nil {
    return nil, err
  }

  filesService := drive.NewFilesService(driveClient)
  return filesService, nil
}

func GetGdocHtml(clientSecretPath string, gdocID string) (*http.Response, error) {
  client, err := NewDriveFilesService(clientSecretPath)
  if err != nil {
    return nil, err
  }

  resp, err := client.Export(gdocID, "text/html").Download()
  if err != nil {
    return nil, err
  }

  return resp, nil
}
