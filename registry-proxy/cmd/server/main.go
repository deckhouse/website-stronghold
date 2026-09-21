package main

import (
	"log"
	"net/http"
	"os"

	"stronghold-registry-proxy/internal/web"
)

func main() {
	addr := ":8080"
	if v := os.Getenv("LISTEN_ADDR"); v != "" {
		addr = v
	}

	prefix := os.Getenv("PATH_PREFIX")

	// Where GET {prefix}/ redirects: the download UI lives in the getting
	// started guide on the documentation site.
	redirectURL := os.Getenv("GS_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = web.DefaultRedirectURL
	}

	srv := web.NewServer(prefix, redirectURL)
	log.Printf("listening on %s (prefix %q, redirect %q)", addr, prefix, redirectURL)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
