package vertigo

import (
	"context"
	"database/sql/driver"
	"net"
	"net/url"
	"strings"
)

// Config holds the configuration for creating database connections.
type Config struct {
	Host        string
	Database    string
	BackupHosts []string
	TLSMode     string

	User             string
	Password         string
	OAuthAccessToken string
	TOTP             string

	ClientLabel           string
	UsePreparedStatements bool
	AutoCommit            bool
	LoadBalance           bool
	Workload              string

	DialContext func(context.Context, string, string) (net.Conn, error)
}

func (c *Config) URL() *url.URL {
	u := &url.URL{
		Scheme: "vertica",
		Host:   c.Host,
		Path:   "/" + c.Database,
	}
	if c.User != "" || c.Password != "" {
		u.User = url.UserPassword(c.User, c.Password)
	} else if c.OAuthAccessToken != "" {
		// Ensure userinfo is preserved / non-nil in OAuth-only configurations.
		u.User = url.User(c.User)
	}
	v := make(url.Values)
	if len(c.BackupHosts) > 0 {
		v.Add("backup_server_node", strings.Join(c.BackupHosts, ","))
	}
	if c.TLSMode != "" {
		v.Add("tlsmode", c.TLSMode)
	}
	if c.OAuthAccessToken != "" {
		v.Add("oauth_access_token", c.OAuthAccessToken)
	}
	if c.TOTP != "" {
		v.Add("totp", c.TOTP)
	}
	if c.ClientLabel != "" {
		v.Add("client_label", c.ClientLabel)
	}
	if c.UsePreparedStatements {
		v.Add("use_prepared_statements", "1")
	} else {
		v.Add("use_prepared_statements", "0")
	}
	if c.AutoCommit {
		v.Add("autocommit", "1")
	} else {
		v.Add("autocommit", "0")
	}
	if c.LoadBalance {
		v.Add("connection_load_balance", "1")
	} else {
		v.Add("connection_load_balance", "0")
	}
	if c.Workload != "" {
		v.Add("workload", c.Workload)
	}
	u.RawQuery = v.Encode()
	return u
}

// Connect implements driver.Connector
func (c *Config) Connect(ctx context.Context) (driver.Conn, error) {
	return newConnection(ctx, c)
}

// Driver implements driver.Connector
func (*Config) Driver() driver.Driver {
	return &Driver{}
}

func (c *Config) dial(ctx context.Context, network, address string) (net.Conn, error) {
	if c.DialContext != nil {
		return c.DialContext(ctx, network, address)
	} else {
		return net.Dial(network, address)
	}
}

// ParseDSN parses a Vertica connection string into a Config.
// The connection string format is: vertica://user:password@host:port/database?key=value
func ParseDSN(connString string) (*Config, error) {
	u, err := url.Parse(connString)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Host:                  u.Host,
		UsePreparedStatements: true,
		AutoCommit:            true,
	}

	if u.User != nil {
		cfg.User = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}

	cfg.Database = strings.TrimPrefix(u.Path, "/")

	q := u.Query()

	if v := q.Get("client_label"); v != "" {
		cfg.ClientLabel = v
	}

	if v := q.Get("use_prepared_statements"); v != "" {
		cfg.UsePreparedStatements = v == "1"
	}

	if v := q.Get("autocommit"); v != "" {
		cfg.AutoCommit = v == "1"
	}

	cfg.OAuthAccessToken = q.Get("oauth_access_token")

	if v := q.Get("totp"); v != "" {
		if err := validateTOTP(v); err != nil {
			return nil, err
		}
		cfg.TOTP = v
	}

	cfg.LoadBalance = q.Get("connection_load_balance") == "1"

	if v := q.Get("backup_server_node"); v != "" {
		cfg.BackupHosts = strings.Split(v, ",")
	}

	cfg.TLSMode = strings.ToLower(q.Get("tlsmode"))
	if cfg.TLSMode == "" {
		cfg.TLSMode = tlsModeNone
	}

	cfg.Workload = q.Get("workload")

	return cfg, nil
}
