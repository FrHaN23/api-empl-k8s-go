package db

import (
	"api-empl-k8s-go/dotenv"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Conn() (*sqlx.DB, error) {
	port, err := strconv.Atoi(dotenv.Env("DB_PORT"))
	if err != nil {
		panic("failed parse db port")
	}
	dsn := fmt.Sprintf(
		"host=%s "+
			"user=%s "+
			"password=%s "+
			"dbname=%s "+
			"port=%d "+
			"sslmode=disable "+
			"TimeZone=%s",
		dotenv.Env("DB_HOST"),
		dotenv.Env("DB_USERNAME"),
		dotenv.Env("DB_PASSWORD"),
		dotenv.Env("DB_NAME"),
		port,
		dotenv.Env("DB_TZ"),
	)

	DB, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		panic("failed to open db: " + err.Error())
	}
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Print("db connection open")
	return DB, nil
}
