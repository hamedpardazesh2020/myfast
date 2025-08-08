# myfast

این مخزن توضیح می‌دهد چگونه می‌توان با استفاده از زبان Go اخباری را از سایت [CBC Local News](https://www.cbc.ca/news/local) جمع‌آوری و آن‌ها را در کانال تلگرام منتشر کرد. در حال حاضر کد برنامه در این مخزن قرار ندارد و تنها راهنمای کلی برای پیاده‌سازی ارائه شده است.

## ساختار پیشنهادی

برای رسیدن به این هدف می‌توان سه برنامهٔ جداگانه طراحی کرد:

1. **scraper**: هر پنج دقیقه صفحهٔ اخبار را بررسی می‌کند، خبرهای جدید را واکشی کرده و در پایگاه دادهٔ MySQL ذخیره می‌کند. تشخیص خبر جدید بر اساس یکتایی ستون `url` انجام می‌شود.
2. **poster**: هر ده دقیقه خبرهای تأییدشده و منتشرنشده را از دیتابیس می‌خواند و از طریق ربات تلگرام در کانال مشخص شده منتشر می‌کند. برای جلوگیری از ارسال تکراری، ستون `posted` در جدول استفاده می‌شود.
3. **admin**: یک پنل وب برای مدیریت خبرها و تنظیم مشخصات ربات تلگرام شامل توکن و شناسهٔ کانال. در این پنل می‌توان خبرها را تأیید یا حذف کرد و وضعیت انتشار آن‌ها را دید.

## نمونه ساختار جدول‌ها

```sql
CREATE TABLE articles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255),
    url VARCHAR(255) UNIQUE,
    content TEXT,
    image_path VARCHAR(255),
    video_path VARCHAR(255),
    approved TINYINT DEFAULT 0,
    posted TINYINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE settings (
    `key`   VARCHAR(64) PRIMARY KEY,
    `value` VARCHAR(255)
);
```

## متغیرهای محیطی نمونه

```bash
export MYSQL_DSN="user:password@tcp(localhost:3306)/newsdb?charset=utf8mb4&parseTime=True"
export TELEGRAM_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
export TELEGRAM_CHANNEL="@your_channel"
```

## نکات پیاده‌سازی

- برای پارس HTML می‌توان از بستهٔ [`goquery`](https://github.com/PuerkitoBio/goquery) استفاده کرد.
- مدیریت اتصال به MySQL از طریق درایور [`go-sql-driver/mysql`](https://github.com/go-sql-driver/mysql) انجام می‌شود.
- توصیه می‌شود برای دانلود فایل‌های چندرسانه‌ای مسیر مناسبی روی دیسک در نظر گرفته و نام فایل‌ها را بر اساس شناسهٔ خبر ذخیره کنید.
- پیش از جمع‌آوری اطلاعات از هر وب‌سایت، قوانین مربوط به استفاده از داده و محدودیت‌های دسترسی (مانند robots.txt) را بررسی کنید.

این README تنها به‌عنوان راهنما ارائه شده است و پیاده‌سازی عملی باید بر اساس نیازهای دقیق پروژه انجام شود.
