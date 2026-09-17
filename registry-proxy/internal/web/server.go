package web

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"stronghold-registry-proxy/internal/extract"
	"stronghold-registry-proxy/internal/registry"
)

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

const (
	cookieLicense = "license"
	cookieEdition = "edition"
)

var tagPattern = regexp.MustCompile(`^v[A-Za-z0-9._+-]+$`)

type Server struct {
	mux    *http.ServeMux
	prefix string
}

func NewServer(prefix string) *Server {
	prefix = normalizePrefix(prefix)
	s := &Server{mux: http.NewServeMux(), prefix: prefix}
	s.mux.HandleFunc("GET "+prefix+"/", s.handleIndex)
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("POST "+prefix+"/login", s.handleLogin)
	s.mux.HandleFunc("POST "+prefix+"/logout", s.handleLogout)
	s.mux.HandleFunc("GET "+prefix+"/download/{tag}", s.handleDownload)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	edition := editionFromRequest(r)
	license := licenseFromRequest(r)
	s.renderHome(w, license, edition, "")
}

func (s *Server) renderHome(w http.ResponseWriter, license string, edition registry.Edition, loginErr string) {
	if license == "" {
		s.renderLogin(w, loginErr, edition)
		return
	}

	client := registry.NewClient(license, edition)
	tags, err := client.ListTags()
	if err != nil {
		if isAuthError(err) {
			clearLicenseCookie(w)
			s.renderLogin(w, "Invalid license code", edition)
			return
		}
		http.Error(w, "failed to list tags: "+err.Error(), http.StatusBadGateway)
		return
	}

	versionTags := filterVersionTags(tags)
	sort.Slice(versionTags, func(i, j int) bool {
		return versionTags[i] > versionTags[j]
	})

	s.renderTags(w, versionTags, edition)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	edition := registry.ParseEdition(r.FormValue("edition"))
	setEditionCookie(w, edition)

	license := strings.TrimSpace(r.FormValue("license"))
	if license == "" {
		s.renderLogin(w, "License code is required", edition)
		return
	}

	client := registry.NewClient(license, edition)
	if _, err := client.ListTags(); err != nil {
		if isAuthError(err) {
			w.WriteHeader(http.StatusUnauthorized)
			s.renderLogin(w, "Invalid license code", edition)
			return
		}
		http.Error(w, "failed to verify license: "+err.Error(), http.StatusBadGateway)
		return
	}

	setLicenseCookie(w, license)
	s.renderHome(w, license, edition, "")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	clearLicenseCookie(w)
	s.renderHome(w, "", editionFromRequest(r), "")
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	if !tagPattern.MatchString(tag) {
		http.Error(w, "invalid tag", http.StatusBadRequest)
		return
	}

	license := licenseFromRequest(r)
	if license == "" {
		http.Error(w, "license required", http.StatusUnauthorized)
		return
	}

	edition := editionFromRequest(r)
	client := registry.NewClient(license, edition)

	moduleLayers, err := client.GetImageLayerDigests(tag)
	if err != nil {
		if isAuthError(err) {
			clearLicenseCookie(w)
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

	if err := extract.WriteStrongholdTar(w, data); err != nil {
		http.Error(w, "failed to write tar: "+err.Error(), http.StatusBadGateway)
		return
	}
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

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unauthorized") || strings.Contains(msg, "auth failed")
}

func licenseFromRequest(r *http.Request) string {
	c, err := r.Cookie(cookieLicense)
	if err != nil {
		return ""
	}
	return c.Value
}

func editionFromRequest(r *http.Request) registry.Edition {
	c, err := r.Cookie(cookieEdition)
	if err != nil {
		return registry.EditionEE
	}
	return registry.ParseEdition(c.Value)
}

func setLicenseCookie(w http.ResponseWriter, license string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieLicense,
		Value:    license,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})
}

func setEditionCookie(w http.ResponseWriter, edition registry.Edition) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieEdition,
		Value:    string(edition),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})
}

func clearLicenseCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieLicense,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

