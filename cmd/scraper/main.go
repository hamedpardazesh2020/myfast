package main

import (
	"myfast/internal/db"
	"myfast/internal/scraper"
)

func main() {
	d, err := db.Connect()
	if err != nil {
		panic(err)
	}
	defer d.Close()

	scraper.Run(d)
}
