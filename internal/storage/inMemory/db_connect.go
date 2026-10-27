package inMemory

import (
	"database/sql"
	"fmt"
)

func RunDB() (*sql.DB, error) {

	dsn := "user=psql password=pgpass host=localhost port=5432 dbname=shortener_db sslmode=disable"

	Db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия пула: %w", err)
	}

	if err := Db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка проверки соединения с БД (Ping): %w", err)

	}

	fmt.Println("Successfully connected to DB")
	return Db, nil
}
