package main

import (
	"context"
	"fmt"
	"mig"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
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
				Usage:   "run the service",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "addr", Value: "localhost:8080", EnvVars: []string{"MIG_ADDR"}, Usage: "host:port address of the server"},
					&cli.StringFlag{Name: "environment", Value: "dev", EnvVars: []string{"MIG_ENVIRONMENT"}, Usage: "deployment environnment (dev, prod) of the server"},
					&cli.StringFlag{Name: "jwt_secret", Value: "devdev", EnvVars: []string{"MIG_JWT_SECRET"}, Usage: "secret to sign JWT"},

					&cli.StringFlag{Name: "database_user", Value: "postgres", EnvVars: []string{"MIG_DATABASE_USER"}, Usage: "database user"},
					&cli.StringFlag{Name: "database_pass", Value: "devdev", EnvVars: []string{"MIG_DATABASE_PASS"}, Usage: "database pass"},
					&cli.StringFlag{Name: "database_host", Value: "localhost", EnvVars: []string{"MIG_DATABASE_HOST"}, Usage: "database host"},
					&cli.StringFlag{Name: "database_port", Value: "5435", EnvVars: []string{"MIG_DATABASE_PORT"}, Usage: "database port"},
					&cli.StringFlag{Name: "database_name", Value: "postgres", EnvVars: []string{"MIG_DATABASE_NAME"}, Usage: "database name"},
					&cli.StringFlag{Name: "database_application_name", Value: "API Server", EnvVars: []string{"MIG_DATABASE_APPLICATION_NAME"}, Usage: "application name"},

					&cli.StringFlag{Name: "kafka_brokers", Value: "localhost:9092", EnvVars: []string{"MIG_KAFKA_BROKERS"}, Usage: "Kafka brokers to connect to, as a comma separated list"},
					&cli.StringFlag{Name: "kafka_group", Value: uuid.NewString(), EnvVars: []string{"MIG_KAFKA_GROUP"}, Usage: "Kafka consumer group definition"},
					&cli.StringFlag{Name: "kafka_topics", Value: "private,group", EnvVars: []string{"MIG_KAFKA_TOPICS"}, Usage: "Kafka topics, as a comma separated list"},
					&cli.StringFlag{Name: "kafka_version", Value: sarama.DefaultVersion.String(), EnvVars: []string{"MIG_KAFKA_VERSION"}, Usage: "Kafka cluster version"},
					&cli.StringFlag{Name: "kafka_assignor", Value: "range", EnvVars: []string{"MIG_KAFKA_ASSIGNOR"}, Usage: "Kafka consumer group partition assignment strategy (range, roundrobin, sticky)"},
				},
				Action: func(c *cli.Context) error {
					err := serve(c)
					if err != nil {
						log.Panic().Msg(err.Error())
					}
					return err
				},
			},
		},
	}

	app.Run(os.Args)
}

func serve(c *cli.Context) error {
	addr := c.String("addr")
	if addr == "" {
		return fmt.Errorf("missing env: MIG_ADDR")
	}

	env := c.String("environment")
	if env == "" {
		return fmt.Errorf("missing env: MIG_ENVIRONMENT")
	}

	jwtSecret := c.String("jwt_secret")
	if env == "" {
		return fmt.Errorf("missing env: MIG_JWT_SECRET")
	}

	dbUser := c.String("database_user")
	if dbUser == "" {
		return fmt.Errorf("missing env: MIG_DATABASE_USER")
	}

	dbPass := c.String("database_pass")
	if dbPass == "" {
		return fmt.Errorf("missing env: MIG_DATABASE_PASS")
	}

	dbHost := c.String("database_host")
	if dbHost == "" {
		return fmt.Errorf("missing env: MIG_DATABASE_HOST")
	}

	dbPort := c.String("database_port")
	if dbPort == "" {
		return fmt.Errorf("missing env: MIG_DATABASE_PORT")
	}

	dbName := c.String("database_name")
	if dbName == "" {
		return fmt.Errorf("missing env: MIG_DATABASE_NAME")
	}

	dbAppName := c.String("database_application_name")
	if dbAppName == "" {
		return fmt.Errorf("missing env: MIG_DATABASE_APPLICATION_NAME")
	}

	nats, err := mig.NewNats("ws", "mig", "devdev", "localhost", "89")

	db, err := mig.NewDBConnection(dbUser, dbPass, dbHost, dbPort, dbName, dbAppName, version)
	if err != nil {
		return err
	}

	auther := mig.NewAuther(jwtSecret, addr)

	groupsRepo := mig.NewGroupsRepositoryPostgreSQL(db)

	hub := mig.NewHub(nats)
	go hub.Run(c.Context)

	controller := mig.NewAPIController(db, auther, hub, groupsRepo)

	router := mig.NewRouter(controller)

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
