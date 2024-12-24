package main

import (
	"context"
	"fmt"
	"mig/api"
	"mig/chatroom"
	"mig/db"
	"mig/repository"
	"mig/seed"
	"mig/user"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"
)

const version string = "v1.0.0"

func main() {
	app := &cli.App{
		Compiled: time.Now(),
		Name:     "mig server",
		Version:  version,
		Authors: []*cli.Author{
			{
				Name:  "Sanam Limbu",
				Email: "sudosanam@gmail.com",
			}},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Usage:   "show version",
				Aliases: []string{"v"},
				Action: func(c *cli.Context) error {
					fmt.Println(c.App.Version)
					return nil
				},
			},
			{
				Name:    "serve",
				Usage:   "run api server",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "addr", Value: "localhost:8080", EnvVars: []string{"MIG_ADDR"}, Usage: "host:port address of server"},
					&cli.StringFlag{Name: "environment", Value: "dev", EnvVars: []string{"MIG_ENVIRONMENT"}, Usage: "deployment environnment (dev, prod) of server"},
					&cli.StringFlag{Name: "jwt_secret", Value: "devdev", EnvVars: []string{"MIG_JWT_SECRET"}, Usage: "secret to sign jwt"},
					&cli.StringFlag{Name: "app_name", Value: "mig-api-server", EnvVars: []string{"MIG_APP_NAME"}, Usage: "application name"},
					&cli.StringFlag{Name: "nats_secret", Value: "my-nats-secret", EnvVars: []string{"MIG_NATS_SECRET"}, Usage: "NATS secret token"},

					&cli.StringFlag{Name: "database_user", Value: "mig", EnvVars: []string{"MIG_DATABASE_USER"}, Usage: "database user"},
					&cli.StringFlag{Name: "database_pass", Value: "devdev", EnvVars: []string{"MIG_DATABASE_PASS"}, Usage: "database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{"MIG_DATABASE_HOST"}, Usage: "database host"},
					&cli.StringFlag{Name: "database_port", Value: "5435", EnvVars: []string{"MIG_DATABASE_PORT"}, Usage: "database port"},
					&cli.StringFlag{Name: "database_name", Value: "mig", EnvVars: []string{"MIG_DATABASE_NAME"}, Usage: "database name"},
				},
				Action: func(c *cli.Context) error {
					err := serve(c)
					if err != nil {
						log.Panic().Msg(err.Error())
					}
					return err
				},
			},
			{
				Name:  "seed",
				Usage: "seed database",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "addr", Value: "localhost:8080", EnvVars: []string{"MIG_ADDR"}, Usage: "host:port address of server"},
					&cli.StringFlag{Name: "environment", Value: "dev", EnvVars: []string{"MIG_ENVIRONMENT"}, Usage: "deployment environnment (dev, prod) of server"},
					&cli.StringFlag{Name: "jwt_secret", Value: "devdev", EnvVars: []string{"MIG_JWT_SECRET"}, Usage: "secret to sign jwt"},
					&cli.StringFlag{Name: "app_name", Value: "mig-api-server", EnvVars: []string{"MIG_APP_NAME"}, Usage: "application name"},

					&cli.StringFlag{Name: "database_user", Value: "mig", EnvVars: []string{"MIG_DATABASE_USER"}, Usage: "database user"},
					&cli.StringFlag{Name: "database_pass", Value: "devdev", EnvVars: []string{"MIG_DATABASE_PASS"}, Usage: "database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{"MIG_DATABASE_HOST"}, Usage: "database host"},
					&cli.StringFlag{Name: "database_port", Value: "5435", EnvVars: []string{"MIG_DATABASE_PORT"}, Usage: "database port"},
					&cli.StringFlag{Name: "database_name", Value: "mig", EnvVars: []string{"MIG_DATABASE_NAME"}, Usage: "database name"},
				},
				Action: func(c *cli.Context) error {
					err := seedDb(c)
					if err != nil {
						log.Panic().Msg(err.Error())
					}
					return err
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func serve(c *cli.Context) error {
	addr := c.String("addr")
	if addr == "" {
		return fmt.Errorf("missing env: MIG_ADDR")
	}

	env := c.String("environment")
	if env == "" {
		return fmt.Errorf("missing env: MIG_ENV")
	}

	jwtSecret := c.String("jwt_secret")
	if jwtSecret == "" {
		return fmt.Errorf("missing env: MIG_JWT_SECRET")
	}

	natsSecret := c.String("nats_secret")
	if natsSecret == "" {
		return fmt.Errorf("missing env: MIG_NATS_SECRET")
	}

	conn, err := connectPostgreSQL(c)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(c.Context); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	err = repository.RegisterDataTypes(c.Context, conn)
	if err != nil {
		return err
	}

	queries := db.New(conn)

	userRepo, err := repository.NewUserRepositoryPostgreSQL(queries)
	if err != nil {
		return err
	}

	userService, err := user.NewService(userRepo)
	if err != nil {
		return err
	}

	chatroomRepo, err := repository.NewChatroomRepositoryPostgreSQL(queries)
	if err != nil {
		return err
	}

	chatroomService, err := chatroom.NewService(chatroomRepo)
	if err != nil {
		return err
	}

	controller, err := api.NewHttpApiController(userService, chatroomService)
	if err != nil {
		return err
	}

	router, err := api.NewHttpRouter(controller)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Info().Msg("starting server...")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().Msg(err.Error())
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	signal := <-signalChan

	ctx, cancel := context.WithTimeout(c.Context, time.Second*5)
	defer cancel()

	log.Info().Msg(fmt.Sprintf("received signal %s, shutting down http server gracefully...", signal))
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Msg(err.Error())
		return err
	}

	return nil
}

