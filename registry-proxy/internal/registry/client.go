package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const Username = "license-token"

type Edition string

const (
	EditionEE  Edition = "ee"
	EditionCSE Edition = "cse"
)

func ParseEdition(s string) Edition {
	if strings.EqualFold(s, string(EditionCSE)) {
		return EditionCSE
	}
	return EditionEE
}

func EndpointFor(edition Edition) (host, repository string) {
	switch edition {
	case EditionCSE:
		return envOr("REGISTRY_CSE_HOST", "registry-cse.deckhouse.ru"),
			envOr("REGISTRY_CSE_REPOSITORY", "stronghold/cse/modules/stronghold")
	default:
		return envOr("REGISTRY_HOST", "registry.deckhouse.ru"),
			envOr("REGISTRY_REPOSITORY", "deckhouse/fe/modules/stronghold")
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type Client struct {
	license    string
	host       string
	repository string
	http       *http.Client

	mu    sync.Mutex
	token string
}

func NewClient(license string, edition Edition) *Client {
	host, repo := EndpointFor(edition)
	return &Client{
		license:    license,
		host:       host,
		repository: repo,
		http:       http.DefaultClient,
	}
}

type tagsResponse struct {
	Tags []string `json:"tags"`
}

type manifestV2 struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		Digest string `json:"digest"`
	} `json:"config"`
	Layers []struct {
		Digest string `json:"digest"`
		Size   int64  `json:"size"`
	} `json:"layers"`
}

type manifestList struct {
	Manifests []struct {
		Digest   string `json:"digest"`
		Platform struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
		} `json:"platform"`
	} `json:"manifests"`
}

type tokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

func (c *Client) ListTags() ([]string, error) {
	body, err := c.doRegistry("GET", "/v2/"+c.repository+"/tags/list?n=1000", nil)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var resp tagsResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode tags: %w", err)
	}
	return resp.Tags, nil
}

func (c *Client) GetImageLayerDigests(ref string) ([]string, error) {
	manifest, err := c.fetchManifest(ref)
	if err != nil {
		return nil, err
	}
	digests := make([]string, 0, len(manifest.Layers))
	for _, layer := range manifest.Layers {
		digests = append(digests, layer.Digest)
	}
	return digests, nil
}

func (c *Client) OpenBlob(digest string) (io.ReadCloser, error) {
	return c.doRegistry("GET", "/v2/"+c.repository+"/blobs/"+digest, nil)
}

func (c *Client) fetchManifest(ref string) (*manifestV2, error) {
	body, err := c.doRegistry("GET", "/v2/"+c.repository+"/manifests/"+ref, map[string]string{
		"Accept": "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.index.v1+json",
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()

	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var list manifestList
	if err := json.Unmarshal(raw, &list); err == nil && len(list.Manifests) > 0 {
		digest, err := pickPlatformDigest(&list)
		if err != nil {
			return nil, err
		}
		return c.fetchManifest(digest)
	}

	var manifest manifestV2
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	if len(manifest.Layers) == 0 {
		return nil, fmt.Errorf("manifest has no layers")
	}
	return &manifest, nil
}

func pickPlatformDigest(list *manifestList) (string, error) {
	for _, m := range list.Manifests {
		if m.Platform.OS == "linux" && m.Platform.Architecture == "amd64" {
			return m.Digest, nil
		}
	}
	if len(list.Manifests) > 0 {
		return list.Manifests[0].Digest, nil
	}
	return "", fmt.Errorf("manifest list is empty")
}

func (c *Client) doRegistry(method, path string, headers map[string]string) (io.ReadCloser, error) {
	reqURL := "https://" + c.host + path
	for attempt := 0; attempt < 2; attempt++ {
		token, err := c.getToken()
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest(method, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}

		switch resp.StatusCode {
		case http.StatusOK:
			return resp.Body, nil
		case http.StatusUnauthorized:
			resp.Body.Close()
			c.invalidateToken()
			continue
		default:
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("registry %s %s: %s", method, path, strings.TrimSpace(string(b)))
		}
	}
	return nil, fmt.Errorf("registry unauthorized after retry")
}

func (c *Client) getToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" {
		return c.token, nil
	}

	challenge, err := c.fetchAuthChallenge()
	if err != nil {
		return "", err
	}

	realm := challenge["realm"]
	service := challenge["service"]
	if realm == "" {
		return "", fmt.Errorf("missing auth realm")
	}

	scope := "repository:" + c.repository + ":pull"
	u, err := url.Parse(realm)
	if err != nil {
		return "", fmt.Errorf("parse auth realm: %w", err)
	}
	q := u.Query()
	q.Set("service", service)
	q.Set("scope", scope)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+basicAuth(Username, c.license))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth failed: %s", strings.TrimSpace(string(b)))
	}

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	token := tr.Token
	if token == "" {
		token = tr.AccessToken
	}
	if token == "" {
		return "", fmt.Errorf("empty token from registry")
	}
	c.token = token
	return token, nil
}

func (c *Client) invalidateToken() {
	c.mu.Lock()
	c.token = ""
	c.mu.Unlock()
}

func (c *Client) fetchAuthChallenge() (map[string]string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://"+c.host+"/v2/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusUnauthorized {
		return nil, fmt.Errorf("expected 401 from /v2/, got %d", resp.StatusCode)
	}
	return parseWwwAuthenticate(resp.Header.Get("Www-Authenticate")), nil
}

func parseWwwAuthenticate(header string) map[string]string {
	result := make(map[string]string)
	if header == "" {
		return result
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return result
	}
	for _, item := range strings.Split(parts[1], ",") {
		item = strings.TrimSpace(item)
		kv := strings.SplitN(item, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		result[key] = val
	}
	return result
}

func basicAuth(user, pass string) string {
	return base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
}
