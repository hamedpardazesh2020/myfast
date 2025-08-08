package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/PuerkitoBio/goquery"

	"myfast/internal/article"
)

// ScrapeAndStore fetches the CBC local news page and stores new articles in the database.
func ScrapeAndStore(db *sql.DB) {
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

		var id int
		err := db.QueryRow("SELECT id FROM articles WHERE url = ?", fullURL).Scan(&id)
		if err != sql.ErrNoRows {
			if err == nil {
				return
			}
			fmt.Printf("DB lookup error: %v\n", err)
			return
		}

		a := article.Article{Title: title, URL: fullURL}
		fetchArticleDetail(&a)

		_, err = db.Exec(`
            INSERT INTO articles (title, url, content, image_path, video_path, approved, posted)
            VALUES (?, ?, ?, ?, ?, 0, 0)`,
			a.Title, a.URL, a.Content, a.ImagePath, a.VideoPath,
		)
		if err != nil {
			fmt.Printf("DB insert error: %v\n", err)
		}
	})
}

func fetchArticleDetail(a *article.Article) {
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

	if _, err = io.Copy(out, resp.Body); err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		return ""
	}
	return fileName
}
