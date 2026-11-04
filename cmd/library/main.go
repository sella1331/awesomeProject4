package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sella1331/awesomeProject4.git/internal/server"
	"github.com/sella1331/awesomeProject4.git/internal/storage/inMemory"
	"github.com/sella1331/awesomeProject4.git/internal/storage/inMemory/db"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
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
		defer func() {
			if err := dbSql.Close(); err != nil {
				log.Printf("error down db: %v", err)
			}
		}()

		psqlStorage := db.NewPsqlStorageTask(dbSql)
		usersStorage = psqlStorage
		tasksStorage = psqlStorage
	}
	server := server.NewLiraryAPI(tasksStorage, usersStorage)
	//TODO start server
	serverErrors := make(chan error, 1)
	go func() {
		log.Println("🚀 HTTP server is running...")
		if err := server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()
	//TODO stop server
	select {
	case err := <-serverErrors:
		log.Fatalf("Critical server error: %v", err)
	case <-ctx.Done():
		log.Println(" stopping...")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(" shutdown error: %v\n", err)
	}
	log.Println("service stopped")

}
