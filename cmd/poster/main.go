package main

import (
	"log"
	"time"

	"myfast/internal/db"
	"myfast/internal/poster"
	"myfast/internal/settings"
)

func main() {
	database, err := db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		token, _ := settings.Get(database, "telegram_token")
		channel, _ := settings.Get(database, "telegram_channel")
		poster.PostApprovedArticles(database, token, channel)
		<-ticker.C
	}
}
