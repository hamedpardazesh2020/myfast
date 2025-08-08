# myfast

این پروژه شامل چند برنامهٔ جداگانه به زبان Go است:

- **scraper** هر پنج دقیقه صفحهٔ [CBC Local News](https://www.cbc.ca/news/local) را بررسی کرده و اخبار جدید را در پایگاه دادهٔ MySQL ذخیره می‌کند.
- **poster** هر ده دقیقه اخبار تأییدشده و منتشرنشده را از دیتابیس خوانده و از طریق ربات تلگرام در کانالی که در پنل مدیریتی تنظیم شده ارسال می‌کند.
- **admin** یک پنل وب ساده برای مدیریت اخبار و تنظیم اطلاعات ربات تلگرام است.

## اجرا

1. ایجاد جدول در پایگاه داده با ساختار مناسب. جدول نمونه:

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

2. تنظیم متغیر محیطی اتصال دیتابیس:

```bash
export MYSQL_DSN="user:password@tcp(localhost:3306)/newsdb?charset=utf8mb4&parseTime=True"
```

3. اجرای پنل مدیریت برای تنظیم توکن و مدیریت خبرها:

```bash
go run cmd/admin/main.go -port 8080
```

4. اجرای سرویس جمع‌آوری خبر:

```bash
go run cmd/scraper/main.go
```

5. اجرای سرویس ارسال به تلگرام:

```bash
go run cmd/poster/main.go
```

برنامهٔ اول به صورت مداوم اجرا شده و هر پنج دقیقه اخبار جدید را ذخیره می‌کند. برای انتشار در تلگرام لازم است سطرهای موردنظر در جدول با مقدار `approved` برابر ۱ تنظیم شوند؛ برنامهٔ دوم هر ده دقیقه موارد تأیید شده و منتشر نشده را به کانال ارسال می‌کند.

برای تشخیص خبر جدید، برنامه با استفاده از ستون `url` (که به صورت `UNIQUE` تعریف شده) بررسی می‌کند که رکوردی با همان آدرس قبلاً در دیتابیس ثبت نشده باشد.
