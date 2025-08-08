package scraper

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"myfast/internal/article"
)

// Run starts the scraping loop checking the CBC Local site every five minutes.
func Run(d *sql.DB) {
	scrape(d)
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		scrape(d)
	}
}

func scrape(d *sql.DB) {
	resp, err := http.Get("https://www.cbc.ca/news/local")
	if err != nil {
		fmt.Println("fetch:", err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("parse:", err)
		return
	}

	doc.Find("a.card").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".headline").Text())
		href, _ := s.Attr("href")
		if title == "" || href == "" {
			return
		}
		url := "https://www.cbc.ca" + href
		var exists int
		err := d.QueryRow("SELECT 1 FROM articles WHERE url=?", url).Scan(&exists)
		if err == nil {
			return
		}

		art := article.Article{Title: title, URL: url}
		fetchDetail(&art)
		if art.Content == "" {
			return
		}
		_, err = d.Exec(`INSERT INTO articles (title,url,content,image_path,video_path,approved,posted) VALUES (?,?,?,?,?,0,0)`,
			art.Title, art.URL, art.Content, art.ImagePath, art.VideoPath)
		if err != nil {
			fmt.Println("insert:", err)
		}
	})
}

func fetchDetail(a *article.Article) {
	resp, err := http.Get(a.URL)
	if err != nil {
		fmt.Println("detail fetch:", err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("detail parse:", err)
		return
	}
	a.Content = strings.TrimSpace(doc.Find(".story").Text())

	if img, ok := doc.Find("img").First().Attr("src"); ok {
		a.ImagePath = download(img, "images")
	}
	if vid, ok := doc.Find("video source").First().Attr("src"); ok {
		a.VideoPath = download(vid, "videos")
	}
}

func download(u, dir string) string {
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("download:", err)
		return ""
	}
	defer resp.Body.Close()

	os.MkdirAll(dir, 0755)
	name := filepath.Join(dir, filepath.Base(u))
	f, err := os.Create(name)
	if err != nil {
		fmt.Println("create:", err)
		return ""
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		fmt.Println("copy:", err)
		return ""
	}
	return name
}
