package seed

import (
	"fmt"
	"mig/db"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
)

type SeederPostgreSQL struct {
	conn    *pgx.Conn
	queries *db.Queries
	faker   *gofakeit.Faker
}

func NewSeederPostgreSQL(conn *pgx.Conn, queries *db.Queries) (*SeederPostgreSQL, error) {
	if conn == nil {
		return nil, fmt.Errorf("missing conn")
	}

	if queries == nil {
		return nil, fmt.Errorf("missing queries")
	}

	seeder := &SeederPostgreSQL{
		conn:    conn,
		queries: queries,
		faker:   gofakeit.New(0),
	}

	return seeder, nil
}
