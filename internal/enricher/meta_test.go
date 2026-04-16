package enricher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func serveHTML(t *testing.T, body string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
}

func TestFetchMeta_OGDescription(t *testing.T) {
	srv := serveHTML(t, `<html><head>
		<meta property="og:description" content="OG description" />
		<meta name="description" content="Meta description" />
	</head><body></body></html>`)
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Description != "OG description" {
		t.Errorf("expected 'OG description', got %q", meta.Description)
	}
}

func TestFetchMeta_FallbackToMetaDescription(t *testing.T) {
	srv := serveHTML(t, `<html><head>
		<meta name="description" content="Plain meta description" />
		<meta name="twitter:description" content="Twitter description" />
	</head><body></body></html>`)
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Description != "Plain meta description" {
		t.Errorf("expected 'Plain meta description', got %q", meta.Description)
	}
}

func TestFetchMeta_FallbackToTwitterDescription(t *testing.T) {
	srv := serveHTML(t, `<html><head>
		<meta name="twitter:description" content="Twitter description" />
	</head><body></body></html>`)
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Description != "Twitter description" {
		t.Errorf("expected 'Twitter description', got %q", meta.Description)
	}
}

func TestFetchMeta_PriorityOGWinsOverAll(t *testing.T) {
	srv := serveHTML(t, `<html><head>
		<meta name="twitter:description" content="Twitter description" />
		<meta name="description" content="Meta description" />
		<meta property="og:description" content="OG wins" />
	</head><body></body></html>`)
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Description != "OG wins" {
		t.Errorf("expected 'OG wins', got %q", meta.Description)
	}
}

func TestFetchMeta_MissingMetaReturnsEmpty(t *testing.T) {
	srv := serveHTML(t, `<html><head><title>No meta here</title></head><body></body></html>`)
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Description != "" {
		t.Errorf("expected empty description, got %q", meta.Description)
	}
	if meta.Title != "" {
		t.Errorf("expected empty title, got %q", meta.Title)
	}
}

func TestFetchMeta_OGTitle(t *testing.T) {
	srv := serveHTML(t, `<html><head>
		<meta property="og:title" content="Page Title" />
		<meta property="og:description" content="Page desc" />
	</head><body></body></html>`)
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Title != "Page Title" {
		t.Errorf("expected 'Page Title', got %q", meta.Title)
	}
}

func TestFetchMeta_UnreachableURLReturnsEmpty(t *testing.T) {
	// Use a port on localhost that is not listening.
	meta := FetchMeta("http://127.0.0.1:19999/no-such-server")
	if meta.Description != "" || meta.Title != "" {
		t.Errorf("expected empty Meta for unreachable URL, got %+v", meta)
	}
}

func TestFetchMeta_TimeoutReturnsEmpty(t *testing.T) {
	// Server that never responds — the 5 s client timeout will fire first, but
	// we override the package-level client with a very short timeout for the
	// test so it completes quickly.
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block // block forever
	}))
	defer srv.Close()
	defer close(block)

	orig := httpClient
	httpClient = &http.Client{Timeout: 1} // 1 nanosecond — effectively immediate
	defer func() { httpClient = orig }()

	meta := FetchMeta(srv.URL)
	if meta.Description != "" || meta.Title != "" {
		t.Errorf("expected empty Meta on timeout, got %+v", meta)
	}
}

func TestFetchMeta_Non200ReturnsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	meta := FetchMeta(srv.URL)
	if meta.Description != "" || meta.Title != "" {
		t.Errorf("expected empty Meta for 404, got %+v", meta)
	}
}
