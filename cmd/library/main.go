package main

import (
	"awesomeProject4/internal/server"
	"awesomeProject4/internal/storage/inMemory"
	"log"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	log.Println("Library service started..")
	//TODO configuration server
	repo := inMemory.NewStorage()
	server := server.NewLiraryAPI(repo)
	//TODO start server
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
	//TODO stop server
	log.Println("Library service stopped")
}
