package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"myfast/internal/db"
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

	send(d, token, chat)
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		send(d, token, chat)
	}
}

func send(d *sql.DB, token, chat string) {
	rows, err := d.Query(`SELECT id, title, url FROM articles WHERE approved=1 AND posted=0`)
	if err != nil {
		fmt.Println("query:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var title, link string
		if err := rows.Scan(&id, &title, &link); err != nil {
			continue
		}
		text := fmt.Sprintf("%s\n%s", title, link)
		if err := postToTelegram(token, chat, text); err != nil {
			fmt.Println("telegram:", err)
			continue
		}
		if _, err := d.Exec(`UPDATE articles SET posted=1 WHERE id=?`, id); err != nil {
			fmt.Println("update:", err)
		}
	}
}

func postToTelegram(token, chat, text string) error {
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	data := url.Values{"chat_id": {chat}, "text": {text}}
	resp, err := http.PostForm(api, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram status %s", resp.Status)
	}
	return nil
}
