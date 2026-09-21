package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"stronghold-registry-proxy/internal/extract"
	"stronghold-registry-proxy/internal/registry"
)

// DefaultRedirectURL is where GET {prefix}/ sends the user: the download UI
// lives in the Stronghold getting started guide on the documentation site.
const DefaultRedirectURL = "/products/stronghold/gs/linux/step2.html"

// LicenseHeader carries the license key for JSON API requests. The key is
// deliberately not accepted in the query string so it never ends up in
// access logs or browser history.
const LicenseHeader = "X-License-Token"

// normalizePrefix ensures the prefix has a leading slash and no trailing
// slash, or is empty when the service is served from the root.
func normalizePrefix(prefix string) string {
	prefix = strings.TrimSuffix(prefix, "/")
	if prefix == "" {
		return ""
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	return prefix
}

var tagPattern = regexp.MustCompile(`^v[A-Za-z0-9._+-]+$`)

// registryClient is the subset of *registry.Client used by the handlers.
// It exists so that tests can substitute a fake registry.
type registryClient interface {
	ListTags() ([]string, error)
	GetImageLayerDigests(ref string) ([]string, error)
	OpenBlob(digest string) (io.ReadCloser, error)
}

type clientFactory func(license string, edition registry.Edition) registryClient

func defaultClientFactory(license string, edition registry.Edition) registryClient {
	return registry.NewClient(license, edition)
}

type Server struct {
	mux         *http.ServeMux
	prefix      string
	redirectURL string
	newClient   clientFactory
}

func NewServer(prefix, redirectURL string) *Server {
	prefix = normalizePrefix(prefix)
	if redirectURL == "" {
		redirectURL = DefaultRedirectURL
	}
	s := &Server{
		mux:         http.NewServeMux(),
		prefix:      prefix,
		redirectURL: redirectURL,
		newClient:   defaultClientFactory,
	}
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET "+prefix+"/{$}", s.handleIndex)
	s.mux.HandleFunc("GET "+prefix+"/api/versions", s.handleVersions)
	s.mux.HandleFunc("POST "+prefix+"/download/{tag}", s.handleDownload)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleIndex redirects to the getting started page that hosts the download UI.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.redirectURL, http.StatusFound)
}

type versionsResponse struct {
	Edition  registry.Edition `json:"edition"`
	Versions []string         `json:"versions"`
}

// handleVersions returns the list of downloadable Stronghold versions for the
// edition given in the query string, newest first. The license key is taken
// from the X-License-Token header.
func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	license := licenseFromHeader(r)
	if license == "" {
		writeJSONError(w, http.StatusUnauthorized, "license required")
		return
	}
	edition := registry.ParseEdition(r.URL.Query().Get("edition"))

	tags, err := s.newClient(license, edition).ListTags()
	if err != nil {
		if isAuthError(err) {
			writeJSONError(w, http.StatusUnauthorized, "invalid license")
			return
		}
		writeJSONError(w, http.StatusBadGateway, "failed to list versions: "+err.Error())
		return
	}

	versions := filterVersionTags(tags)
	sortVersionsDesc(versions)

	writeJSON(w, http.StatusOK, versionsResponse{Edition: edition, Versions: versions})
}

// handleDownload streams a tar archive with the stronghold binary for the
// requested tag. License and edition come from the POST form body.
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	if !tagPattern.MatchString(tag) {
		http.Error(w, "invalid tag", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	license := strings.TrimSpace(r.PostFormValue("license"))
	if license == "" {
		http.Error(w, "license required", http.StatusUnauthorized)
		return
	}
	edition := registry.ParseEdition(r.PostFormValue("edition"))

	client := s.newClient(license, edition)

	moduleLayers, err := client.GetImageLayerDigests(tag)
	if err != nil {
		if isAuthError(err) {
			http.Error(w, "invalid license", http.StatusUnauthorized)
			return
		}
		http.Error(w, "failed to fetch module image: "+err.Error(), http.StatusBadGateway)
		return
	}

	open := client.OpenBlob
	strongholdDigest, err := extract.FindStrongholdDigest(moduleLayers, open)
	if err != nil {
		http.Error(w, "failed to read images_digests.json: "+err.Error(), http.StatusBadGateway)
		return
	}

	strongholdLayers, err := client.GetImageLayerDigests(strongholdDigest)
	if err != nil {
		http.Error(w, "failed to fetch stronghold image: "+err.Error(), http.StatusBadGateway)
		return
	}

	data, err := extract.FindStrongholdBinary(strongholdLayers, open)
	if err != nil {
		http.Error(w, "failed to extract stronghold binary: "+err.Error(), http.StatusBadGateway)
		return
	}

	filename := fmt.Sprintf("stronghold-%s.tar", tag)
	w.Header().Set("Content-Type", "application/x-tar")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")

	if err := extract.WriteStrongholdTar(w, data); err != nil {
		http.Error(w, "failed to write tar: "+err.Error(), http.StatusBadGateway)
		return
	}
}

func licenseFromHeader(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get(LicenseHeader))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func filterVersionTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		if strings.HasPrefix(tag, "v") && tagPattern.MatchString(tag) {
			out = append(out, tag)
		}
	}
	return out
}

// sortVersionsDesc orders version tags newest first using numeric-aware
// comparison, so that v1.19.3 sorts above v1.9.0.
func sortVersionsDesc(versions []string) {
	sort.SliceStable(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})
}

// compareVersions compares two version tags like "v1.19.3" or
// "v1.19.3-alpha.0". Numeric segments are compared as numbers, other segments
// as strings. A release ("v1.19.3") is considered newer than any pre-release
// of the same version ("v1.19.3-alpha.0").
func compareVersions(a, b string) int {
	aCore, aPre := splitPreRelease(strings.TrimPrefix(a, "v"))
	bCore, bPre := splitPreRelease(strings.TrimPrefix(b, "v"))

	if c := compareSegments(aCore, bCore); c != 0 {
		return c
	}

	switch {
	case aPre == "" && bPre == "":
		return 0
	case aPre == "":
		return 1
	case bPre == "":
		return -1
	}
	return compareSegments(aPre, bPre)
}

func splitPreRelease(v string) (core, pre string) {
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		return v[:i], v[i+1:]
	}
	return v, ""
}

func compareSegments(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		if i >= len(as) {
			return -1
		}
		if i >= len(bs) {
			return 1
		}
		an, aErr := strconv.Atoi(as[i])
		bn, bErr := strconv.Atoi(bs[i])
		switch {
		case aErr == nil && bErr == nil:
			if an != bn {
				if an < bn {
					return -1
				}
				return 1
			}
		case aErr == nil:
			// numeric segments sort above non-numeric ones ("3" > "alpha")
			return 1
		case bErr == nil:
			return -1
		default:
			if c := strings.Compare(as[i], bs[i]); c != 0 {
				return c
			}
		}
	}
	return 0
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unauthorized") || strings.Contains(msg, "auth failed")
}
