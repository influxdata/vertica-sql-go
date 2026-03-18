package vertigo

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		want    *Config
		wantErr bool
	}{
		{
			name: "full DSN",
			dsn:  "vertica://dbadmin:password@localhost:5433/testdb?tlsmode=none&client_label=mylabel&use_prepared_statements=1&autocommit=1&connection_load_balance=1&workload=analytics&backup_server_node=host2:5433,host3:5433",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "testdb",
				User:                  "dbadmin",
				Password:              "password",
				TLSMode:               "none",
				ClientLabel:           "mylabel",
				UsePreparedStatements: true,
				AutoCommit:            true,
				LoadBalance:           true,
				Workload:              "analytics",
				BackupHosts:           []string{"host2:5433", "host3:5433"},
			},
		},
		{
			name: "minimal DSN",
			dsn:  "vertica://localhost:5433/testdb",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "testdb",
				TLSMode:               "none",
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
		{
			name: "defaults for use_prepared_statements and autocommit",
			dsn:  "vertica://user:pass@host:5433/db",
			want: &Config{
				Host:                  "host:5433",
				Database:              "db",
				User:                  "user",
				Password:              "pass",
				TLSMode:               "none",
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
		{
			name: "use_prepared_statements disabled",
			dsn:  "vertica://localhost:5433/db?use_prepared_statements=0",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "none",
				UsePreparedStatements: false,
				AutoCommit:            true,
			},
		},
		{
			name: "autocommit disabled",
			dsn:  "vertica://localhost:5433/db?autocommit=0",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "none",
				UsePreparedStatements: true,
				AutoCommit:            false,
			},
		},
		{
			name: "tlsmode prefer",
			dsn:  "vertica://localhost:5433/db?tlsmode=Prefer",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "prefer",
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
		{
			name: "oauth access token",
			dsn:  "vertica://localhost:5433/db?oauth_access_token=mytoken",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "none",
				OAuthAccessToken:      "mytoken",
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
		{
			name: "valid TOTP",
			dsn:  "vertica://localhost:5433/db?totp=123456",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "none",
				TOTP:                  "123456",
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
		{
			name:    "invalid TOTP",
			dsn:     "vertica://localhost:5433/db?totp=bad",
			wantErr: true,
		},
		{
			name: "single backup host",
			dsn:  "vertica://localhost:5433/db?backup_server_node=backup1:5433",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "none",
				BackupHosts:           []string{"backup1:5433"},
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
		{
			name: "load balance disabled by default",
			dsn:  "vertica://localhost:5433/db",
			want: &Config{
				Host:                  "localhost:5433",
				Database:              "db",
				TLSMode:               "none",
				UsePreparedStatements: true,
				AutoCommit:            true,
				LoadBalance:           false,
			},
		},
		{
			name: "no database",
			dsn:  "vertica://localhost:5433",
			want: &Config{
				Host:                  "localhost:5433",
				TLSMode:               "none",
				UsePreparedStatements: true,
				AutoCommit:            true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDSN(tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseDSN() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDSN() =\n  %+v\nwant\n  %+v", got, tt.want)
			}
		})
	}
}

func TestConfigURL(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantQuery map[string]string
	}{
		{
			name: "full config",
			cfg: Config{
				Host:                  "localhost:5433",
				Database:              "testdb",
				User:                  "dbadmin",
				Password:              "password",
				TLSMode:               "prefer",
				ClientLabel:           "mylabel",
				UsePreparedStatements: true,
				AutoCommit:            true,
				LoadBalance:           true,
				Workload:              "analytics",
				BackupHosts:           []string{"host2:5433", "host3:5433"},
			},
			wantQuery: map[string]string{
				"tlsmode":                 "prefer",
				"client_label":            "mylabel",
				"use_prepared_statements": "1",
				"autocommit":              "1",
				"connection_load_balance": "1",
				"workload":                "analytics",
				"backup_server_node":      "host2:5433,host3:5433",
			},
		},
		{
			name: "minimal config",
			cfg: Config{
				Host:     "localhost:5433",
				Database: "db",
			},
			wantQuery: map[string]string{
				"use_prepared_statements": "0",
				"autocommit":              "0",
				"connection_load_balance": "0",
			},
		},
		{
			name: "disabled flags",
			cfg: Config{
				Host:                  "localhost:5433",
				Database:              "db",
				UsePreparedStatements: false,
				AutoCommit:            false,
				LoadBalance:           false,
			},
			wantQuery: map[string]string{
				"use_prepared_statements": "0",
				"autocommit":              "0",
				"connection_load_balance": "0",
			},
		},
		{
			name: "oauth token",
			cfg: Config{
				Host:             "localhost:5433",
				Database:         "db",
				OAuthAccessToken: "tok123",
			},
			wantQuery: map[string]string{
				"oauth_access_token": "tok123",
			},
		},
		{
			name: "totp",
			cfg: Config{
				Host:     "localhost:5433",
				Database: "db",
				TOTP:     "654321",
			},
			wantQuery: map[string]string{
				"totp": "654321",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.cfg.URL()

			if u.Scheme != "vertica" {
				t.Errorf("URL().Scheme = %q, want %q", u.Scheme, "vertica")
			}
			if u.Host != tt.cfg.Host {
				t.Errorf("URL().Host = %q, want %q", u.Host, tt.cfg.Host)
			}
			if u.Path != "/"+tt.cfg.Database {
				t.Errorf("URL().Path = %q, want %q", u.Path, tt.cfg.Database)
			}
			if tt.cfg.User != "" {
				if u.User.Username() != tt.cfg.User {
					t.Errorf("URL().User.Username() = %q, want %q", u.User.Username(), tt.cfg.User)
				}
				p, _ := u.User.Password()
				if p != tt.cfg.Password {
					t.Errorf("URL().User.Password() = %q, want %q", p, tt.cfg.Password)
				}
			} else if u.User != nil {
				t.Errorf("URL().User = %v, want nil", u.User)
			}

			q := u.Query()
			for key, want := range tt.wantQuery {
				if got := q.Get(key); got != want {
					t.Errorf("URL() query %q = %q, want %q", key, got, want)
				}
			}

			// Verify absent fields don't leak into query.
			if tt.cfg.TLSMode == "" {
				if q.Has("tlsmode") {
					t.Errorf("URL() has tlsmode query param but TLSMode is empty")
				}
			}
			if tt.cfg.ClientLabel == "" {
				if q.Has("client_label") {
					t.Errorf("URL() has client_label query param but ClientLabel is empty")
				}
			}
			if tt.cfg.OAuthAccessToken == "" {
				if q.Has("oauth_access_token") {
					t.Errorf("URL() has oauth_access_token query param but OAuthAccessToken is empty")
				}
			}
			if tt.cfg.TOTP == "" {
				if q.Has("totp") {
					t.Errorf("URL() has totp query param but TOTP is empty")
				}
			}
			if tt.cfg.Workload == "" {
				if q.Has("workload") {
					t.Errorf("URL() has workload query param but Workload is empty")
				}
			}
			if len(tt.cfg.BackupHosts) == 0 {
				if q.Has("backup_server_node") {
					t.Errorf("URL() has backup_server_node query param but BackupHosts is empty")
				}
			}
		})
	}
}

func TestConfigURLRoundTrip(t *testing.T) {
	// Verify that building a URL from a Config and parsing it back
	// produces an equivalent Config (for fields that survive the round trip).
	cfg := Config{
		Host:                  "localhost:5433",
		Database:              "testdb",
		User:                  "dbadmin",
		Password:              "secret",
		TLSMode:               "prefer",
		ClientLabel:           "mylabel",
		UsePreparedStatements: true,
		AutoCommit:            false,
		LoadBalance:           true,
		Workload:              "analytics",
	}

	u := cfg.URL()
	dsn := u.String()

	if !strings.Contains(dsn, "vertica://") {
		t.Fatalf("URL().String() = %q, expected vertica:// scheme", dsn)
	}
	if !strings.Contains(dsn, "dbadmin:secret@") {
		t.Fatalf("URL().String() = %q, expected user info", dsn)
	}
	if !strings.Contains(dsn, "localhost:5433") {
		t.Fatalf("URL().String() = %q, expected host", dsn)
	}
}

func TestConfigDriver(t *testing.T) {
	cfg := &Config{}
	d := cfg.Driver()
	if _, ok := d.(*Driver); !ok {
		t.Errorf("Driver() returned %T, want *Driver", d)
	}
}
