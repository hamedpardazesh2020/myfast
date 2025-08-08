# MyFast News Bot

این مخزن شامل دو برنامهٔ جداگانه است:

- `scraper` برای جمع‌آوری خبرهای وب‌سایت CBC Local و ذخیره در MySQL
- `poster` برای ارسال خبرهای تایید شده به کانال تلگرام

## راه‌اندازی جدول دیتابیس

```sql
CREATE TABLE articles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255),
    url VARCHAR(255) UNIQUE,
    content TEXT,
    image_path VARCHAR(255),
    video_path VARCHAR(255),
    approved TINYINT(1) DEFAULT 0,
    posted TINYINT(1) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## اجرای Scraper

```bash
export MYSQL_DSN="user:pass@tcp(localhost:3306)/newsdb?parseTime=true"
go run ./cmd/scraper
```
این برنامه هر پنج دقیقه صفحه را بررسی می‌کند و خبرهای جدید را به همراه متن، تصویر و ویدیو ذخیره می‌کند.

## اجرای Poster

```bash
export MYSQL_DSN="user:pass@tcp(localhost:3306)/newsdb?parseTime=true"
export TELEGRAM_TOKEN="توکن ربات"
export TELEGRAM_CHAT="شناسه کانال یا چت"
go run ./cmd/poster
```
این برنامه هر ده دقیقه خبرهای تایید شده که هنوز منتشر نشده‌اند را به تلگرام می‌فرستد و فیلد `posted` را به‌روزرسانی می‌کند.

