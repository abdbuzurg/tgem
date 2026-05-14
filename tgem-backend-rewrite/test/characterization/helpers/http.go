package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// Envelope mirrors internal/http/response.ResponseFormat verbatim. Every endpoint in this
// app returns HTTP 200 with one of these. Decode the payload from Data using
// json.Unmarshal yourself.
type Envelope struct {
	Data       json.RawMessage `json:"data"`
	Error      string          `json:"error"`
	Success    bool            `json:"success"`
	Permission bool            `json:"permission"`
}

// AuthedJSON sends a JSON request with Authorization: Bearer <token>, decodes
// the standard envelope, and asserts the HTTP status was 200.
func AuthedJSON(t *testing.T, method, path, token string, body any) Envelope {
	t.Helper()
	return doJSON(t, method, path, token, body)
}

// RawJSON is AuthedJSON without the Authorization header.
func RawJSON(t *testing.T, method, path string, body any) Envelope {
	t.Helper()
	return RequestJSON(t, method, path, nil, body)
}

// RequestJSON sends a JSON request with an arbitrary header map. Use this when
// you need to test middleware behavior with non-standard Authorization values
// (e.g. `Basic xxx`, malformed bearer tokens). When body is nil no
// Content-Type is set; otherwise application/json.
func RequestJSON(t *testing.T, method, path string, headers map[string]string, body any) Envelope {
	t.Helper()

	var buf io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		buf = bytes.NewReader(raw)
	}

	url := BaseURL() + path
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 from %s %s, got %d: %s", method, url, resp.StatusCode, string(raw))
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode envelope from %s %s: %v\nbody: %s", method, url, err, string(raw))
	}
	return env
}

func doJSON(t *testing.T, method, path, token string, body any) Envelope {
	t.Helper()
	var headers map[string]string
	if token != "" {
		headers = map[string]string{"Authorization": "Bearer " + token}
	}
	return RequestJSON(t, method, path, headers, body)
}

// MultipartUpload posts a single-file multipart form. Used for Excel imports.
func MultipartUpload(t *testing.T, path, token, fieldName, filePath string, extra map[string]string) Envelope {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range extra {
		if err := writer.WriteField(k, v); err != nil {
			t.Fatalf("write field %q: %v", k, err)
		}
	}

	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("open %s: %v", filePath, err)
	}
	defer f.Close()

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		t.Fatalf("copy file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	url := BaseURL() + path
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do POST %s: %v", url, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", resp.StatusCode, string(raw))
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode envelope: %v\nbody: %s", err, string(raw))
	}
	return env
}

// Download fetches a binary payload (e.g. xlsx export). Returns body bytes and
// content-type.
func Download(t *testing.T, path, token string) ([]byte, string) {
	t.Helper()
	url := BaseURL() + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", resp.StatusCode, string(raw))
	}
	return raw, resp.Header.Get("Content-Type")
}

// MustDecode is a convenience for unmarshalling Envelope.Data into a typed
// destination, reporting via t.Fatalf on failure.
func MustDecode(t *testing.T, env Envelope, dst any) {
	t.Helper()
	if !env.Success {
		t.Fatalf("envelope not successful: error=%q", env.Error)
	}
	if err := json.Unmarshal(env.Data, dst); err != nil {
		t.Fatalf("decode envelope.data into %T: %v\ndata: %s", dst, err, string(env.Data))
	}
}

// AssertSuccess fatals when env.Success is false.
func AssertSuccess(t *testing.T, env Envelope, context string) {
	t.Helper()
	if !env.Success {
		t.Fatalf("%s: envelope not successful: error=%q", context, env.Error)
	}
}

// AssertFailure fatals when env.Success is true; returns the error message.
func AssertFailure(t *testing.T, env Envelope, context string) string {
	t.Helper()
	if env.Success {
		t.Fatalf("%s: expected failure, got success: data=%s", context, string(env.Data))
	}
	return env.Error
}

