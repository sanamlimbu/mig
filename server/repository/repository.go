package repository

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"
)

type PostgreSQLConnectionConfig struct {
	User       string
	Pass       string
	Host       string
	Port       string
	DbName     string
	AppName    string
	AppVersion string
}

func NewPostgreSQLConnection(ctx context.Context, config PostgreSQLConnectionConfig) (*pgx.Conn, error) {
	params := url.Values{}

	params.Add("sslmode", "disable")
	params.Add("application_name", fmt.Sprintf("%s-%s", config.AppName, config.AppVersion))

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		config.User,
		config.Pass,
		config.Host,
		config.Port,
		config.DbName,
		params.Encode(),
	)

	connConfig, err := pgx.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	conn, err := pgx.ConnectConfig(ctx, connConfig)

	return conn, err
}

func RegisterDataTypes(ctx context.Context, conn *pgx.Conn) error {
	dataTypeNames := []string{
		"chatroom_workflow_state",
		"_chatroom_workflow_state",
		"chatroom_type",
		"_chatroom_type",
		"friendship_workflow_state",
		"_friendship_workflow_state",
		"message_workflow_state",
		"_message_workflow_state",
		"message_type",
		"_message_type",
		"user_workflow_state",
		"_user_workflow_state",
	}

	for _, typeName := range dataTypeNames {
		dataType, err := conn.LoadType(ctx, typeName)
		if err != nil {
			return err
		}
		conn.TypeMap().RegisterType(dataType)
	}

	return nil
}
