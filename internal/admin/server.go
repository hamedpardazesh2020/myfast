package admin

import (
	"database/sql"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"myfast/internal/article"
	"myfast/internal/settings"
)

// Run starts the admin HTTP server on the given address.
func Run(db *sql.DB, addr string) error {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	http.HandleFunc("/", settingsHandler(db))
	http.HandleFunc("/unapproved", listHandler(db, "تایید نشده", "WHERE approved=0", "/unapproved/approve", "/unapproved/delete"))
	http.HandleFunc("/unapproved/approve", approveHandler(db))
	http.HandleFunc("/unapproved/delete", deleteHandler(db, "/unapproved"))
	http.HandleFunc("/approved", listHandler(db, "تایید شده", "WHERE approved=1", "", "/approved/delete"))
	http.HandleFunc("/approved/delete", deleteHandler(db, "/approved"))
	http.HandleFunc("/pending", listHandler(db, "در انتظار انتشار", "WHERE approved=1 AND posted=0", "", ""))
	http.HandleFunc("/posted", listHandler(db, "منتشر شده", "WHERE posted=1", "", ""))
	http.HandleFunc("/article", detailHandler(db))

	return http.ListenAndServe(addr, nil)
}

var settingsTmpl = template.Must(template.New("settings").Parse(`
<html><body>
<h1>تنظیمات تلگرام</h1>
<form method="POST" action="/settings">
<label>توکن ربات: <input name="token" value="{{.Token}}"></label><br>
<label>آدرس کانال: <input name="channel" value="{{.Channel}}"></label><br>
<button type="submit">ذخیره</button>
</form>
<ul>
<li><a href="/unapproved">اخبار تایید نشده</a></li>
<li><a href="/approved">اخبار تایید شده</a></li>
<li><a href="/pending">در انتظار انتشار در تلگرام</a></li>
<li><a href="/posted">منتشر شده در تلگرام</a></li>
</ul>
</body></html>
`))

func settingsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			settings.Set(db, "telegram_token", r.FormValue("token"))
			settings.Set(db, "telegram_channel", r.FormValue("channel"))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		token, _ := settings.Get(db, "telegram_token")
		channel, _ := settings.Get(db, "telegram_channel")
		settingsTmpl.Execute(w, struct{ Token, Channel string }{token, channel})
	}
}

var listTmpl = template.Must(template.New("list").Parse(`
<html><body>
<h1>{{.Title}}</h1>
{{if or .ApproveURL .DeleteURL}}
<form method="POST">
{{range .Articles}}
<div><input type="checkbox" name="id" value="{{.ID}}"> <a href="/article?id={{.ID}}">{{.Title}}</a></div>
{{end}}
{{if .ApproveURL}}<button formaction="{{.ApproveURL}}">تایید موارد انتخابی</button>{{end}}
{{if .DeleteURL}}<button formaction="{{.DeleteURL}}">حذف موارد انتخابی</button>{{end}}
</form>
{{else}}
<ul>
{{range .Articles}}
<li><a href="/article?id={{.ID}}">{{.Title}}</a></li>
{{end}}
</ul>
{{end}}
<a href="/">بازگشت</a>
</body></html>
`))

func listHandler(db *sql.DB, title, condition, approveURL, deleteURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, title FROM articles " + condition + " ORDER BY id DESC")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var list []article.Article
		for rows.Next() {
			var a article.Article
			if err := rows.Scan(&a.ID, &a.Title); err == nil {
				list = append(list, a)
			}
		}
		data := struct {
			Title      string
			Articles   []article.Article
			ApproveURL string
			DeleteURL  string
		}{title, list, approveURL, deleteURL}
		listTmpl.Execute(w, data)
	}
}

func approveHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := parseIDs(r)
		if len(ids) > 0 {
			q := "UPDATE articles SET approved=1 WHERE id IN (" + placeholders(len(ids)) + ")"
			args := make([]interface{}, len(ids))
			for i, id := range ids {
				args[i] = id
			}
			db.Exec(q, args...)
		}
		http.Redirect(w, r, "/unapproved", http.StatusSeeOther)
	}
}

func deleteHandler(db *sql.DB, redirect string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ids := parseIDs(r)
		if len(ids) > 0 {
			q := "DELETE FROM articles WHERE id IN (" + placeholders(len(ids)) + ")"
			args := make([]interface{}, len(ids))
			for i, id := range ids {
				args[i] = id
			}
			db.Exec(q, args...)
		}
		http.Redirect(w, r, redirect, http.StatusSeeOther)
	}
}

func detailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		var a article.Article
		err := db.QueryRow("SELECT id, title, content, image_path, video_path FROM articles WHERE id=?", id).Scan(&a.ID, &a.Title, &a.Content, &a.ImagePath, &a.VideoPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		detailTmpl.Execute(w, a)
	}
}

var detailTmpl = template.Must(template.New("detail").Parse(`
<html><body>
<h1>{{.Title}}</h1>
<p>{{.Content}}</p>
{{if .ImagePath}}<img src="/files/{{.ImagePath}}" style="max-width:400px"><br>{{end}}
{{if .VideoPath}}<video controls src="/files/{{.VideoPath}}" style="max-width:400px"></video><br>{{end}}
<a href="/">بازگشت</a>
</body></html>
`))

func parseIDs(r *http.Request) []int {
	r.ParseForm()
	var ids []int
	for _, v := range r.Form["id"] {
		if id, err := strconv.Atoi(v); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

func placeholders(n int) string {
	return strings.TrimRight(strings.Repeat("?,", n), ",")
}
