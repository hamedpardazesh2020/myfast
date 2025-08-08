package main

import (
	"flag"
	"fmt"
	"log"

	"myfast/internal/admin"
	"myfast/internal/db"
)

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	flag.Parse()

	database, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	addr := fmt.Sprintf(":%d", *port)
	if err := admin.Run(database, addr); err != nil {
		log.Fatal(err)
	}
}
