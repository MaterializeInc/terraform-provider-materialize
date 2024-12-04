package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

func doRequest(ctx context.Context, client *clients.FronteggClient, method, endpoint string, body io.Reader) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	return clients.FronteggRequest(ctx, client, method, endpoint, bodyBytes)
}

func jsonEncode(payload interface{}) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(payload); err != nil {
		return nil, err
	}
	return buffer, nil
}
