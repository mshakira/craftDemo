package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWalkIncs(t *testing.T) {
	// given string, it should send the string to output channel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := []Incident{{"a", "b", "c", "d", "High", "f"}}

	out, err := walkIncs(ctx, obj)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}

	for n := range out {
		if v, ok := n["High"]; ok {
			if v != 1 {
				t.Errorf("Expected 1, got %v\n", v)
			}
		} else {
			t.Errorf("Expected `High` key, but not found")
		}
	}
}

func TestValidateResponse(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	// statuscode non 200
	resp.StatusCode = 500
	err := ValidateResponse(resp)
	if err == nil {
		t.Errorf("Expected 500 error")
	}

	// content type
	resp.StatusCode = 200
	resp.Header["Content-Type"][0] = "text/html; charset=utf-8"
	if err == nil {
		t.Errorf("Expected content-type mismatch error")
	}
}
func Map(m map[string]int) <-chan map[string]int {
	ch := make(chan map[string]int)
	go func() {
		ch <- m
		close(ch)
	}()
	return ch
}

func TestReduceFunc(t *testing.T) {

	m := make(map[string]int)
	m["High"] = 1

	ch := make(chan map[string]int)

	go func() {
		ch <- m
		ch <- m
		close(ch)
	}()

	out := make(chan map[string]int)

	go func() {
		defer close(out)
		mergeIncs(ch,out)
	}()

	in := 0
	for n := range out {
		if v, ok := n["High"]; ok {
			in++
			if v != 2 {
				t.Errorf("Expected 2, got %v\n", v)
			}
		} else {
			t.Errorf("Expected `High` key, but not found")
		}
	}

	if in != 1 {
		t.Errorf("Expected 1, got %v\n", in)
	}

}
