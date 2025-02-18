package repository

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreSQLConnPoolConfig struct {
	User       string
	Pass       string
	Host       string
	Port       string
	DbName     string
	AppName    string
	AppVersion string
	Env        string
}

func NewPostgreSQLConnPool(ctx context.Context, config PostgreSQLConnPoolConfig) (*pgxpool.Pool, error) {
	params := url.Values{}

	if config.Env == "development" {
		params.Add("sslmode", "disable")
	} else {
		params.Add("sslmode", "require")
	}

	params.Add("application_name", fmt.Sprintf("%s-%s", config.AppName, config.AppVersion))

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?%s",
		config.User,
		config.Pass,
		config.Host,
		config.Port,
		config.DbName,
		params.Encode(),
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	poolConfig.AfterConnect = RegisterDataTypes

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	return pool, err
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

func UUIDToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}

	return encodeUUID(uuid.Bytes)
}

func StringToUUID(str string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(str)
	return uuid, err
}

func encodeUUID(src [16]byte) string {
	var buf [36]byte

	hex.Encode(buf[0:8], src[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], src[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], src[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], src[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], src[10:])

	return string(buf[:])
}

func NewPgTypeUUID() (pgtype.UUID, error) {
	var pgtypeUUID pgtype.UUID

	uuid := uuid.New()
	err := pgtypeUUID.Scan(uuid.String())

	return pgtypeUUID, err
}
