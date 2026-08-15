package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetNodeTrimsName is a regression test: AddNode and RemoveNode trim
// surrounding whitespace from node names, but GetNode did not. A node added
// as "  mynode  " is stored as "mynode"; querying it with the same spaced
// name must find it, just as adding and removing with the spaced name do.
func TestGetNodeTrimsName(t *testing.T) {
	s := NewService()

	// AddNode trims, so the node is stored under "mynode".
	if _, err := s.AddNode("  mynode  ", 0); err != nil {
		t.Fatalf("AddNode with spaced name: %v", err)
	}

	// GetNode must trim too; otherwise the spaced lookup misses "mynode".
	got, err := s.GetNode("  mynode  ")
	if err != nil {
		t.Fatalf("GetNode(%q) returned error: %v (AddNode trims names, so GetNode should trim too)", "  mynode  ", err)
	}
	if got.Name != "mynode" {
		t.Fatalf("GetNode name = %q, want %q", got.Name, "mynode")
	}

	// Sanity: the trimmed name still resolves.
	if _, err := s.GetNode("mynode"); err != nil {
		t.Fatalf("GetNode(%q) returned error: %v", "mynode", err)
	}
}

// TestGetNodeTrimsNameHTTP reproduces the same bug end-to-end through the
// HTTP API: GET /nodes/{name} with a URL-encoded spaced name must return the
// node (200) rather than 404.
func TestGetNodeTrimsNameHTTP(t *testing.T) {
	srv := NewService()
	ts := httptest.NewServer(buildMux(srv))
	defer ts.Close()

	// Add via the API with surrounding spaces; stored as "mynode".
	resp, err := http.Post(ts.URL+"/nodes", "application/json", bytes.NewBufferString(`{"name":"  mynode  "}`))
	if err != nil {
		t.Fatalf("add node: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("add status = %d, want 201", resp.StatusCode)
	}

	// Query with the spaced name URL-encoded; before the fix this returned 404.
	resp, err = http.Get(ts.URL + "/nodes/%20%20mynode%20%20")
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want 200 (GetNode must trim the name like AddNode does)", resp.StatusCode)
	}
}
