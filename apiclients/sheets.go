package apiclients

import (
  "golang.org/x/net/context"
  "google.golang.org/api/sheets/v4"
)

func init() {
  scopes = append(scopes, sheets.SpreadsheetsReadonlyScope);
}

func NewSheetsClient(clientSecretPath string) (*sheets.Service, error) {
  ctx := context.Background()

  client, err := NewClient(ctx, clientSecretPath)
  if err != nil {
    return nil, err
  }

  return sheets.New(client)
}
