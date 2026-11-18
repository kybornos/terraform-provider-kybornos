package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVMLifecycle(t *testing.T) {
	exists := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing bearer token")
			w.WriteHeader(401)
			return
		}
		switch {
		case r.Method == "POST" && r.URL.Path == "/v1/vms/deploy":
			var m map[string]string
			if err := json.NewDecoder(r.Body).Decode(&m); err != nil || m["name"] != "web01" || m["size"] != "small" {
				t.Errorf("wrong deploy payload: %v, %v", m, err)
			}
			exists = true
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"id":"vm-1","name":"web01","size":"small"}`))
		case r.URL.Path == "/v1/vms/vm-1" && !exists:
			http.NotFound(w, r)
		case r.Method == "GET" && r.URL.Path == "/v1/vms/vm-1":
			w.Write([]byte(`{"id":"vm-1","name":"web01","size":"small"}`))
		case r.Method == "DELETE" && r.URL.Path == "/v1/vms/vm-1":
			exists = false
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(400)
		}
	}))
	defer srv.Close()
	c, err := New(srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	created, err := c.Create(ctx, "web01", "small")
	if err != nil || created.ID != "vm-1" {
		t.Fatalf("create: %+v %v", created, err)
	}
	vm, err := c.Get(ctx, created.ID)
	if err != nil || vm.Name != "web01" {
		t.Fatalf("get: %+v %v", vm, err)
	}
	if err := c.Delete(ctx, vm.ID); err != nil {
		t.Fatal(err)
	}
	_, err = c.Get(ctx, vm.ID)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Status != 404 {
		t.Fatalf("expected 404, got %v", err)
	}
}
