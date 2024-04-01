package frontegg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/clients"
)

func doRequest(ctx context.Context, client *clients.FronteggClient, method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Token)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var sb strings.Builder
		_, err = io.Copy(&sb, resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("HTTP request error: status %d, response: %s", resp.StatusCode, sb.String())
	}

	return resp, nil
}

func jsonEncode(payload interface{}) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(payload); err != nil {
		return nil, err
	}
	return buffer, nil
}
