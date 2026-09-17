package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stronghold-registry-proxy/internal/registry"
)

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

func TestEditionFromRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if editionFromRequest(r) != registry.EditionEE {
		t.Fatal("default should be ee")
	}
	r.AddCookie(&http.Cookie{Name: cookieEdition, Value: "cse"})
	if editionFromRequest(r) != registry.EditionCSE {
		t.Fatal("expected cse from cookie")
	}
}
