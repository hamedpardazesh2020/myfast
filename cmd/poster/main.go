package main

import (
	"os"

	"myfast/internal/db"
	"myfast/internal/poster"
)

func main() {
	d, err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	token := os.Getenv("TELEGRAM_TOKEN")
	chat := os.Getenv("TELEGRAM_CHAT")
	if token == "" || chat == "" {
		panic("TELEGRAM_TOKEN and TELEGRAM_CHAT must be set")
	}

	poster.Run(d, token, chat)
}
