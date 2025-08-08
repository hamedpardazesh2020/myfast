package poster

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"myfast/internal/article"
)

// PostApprovedArticles sends approved and unposted articles to a Telegram channel.
func PostApprovedArticles(db *sql.DB, token, channel string) {
	if token == "" || channel == "" {
		fmt.Println("Telegram credentials not set")
		return
	}

	rows, err := db.Query("SELECT id, title, content, image_path, video_path FROM articles WHERE approved=1 AND posted=0")
	if err != nil {
		fmt.Printf("DB query error: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var a article.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.ImagePath, &a.VideoPath); err != nil {
			fmt.Printf("Row scan error: %v\n", err)
			continue
		}

		if err := sendToTelegram(token, channel, &a); err != nil {
			fmt.Printf("Telegram send error: %v\n", err)
			continue
		}

		if _, err := db.Exec("UPDATE articles SET posted=1 WHERE id=?", a.ID); err != nil {
			fmt.Printf("DB update error: %v\n", err)
		}
	}
}

func sendToTelegram(token, channel string, a *article.Article) error {
	if err := sendText(token, channel, a.Title+"\n"+a.Content); err != nil {
		return err
	}
	if a.ImagePath != "" {
		if err := sendMedia(token, channel, "sendPhoto", "photo", a.ImagePath, ""); err != nil {
			return err
		}
	}
	if a.VideoPath != "" {
		if err := sendMedia(token, channel, "sendVideo", "video", a.VideoPath, ""); err != nil {
			return err
		}
	}
	return nil
}

func sendText(token, channel, text string) error {
	resp, err := http.PostForm(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token), url.Values{
		"chat_id": {channel},
		"text":    {text},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func sendMedia(token, channel, endpoint, field, path, caption string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	w.WriteField("chat_id", channel)
	if caption != "" {
		w.WriteField("caption", caption)
	}
	fw, err := w.CreateFormFile(field, filepath.Base(path))
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, file); err != nil {
		return err
	}
	w.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, endpoint), &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
