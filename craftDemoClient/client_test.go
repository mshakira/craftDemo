package main

import (
	"context"
	"fmt"
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

func TestGetResponse(t *testing.T) {
	// failure case
	res, err := GetResponse("not_found")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if res != nil {
		t.Errorf("Expected response, got nil")
	}

	// success case
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World")
	}))
	defer ts.Close()
	res, err = GetResponse(ts.URL)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}
	if res == nil {
		t.Errorf("Expected response, got nil")
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
	err = ValidateResponse(resp)
	if err == nil {
		t.Errorf("Expected content-type mismatch error")
	}

	// content length
	resp.Header["Content-Type"][0] = "application/json"
		resp.Header["Content-Length"] = []string{"50"}
	err = ValidateResponse(resp)
	if err == nil {
		t.Errorf("Expected error, got no error" )
	}
}

func TestParseBody(t *testing.T) {
	// failure case
	handler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "<html><body>Hello World!</body></html>")
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	incidents, err := ParseBody(resp)
	if err == nil {
		t.Errorf("Expected error, got no error")
	}
	if incidents != nil {
		t.Errorf("Expected nil, got %v\n", incidents )
	}

	// success case
	handler = func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"Name":"ServiceNowQuery","Report":[{"number":"INC1234"}]}`)
	}

	//req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w = httptest.NewRecorder()
	handler(w, req)

	resp = w.Result()
	incidents, err = ParseBody(resp)
	if err != nil {
		t.Errorf("Expected nil, got error %v\n",err)
	}
	if incidents == nil {
		t.Errorf("Expected data, got nil" )
	}

}

func TestMergeIncs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		mergeIncs(ctx,ch,out)
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

func TestGenerateAggReportPriority(t *testing.T) {
	obj := []Incident{{"a", "b", "c", "d", "High", "f"},
		{"b", "b", "c", "d", "High", "f"}}
	sum, err := GenerateAggReportPriority(obj)
	if err != nil {
		t.Errorf("Expected nil, got %v\n", err)
	}
	in := 0
	for _, elem := range *sum {
		in++
		if elem.Sum != 2 {
			t.Errorf("Expected 2, got %v\n", elem.Sum)
		}
	}

	if in != 1 {
		t.Errorf("Expected 1, got %v\n", in)
	}

}
