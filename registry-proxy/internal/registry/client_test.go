package registry

import "testing"

func TestParseWwwAuthenticate(t *testing.T) {
	header := `Bearer realm="https://registry.deckhouse.ru/auth",service="Docker registry"`
	got := parseWwwAuthenticate(header)
	if got["realm"] != "https://registry.deckhouse.ru/auth" {
		t.Fatalf("realm: %q", got["realm"])
	}
	if got["service"] != "Docker registry" {
		t.Fatalf("service: %q", got["service"])
	}
}

func TestEndpointForDefaults(t *testing.T) {
	host, repo := EndpointFor(EditionEE)
	if host != "registry.deckhouse.ru" || repo != "deckhouse/fe/modules/stronghold" {
		t.Fatalf("EE: %s / %s", host, repo)
	}
	host, repo = EndpointFor(EditionCSE)
	if host != "registry-cse.deckhouse.ru" || repo != "stronghold/cse/modules/stronghold" {
		t.Fatalf("CSE: %s / %s", host, repo)
	}
}

func TestParseEdition(t *testing.T) {
	if ParseEdition("cse") != EditionCSE {
		t.Fatal("expected cse")
	}
	if ParseEdition("") != EditionEE {
		t.Fatal("expected ee default")
	}
}
