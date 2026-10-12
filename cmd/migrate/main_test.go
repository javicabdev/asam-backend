package main

import (
	"net/url"
	"testing"
)

func TestBuildDBURL(t *testing.T) {
	tests := []struct {
		name string
		user string
		pass string
	}{
		{name: "plain", user: "asam", pass: "secret"},
		{name: "at sign", user: "asam", pass: "p@ss"},
		{name: "colon", user: "asam", pass: "p:ss"},
		{name: "slash", user: "asam", pass: "p/ss"},
		{name: "hash", user: "asam", pass: "p#ss"},
		{name: "percent", user: "asam", pass: "p%ss"},
		{name: "question mark", user: "asam", pass: "p?ss"},
		{name: "space", user: "asam", pass: "p ss"},
		{name: "all reserved", user: "as@m:user", pass: "a@b:c/d#e%f?g h"},
	}

	const (
		host    = "db.example.com"
		port    = "5432"
		dbName  = "asam_db"
		sslMode = "require"
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := buildDBURL("postgres", tt.user, tt.pass, host, port, dbName, sslMode)

			u, err := url.Parse(raw)
			if err != nil {
				t.Fatalf("url.Parse(%q) error: %v", raw, err)
			}
			if u.Scheme != "postgres" {
				t.Errorf("scheme = %q, want %q", u.Scheme, "postgres")
			}
			if got := u.User.Username(); got != tt.user {
				t.Errorf("user = %q, want %q", got, tt.user)
			}
			if got, _ := u.User.Password(); got != tt.pass {
				t.Errorf("password = %q, want %q", got, tt.pass)
			}
			if u.Hostname() != host || u.Port() != port {
				t.Errorf("host = %q, want %q", u.Host, host+":"+port)
			}
			if u.Path != "/"+dbName {
				t.Errorf("path = %q, want %q", u.Path, "/"+dbName)
			}
			if got := u.Query().Get("sslmode"); got != sslMode {
				t.Errorf("sslmode = %q, want %q", got, sslMode)
			}
		})
	}
}
