# myfast

این پروژه شامل اسکریپتی به زبان Go است که هر پنج دقیقه صفحه‌ی [CBC Local News](https://www.cbc.ca/news/local) را بررسی کرده و اطلاعات خبرهای جدید را در پایگاه داده‌ی MySQL ذخیره می‌کند. همچنین اخبار تأیید شده را هر ده دقیقه از طریق ربات تلگرام در کانال مشخص شده منتشر می‌کند.

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
```

2. به‌روزرسانی اطلاعات اتصال به دیتابیس در `main.go` و تنظیم متغیرهای محیطی ربات:

```bash
export TELEGRAM_BOT_TOKEN="توکن_بات"
export TELEGRAM_CHANNEL="@کانال_شما"
```

3. اجرای برنامه:

```bash
go run main.go
```

برنامه به صورت مداوم اجرا شده و هر پنج دقیقه اخبار جدید را ذخیره می‌کند. برای انتشار در تلگرام لازم است سطرهای موردنظر در جدول با مقدار `approved` برابر ۱ تنظیم شوند؛ برنامه هر ده دقیقه موارد تأیید شده و منتشر نشده را به کانال ارسال می‌کند.
