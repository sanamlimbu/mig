package repotest

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"mig/db"
	"mig/repository"
	"mig/seed"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pressly/goose"
)

var userRepo repository.UserRepository
var chatroomRepo repository.ChatroomRepository

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	dbUser := "mig"
	dbPassword := "devdev"
	dbName := "mig"
	dbPort := "5436"

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16-alpine",
		Env: []string{
			fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			"listen_addresses = '*'",
		},
		ExposedPorts: []string{dbPort},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432/tcp": {
				{HostIP: "0.0.0.0", HostPort: dbPort},
			},
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable&application_name=test", dbUser, dbPassword, hostAndPort, dbName)

	if err := resource.Expire(120); err != nil {
		log.Println(err.Error())
	}

	ctx := context.Background()

	var conn *pgx.Conn

	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		conn, err = pgx.Connect(ctx, dbConnStr)
		if err != nil {
			return err
		}

		return conn.Ping(ctx)
	}); err != nil {
		log.Fatalf("could not connect to docker: %s", err)
	}

	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("could not purge resource: %s", err)
		}
	}()

	defer func() {
		if err := conn.Close(ctx); err != nil {
			log.Printf("unable to close database connection: %s", err)
		}
	}()

	err = runMigrationsUp(dbConnStr)
	if err != nil {
		log.Fatalf("unable to run migrations: %s", err)
	}

	err = repository.RegisterDataTypes(ctx, conn)
	if err != nil {
		log.Fatalf("unable to register pgx data types: %s", err)
	}

	queries := db.New(conn)

	userRepo, err = repository.NewUserRepositoryPostgreSQL(queries)
	if err != nil {
		log.Fatalf("could not create user repository: %s", err)
	}

	chatroomRepo, err = repository.NewChatroomRepositoryPostgreSQL(queries)
	if err != nil {
		log.Fatalf("could not create chatroom repository: %s", err)
	}

	err = seedDB(ctx, conn, queries)
	if err != nil {
		log.Fatalf("unable to seed database: %s", err)
	}

	m.Run()
}

func runMigrationsUp(dbConnStr string) error {
	migrations := "../../migrations"

	db, err := sql.Open("pgx", dbConnStr)
	if err != nil {
		return fmt.Errorf("could not open database connection: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("unable to close database connection: %s", err)
		}
	}()

	// Goose will log - goose: no migrations to run. current version: 20241209030923.
	// Ignore this message printed in console. Everthing is fine.
	if err := goose.Up(db, migrations); err != nil {
		return fmt.Errorf("could not apply migrations up: %w", err)
	}

	return nil
}

func seedDB(ctx context.Context, conn *pgx.Conn, queries *db.Queries) error {
	seeder, err := seed.NewSeederPostgreSQL(conn, queries)
	if err != nil {
		return err
	}

	_, err = seeder.Users(ctx, seed.UsersUUIDs[:])
	if err != nil {
		return err
	}
	fmt.Println("seeded users...")

	_, err = seeder.Chatrooms(ctx, seed.UsersUUIDs[:], seed.ChatroomUUIDs[:])
	if err != nil {
		return err
	}
	fmt.Println("seeded chatrooms...")

	if err := seeder.Friendships(ctx, seed.UsersUUIDs[:]); err != nil {
		return err
	}
	fmt.Println("seeded friendships...")

	if err := seeder.Messages(ctx, seed.UsersUUIDs[:], seed.ChatroomUUIDs[:]); err != nil {
		return err
	}
	fmt.Println("seeded messages...")

	return nil
}
