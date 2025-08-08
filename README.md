# myfast
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
)