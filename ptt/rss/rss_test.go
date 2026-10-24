package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckBoardExist(t *testing.T) {
	type args struct {
		board string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"exist", args{"movie"}, true},
		{"exist", args{"movies"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckBoardExist(tt.args.board); got != tt.want {
				t.Errorf("CheckBoardExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/toomany":
			w.WriteHeader(http.StatusTooManyRequests)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		case "/empty":
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom"><title>t</title></feed>`))
		}
	}))
	defer srv.Close()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"not found", srv.URL + "/notfound", true},
		{"too many requests", srv.URL + "/toomany", false},
		{"server error", srv.URL + "/error", false},
		{"empty feed", srv.URL + "/empty", false},
		{"unreachable", "http://127.0.0.1:1/unreachable", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotFound(tt.url); got != tt.want {
				t.Errorf("isNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