var pageTemplate = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Stronghold Download</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Onest:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>
    :root {
      --brand: #0a6eff;
      --brand-dark: #004df2;
      --bg: #f9f9fd;
      --surface: #fff;
      --border: #e8e9f3;
      --text: #1c1b1f;
      --text-muted: #6e7084;
      --danger: #c92230;
    }
    * { box-sizing: border-box; }
    body {
      font-family: Onest, system-ui, sans-serif;
      background: var(--bg);
      color: var(--text);
      max-width: 640px;
      margin: 0 auto;
      padding: 3rem 1.5rem;
      line-height: 1.5;
    }
    h1 { font-size: 1.75rem; font-weight: 700; margin: 0 0 1.5rem; }
    .card {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: 16px;
      padding: 1.75rem;
    }
    .error {
      color: var(--danger);
      background: #fff2f0;
      border-radius: 8px;
      padding: 0.6rem 0.9rem;
      margin-bottom: 1rem;
      font-size: 0.9rem;
    }
    form { margin: 0; }
    label { font-size: 0.9rem; font-weight: 500; }
    input[type=text], input[type=password] {
      width: 100%;
      max-width: 100%;
      padding: 0.65rem 0.8rem;
      margin-top: 0.4rem;
      border: 1px solid var(--border);
      border-radius: 8px;
      font-family: inherit;
      font-size: 1rem;
      background: var(--bg);
      color: var(--text);
    }
    input[type=text]:focus, input[type=password]:focus { outline: 2px solid var(--brand); outline-offset: 1px; }
    button {
      margin-top: 1.25rem;
      padding: 0.65rem 1.4rem;
      border: none;
      border-radius: 999px;
      background: var(--brand);
      color: #fff;
      font-family: inherit;
      font-size: 1rem;
      font-weight: 600;
      cursor: pointer;
    }
    button:hover { background: var(--brand-dark); }
    ul { list-style: none; padding: 0; margin: 1rem 0 0; }
    li { margin: 0.5rem 0; }
    a.tag-link {
      display: inline-block;
      width: 100%;
      padding: 0.65rem 0.9rem;
      background: var(--bg);
      border: 1px solid var(--border);
      border-radius: 8px;
      text-decoration: none;
      color: var(--brand-dark);
      font-weight: 600;
    }
    a.tag-link:hover { background: #eef3ff; border-color: var(--brand); }
    .logout { margin-top: 1.5rem; }
    .logout button { background: transparent; color: var(--text-muted); border: 1px solid var(--border); }
    .logout button:hover { background: var(--bg); color: var(--text); }
    .edition-label { display: block; margin-bottom: 0.5rem; font-weight: 500; }
    .edition-toggle { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-bottom: 1.25rem; }
    .edition-toggle input { position: absolute; opacity: 0; pointer-events: none; }
    .edition-toggle label {
      display: inline-block; padding: 0.5rem 0.9rem; border: 1px solid var(--border);
      border-radius: 999px; cursor: pointer; background: var(--bg); color: var(--text-muted);
      font-size: 0.9rem; font-weight: 500;
      user-select: none;
    }
    .edition-toggle input:checked + label {
      background: var(--brand); border-color: var(--brand); color: #fff;
    }
    .edition { margin: 0 0 1rem; color: var(--text-muted); }
  </style>
</head>
<body>
  <h1>Stronghold Download</h1>
  <div class="card">
  {{if .ShowLogin}}
    {{if .Error}}<p class="error">{{.Error}}</p>{{end}}
    <form method="post" action="{{.Prefix}}/login">
      <span class="edition-label">Select edition:</span>
      <div class="edition-toggle" role="group" aria-label="Select edition">
        <input type="radio" id="edition-ee" name="edition" value="ee" {{if .EditionEE}}checked{{end}}>
        <label for="edition-ee">Enterprise Edition (EE)</label>
        <input type="radio" id="edition-cse" name="edition" value="cse" {{if .EditionCSE}}checked{{end}}>
        <label for="edition-cse">Certified Security Edition (CSE)</label>
      </div>
      <label for="license">License code</label><br>
      <input id="license" name="license" type="password" autocomplete="new-password" required>
      <br>
      <button type="submit">Continue</button>
    </form>
  {{else}}
    <p class="edition">Edition: {{.EditionLabel}}</p>
    <p>Select a version to download:</p>
    <ul>
      {{range .Tags}}
      <li><a class="tag-link" href="{{$.Prefix}}/download/{{.}}">{{.}}</a></li>
      {{else}}
      <li>No version tags found.</li>
      {{end}}
    </ul>
    <form class="logout" method="post" action="{{.Prefix}}/logout">
      <button type="submit">Change license</button>
    </form>
  {{end}}
  </div>
</body>
</html>`))

type pageData struct {
	Prefix       string
	ShowLogin    bool
	Error        string
	Tags         []string
	EditionEE    bool
	EditionCSE   bool
	EditionLabel string
}

func (s *Server) renderLogin(w http.ResponseWriter, errMsg string, edition registry.Edition) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(w, pageData{
		Prefix:     s.prefix,
		ShowLogin:  true,
		Error:      errMsg,
		EditionEE:  edition != registry.EditionCSE,
		EditionCSE: edition == registry.EditionCSE,
	})
}

func (s *Server) renderTags(w http.ResponseWriter, tags []string, edition registry.Edition) {
	label := "Enterprise Edition (EE)"
	if edition == registry.EditionCSE {
		label = "Certified Security Edition (CSE)"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(w, pageData{
		Prefix:       s.prefix,
		Tags:         tags,
		EditionLabel: label,
	})
}
