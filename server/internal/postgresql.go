package mig

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func NewDBConnection(user, pass, host, port, dbName, appName, appVersion string) (*sql.DB, error) {
	params := url.Values{}

	params.Add("sslmode", "disable")
	params.Add("application_name", fmt.Sprintf("%s %s", appName, appVersion))

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		user,
		pass,
		host,
		port,
		dbName,
		params.Encode(),
	)

	config, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	conn := stdlib.OpenDB(*config)

	return conn, nil
}
