package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"stronghold-registry-proxy/internal/registry"
)

const testPrefix = "/products/stronghold/get"

type fakeClient struct {
	tags    []string
	tagsErr error

	license string
	edition registry.Edition
}

func (f *fakeClient) ListTags() ([]string, error) { return f.tags, f.tagsErr }
func (f *fakeClient) GetImageLayerDigests(ref string) ([]string, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeClient) OpenBlob(digest string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func newTestServer(t *testing.T, fake *fakeClient) *Server {
	t.Helper()
	s := NewServer(testPrefix, "")
	s.newClient = func(license string, edition registry.Edition) registryClient {
		fake.license = license
		fake.edition = edition
		return fake
	}
	return s
}

func postForm(path string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestFilterVersionTags(t *testing.T) {
	in := []string{"v1.19.3", "meta-foo", "sha256-abc.att", "v1.15.2-alpha.0", "000abc"}
	got := filterVersionTags(in)
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
	if got[0] != "v1.19.3" || got[1] != "v1.15.2-alpha.0" {
		t.Fatalf("unexpected tags: %v", got)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.19.3", "v1.9.0", 1},
		{"v1.9.0", "v1.19.3", -1},
		{"v1.19.3", "v1.19.3", 0},
		{"v1.19.3", "v1.19.3-alpha.0", 1},
		{"v1.19.3-alpha.0", "v1.19.3", -1},
		{"v1.19.3-alpha.1", "v1.19.3-alpha.0", 1},
		{"v1.19.3-beta.0", "v1.19.3-alpha.5", 1},
		{"v1.20", "v1.19.9", 1},
		{"v2.0.0", "v1.99.99", 1},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestNormalizePrefix(t *testing.T) {
	cases := map[string]string{
		"":                          "",
		"/":                         "",
		"get":                       "/get",
		"/products/stronghold/get/": "/products/stronghold/get",
		"/products/stronghold/get":  "/products/stronghold/get",
	}
	for in, want := range cases {
		if got := normalizePrefix(in); got != want {
			t.Errorf("normalizePrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIndexRedirects(t *testing.T) {
	s := newTestServer(t, &fakeClient{})
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testPrefix+"/", nil))

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != DefaultRedirectURL {
		t.Fatalf("Location = %q, want %q", loc, DefaultRedirectURL)
	}
}

func TestIndexRedirectsToCustomURL(t *testing.T) {
	s := NewServer(testPrefix, "/custom/page.html")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testPrefix+"/", nil))
	if loc := rec.Header().Get("Location"); loc != "/custom/page.html" {
		t.Fatalf("Location = %q", loc)
	}
}

func TestVersionsRequiresLicense(t *testing.T) {
	s := newTestServer(t, &fakeClient{tags: []string{"v1.0.0"}})
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testPrefix+"/api/versions?edition=ee", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["error"] != "license required" {
		t.Fatalf("error = %q", body["error"])
	}
}

func TestVersionsListsSortedTags(t *testing.T) {
	fake := &fakeClient{tags: []string{"v1.9.0", "meta-foo", "v1.19.3", "v1.19.3-alpha.0", "sha256-abc.att"}}
	s := newTestServer(t, fake)

	req := httptest.NewRequest(http.MethodGet, testPrefix+"/api/versions?edition=cse", nil)
	req.Header.Set(LicenseHeader, "  secret-key  ")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q", cc)
	}
	if fake.license != "secret-key" {
		t.Fatalf("license passed to client = %q", fake.license)
	}
	if fake.edition != registry.EditionCSE {
		t.Fatalf("edition passed to client = %q", fake.edition)
	}

	var resp versionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.Edition != registry.EditionCSE {
		t.Fatalf("edition = %q", resp.Edition)
	}
	want := []string{"v1.19.3", "v1.19.3-alpha.0", "v1.9.0"}
	if !reflect.DeepEqual(resp.Versions, want) {
		t.Fatalf("versions = %v, want %v", resp.Versions, want)
	}
}

func TestVersionsDefaultsToEE(t *testing.T) {
	fake := &fakeClient{tags: []string{"v1.0.0"}}
	s := newTestServer(t, fake)

	req := httptest.NewRequest(http.MethodGet, testPrefix+"/api/versions", nil)
	req.Header.Set(LicenseHeader, "key")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if fake.edition != registry.EditionEE {
		t.Fatalf("edition = %q, want ee", fake.edition)
	}
}

func TestVersionsInvalidLicense(t *testing.T) {
	s := newTestServer(t, &fakeClient{tagsErr: errors.New("auth failed: {\"errors\":[...]}")})

	req := httptest.NewRequest(http.MethodGet, testPrefix+"/api/versions", nil)
	req.Header.Set(LicenseHeader, "bad")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "invalid license") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestVersionsRegistryFailure(t *testing.T) {
	s := newTestServer(t, &fakeClient{tagsErr: errors.New("registry GET /v2/...: boom")})

	req := httptest.NewRequest(http.MethodGet, testPrefix+"/api/versions", nil)
	req.Header.Set(LicenseHeader, "key")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestDownloadRejectsGet(t *testing.T) {
	s := newTestServer(t, &fakeClient{})
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, testPrefix+"/download/v1.0.0?license=x", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestDownloadBadTag(t *testing.T) {
	s := newTestServer(t, &fakeClient{})
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, postForm(testPrefix+"/download/latest", url.Values{"license": {"key"}, "edition": {"ee"}}))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestDownloadRequiresLicense(t *testing.T) {
	s := newTestServer(t, &fakeClient{})
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, postForm(testPrefix+"/download/v1.0.0", url.Values{"edition": {"ee"}}))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestDownloadPassesFormValuesToClient(t *testing.T) {
	fake := &fakeClient{}
	s := newTestServer(t, fake)
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, postForm(testPrefix+"/download/v1.0.0", url.Values{"license": {"form-key"}, "edition": {"cse"}}))

	// The fake cannot fetch manifests, so the handler fails downstream; we only
	// verify that credentials from the form reached the client.
	if fake.license != "form-key" || fake.edition != registry.EditionCSE {
		t.Fatalf("client got license=%q edition=%q", fake.license, fake.edition)
	}
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
