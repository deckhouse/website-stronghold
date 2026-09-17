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

	srv := web.NewServer(prefix)
	log.Printf("listening on %s (prefix %q)", addr, prefix)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
