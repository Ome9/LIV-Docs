package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleIndex(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleIndex)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "LIV Document Viewer") {
		t.Errorf("handler returned unexpected body: missing title")
	}

	if !strings.Contains(body, "Progressive Web App") {
		t.Errorf("handler returned unexpected body: missing PWA features")
	}

	if !strings.Contains(body, "manifest.json") {
		t.Errorf("handler returned unexpected body: missing PWA manifest")
	}
}

func TestHandleViewer(t *testing.T) {
	req, err := http.NewRequest("GET", "/viewer?id=test123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleViewer)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "LIV Viewer") {
		t.Errorf("handler returned unexpected body: missing viewer title")
	}

	if !strings.Contains(body, "WASM") {
		t.Errorf("handler returned unexpected body: missing WASM integration")
	}

	if !strings.Contains(body, "responsive") {
		t.Errorf("handler returned unexpected body: missing responsive design")
	}
}

func TestHandleDocument(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/document?id=test123", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleDocument)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "test123") {
		t.Errorf("handler returned unexpected body: missing document ID")
	}

	if !strings.Contains(body, "Sample LIV Document") {
		t.Errorf("handler returned unexpected body: missing document title")
	}
}

func TestHandleManifest(t *testing.T) {
	req, err := http.NewRequest("GET", "/manifest.json", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleManifest)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/manifest+json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/manifest+json")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "LIV Viewer") {
		t.Errorf("handler returned unexpected body: missing app name")
	}

	if !strings.Contains(body, "standalone") {
		t.Errorf("handler returned unexpected body: missing display mode")
	}
}

func TestHandleServiceWorker(t *testing.T) {
	req, err := http.NewRequest("GET", "/sw.js", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handleServiceWorker)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/javascript" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/javascript")
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Service Worker") {
		t.Errorf("handler returned unexpected body: missing service worker")
	}

	if !strings.Contains(body, "cache") {
		t.Errorf("handler returned unexpected body: missing caching functionality")
	}
}

func TestHandleStatic(t *testing.T) {
	tests := []struct {
		path        string
		contentType string
		shouldExist bool
	}{
		{"/static/wasm/interactive-engine.wasm", "application/wasm", true},
		{"/static/js/app.js", "application/javascript", true},
		{"/static/css/app.css", "text/css", true},
		{"/static/icons/icon-192x192.png", "image/png", true},
		{"/static/nonexistent.txt", "", false},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", tt.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handleStatic)

		handler.ServeHTTP(rr, req)

		if tt.shouldExist {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code for %s: got %v want %v",
					tt.path, status, http.StatusOK)
			}

			if contentType := rr.Header().Get("Content-Type"); contentType != tt.contentType {
				t.Errorf("handler returned wrong content type for %s: got %v want %v",
					tt.path, contentType, tt.contentType)
			}
		} else {
			if status := rr.Code; status != http.StatusNotFound {
				t.Errorf("handler returned wrong status code for %s: got %v want %v",
					tt.path, status, http.StatusNotFound)
			}
		}
	}
}

func TestRunViewer(t *testing.T) {
	// Test that the viewer function handles different modes correctly
	tests := []struct {
		web      bool
		fallback bool
		debug    bool
	}{
		{true, false, false},
		{true, true, false},
		{true, false, true},
		{false, false, false}, // This should return an error for desktop mode
	}

	for _, tt := range tests {
		// We can't actually run the server in tests, but we can test the logic
		// The desktop viewer should return an error since it's not implemented
		if !tt.web {
			err := runDesktopViewer("", tt.fallback, tt.debug)
			if err == nil {
				t.Errorf("expected error for desktop viewer, got nil")
			}
		}
	}
}