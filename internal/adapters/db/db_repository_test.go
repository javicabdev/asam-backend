package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestBuildDSN(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		user     string
		password string
	}{
		{name: "plain", host: "db.example.com", user: "asam", password: "secret"},
		{name: "empty password", host: "db.example.com", user: "asam", password: ""},
		{name: "space", host: "db.example.com", user: "asam", password: "p ss"},
		{name: "single quote", host: "db.example.com", user: "asam", password: "p'ss"},
		{name: "backslash", host: "db.example.com", user: "asam", password: `p\ss`},
		{name: "escaped quote sequence", host: "db.example.com", user: "asam", password: `p\'ss`},
		{name: "trailing backslash", host: "db.example.com", user: "asam", password: `pss\`},
		{name: "equals sign", host: "db.example.com", user: "asam", password: "p=ss"},
		{name: "url reserved chars", host: "db.example.com", user: "asam", password: "p@:/#%?ss"},
		{name: "all combined", host: "db.example.com", user: "as am'", password: `p ss'w\rd /#%?@:=`},
		{name: "unix socket host", host: "/cloudsql/project:region:instance", user: "asam", password: "p ss"},
	}

	const (
		dbName  = "asam_db"
		port    = "5432"
		sslMode = "disable"
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := buildDSN(tt.host, tt.user, tt.password, dbName, port, sslMode)

			// Parse with the same parser the pgx driver uses at connect time.
			cfg, err := pgconn.ParseConfig(dsn)
			if err != nil {
				t.Fatalf("pgconn.ParseConfig error: %v", err)
			}
			if cfg.Host != tt.host {
				t.Errorf("host = %q, want %q", cfg.Host, tt.host)
			}
			if cfg.Port != 5432 {
				t.Errorf("port = %d, want 5432", cfg.Port)
			}
			if cfg.User != tt.user {
				t.Errorf("user = %q, want %q", cfg.User, tt.user)
			}
			if cfg.Password != tt.password {
				t.Errorf("password = %q, want %q", cfg.Password, tt.password)
			}
			if cfg.Database != dbName {
				t.Errorf("database = %q, want %q", cfg.Database, dbName)
			}
			if cfg.TLSConfig != nil {
				t.Errorf("TLSConfig is set, want nil for sslmode=%s", sslMode)
			}
		})
	}
}
