package main

import (
	"log"
	"time"

	"myfast/internal/db"
	"myfast/internal/scraper"
)

func main() {
	database, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	scraper.ScrapeAndStore(database)
	for range ticker.C {
		scraper.ScrapeAndStore(database)
	}
}