func connectPostgreSQL(c *cli.Context) (*pgx.Conn, error) {
	dbUser := c.String("database_user")
	if dbUser == "" {
		return nil, fmt.Errorf("missing env: MIG_DATABASE_USER")
	}

	dbPass := c.String("database_pass")
	if dbPass == "" {
		return nil, fmt.Errorf("missing env: MIG_DATABASE_PASS")
	}

	dbHost := c.String("database_host")
	if dbHost == "" {
		return nil, fmt.Errorf("missing env: MIG_DATABASE_HOST")
	}

	dbPort := c.String("database_port")
	if dbPort == "" {
		return nil, fmt.Errorf("missing env: MIG_DATABASE_PORT")
	}

	dbName := c.String("database_name")
	if dbName == "" {
		return nil, fmt.Errorf("missing env: MIG_DATABASE_NAME")
	}

	appName := c.String("app_name")
	if appName == "" {
		return nil, fmt.Errorf("missing env: MIG_APP_NAME")
	}

	config := repository.PostgreSQLConnectionConfig{
		User:       dbUser,
		Pass:       dbPass,
		Host:       dbHost,
		Port:       dbPort,
		DbName:     dbName,
		AppName:    appName,
		AppVersion: version,
	}

	conn, err := repository.NewPostgreSQLConnection(c.Context, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func seedDb(c *cli.Context) error {
	ctx := c.Context

	conn, err := connectPostgreSQL(c)
	if err != nil {
		return err
	}

	defer func() {
		if err := conn.Close(ctx); err != nil {
			log.Error().Msg(err.Error())
		}
	}()

	queries := db.New(conn)

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

	if err := seeder.Friendships(ctx, seed.ChatroomUUIDs[:]); err != nil {
		return err
	}
	fmt.Println("seeded friendships...")

	if err := seeder.Messages(ctx, seed.UsersUUIDs[:], seed.ChatroomUUIDs[:]); err != nil {
		return err
	}
	fmt.Println("seeded messages...")

	return nil
}
