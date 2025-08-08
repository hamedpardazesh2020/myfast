package main

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
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
)

type Article struct {
	ID        int
	Title     string
	URL       string
	Content   string
	ImagePath string
	VideoPath string
}

func main() {
	db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/newsdb?charset=utf8mb4&parseTime=True")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	channelID := os.Getenv("TELEGRAM_CHANNEL")

	scrapeTicker := time.NewTicker(5 * time.Minute)
	postTicker := time.NewTicker(10 * time.Minute)
	defer scrapeTicker.Stop()
	defer postTicker.Stop()

	scrapeAndStore(db)
	postApprovedArticles(db, botToken, channelID)

	for {
		select {
		case <-scrapeTicker.C:
			scrapeAndStore(db)
		case <-postTicker.C:
			postApprovedArticles(db, botToken, channelID)
		}
	}
}

func scrapeAndStore(db *sql.DB) {
	fmt.Println("Scraping...")

	resp, err := http.Get("https://www.cbc.ca/news/local")
	if err != nil {
		fmt.Printf("Error fetching page: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Non-OK HTTP status: %s\n", resp.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing page: %v\n", err)
		return
	}

	doc.Find("a.card").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".headline").Text()
		link, _ := s.Attr("href")
		if link == "" || title == "" {
			return
		}

		fullURL := "https://www.cbc.ca" + link

		var exists bool
		err := db.QueryRow("SELECT 1 FROM articles WHERE url = ?", fullURL).Scan(&exists)
		if err == nil {
			return
		}

		article := Article{Title: title, URL: fullURL}
		fetchArticleDetail(&article)

		_, err = db.Exec(`
            INSERT INTO articles (title, url, content, image_path, video_path)
            VALUES (?, ?, ?, ?, ?)`,
			article.Title, article.URL, article.Content, article.ImagePath, article.VideoPath,
		)
		if err != nil {
			fmt.Printf("DB insert error: %v\n", err)
		}
	})
}

func fetchArticleDetail(a *Article) {
	resp, err := http.Get(a.URL)
	if err != nil {
		fmt.Printf("Error fetching article: %v\n", err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing article: %v\n", err)
		return
	}

	a.Content = doc.Find(".story").Text()

	if imgURL, ok := doc.Find("img").First().Attr("src"); ok {
		a.ImagePath = downloadFile(imgURL, "images")
	}

	if videoURL, ok := doc.Find("video source").First().Attr("src"); ok {
		a.VideoPath = downloadFile(videoURL, "videos")
	}
}

func downloadFile(url string, folder string) string {
	if url == "" {
		return ""
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", url, err)
		return ""
	}
	defer resp.Body.Close()

	os.MkdirAll(folder, os.ModePerm)
	fileName := filepath.Join(folder, filepath.Base(url))

	out, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		return ""
	}
	return fileName
}

func postApprovedArticles(db *sql.DB, token, channel string) {
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
		var a Article
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

func sendToTelegram(token, channel string, a *Article) error {
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
