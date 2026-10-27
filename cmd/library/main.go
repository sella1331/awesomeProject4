package main

import (
	"awesomeProject4/internal/server"
	"awesomeProject4/internal/storage/inMemory"
	"awesomeProject4/internal/storage/inMemory/db"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Println("Library service started..")
	//TODO configuration server
	var usersStorage server.UserRepository
	var tasksStorage server.TaskRepository
	if dbSql, err := inMemory.RunDB(); err != nil {
		fmt.Println(err)
		log.Println("🔄 Переключение на резервное хранилище (In-Memory)...")
		usersStorage = inMemory.UserStorage()
		tasksStorage = inMemory.NewStorage()
	} else {
		log.Println("✅ Успешное подключение к PostgreSQL! Используется база данных.")
		defer dbSql.Close()

		usersStorage = db.NewPsqlStorage(dbSql)
		tasksStorage = db.NewPsqlStorageTask(dbSql)
	}
	//
	server := server.NewLiraryAPI(tasksStorage, usersStorage)
	//TODO start server
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
	//TODO stop server
	log.Println("Library service stopped")

}
