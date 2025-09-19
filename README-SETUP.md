Currency Converter API — Quick Start

API แปลงสกุลเงินด้วย Go + Gin + JWT (single-session login) ดึงเรตจาก Frankfurter (ฟรี/ไม่ต้องมี API key)

เครื่องมือที่ต้องมี ติดตั้ง
Docker Desktop


```` bash
git clone <your-repo-url> currency-converter-api
cd currency-converter-api
````

setup  .env
````
APP_ENV=local
APP_PORT=8080
JWT_SECRET=secretexchangecurrency

DB_HOST=ep-spring-bar-a1l2oikr-pooler.ap-southeast-1.aws.neon.tech
DB_PORT=5432
DB_USER=neondb_owner
DB_PASSWORD=npg_UFKD8nydJ3iT
DB_NAME=neondb
DB_SSLMODE=require

# Rate provider (frankfurter or exchangeratehost)
RATES_PROVIDER=frankfurter
RATES_REFRESH_INTERVAL_HOURS=6
````

```` bash
docker compose -f docker-compose.neon.yml --env-file .env build --no-cache --progress=plain
docker compose -f docker-compose.neon.yml --env-file .env up
````

````
test api 
curl http://localhost:8080/check
````

api ใช้งานอยู่ใน file
```Currency Converter API.postman_collection.json```


```
Database PostgreSQL ใช้ตัวฟรีของ neon เป็น cloud database 
```

```
api convert currency ที่ใช้ Frankfurter
สกุลเงินที่ครอบคลุม (ภาพรวม)
Frankfurter อิงชุดสกุลที่ ECB เผยแพร่ ซึ่งครอบคลุม “เมเจอร์” และ “ภูมิภาคหลัก” ค่อนข้างครบ ตัวอย่างเช่น:
เมเจอร์: USD, JPY, GBP, CHF, AUD, CAD, NZD 

ยุโรป: BGN, CZK, DKK, HUF, PLN, RON, SEK, NOK

เอเชีย: CNY, HKD, SGD, KRW, INR, IDR, MYR, PHP, THB ฯลฯ 

อเมริกา/อื่น ๆ: BRL, MXN, ZAR ฯลฯ 

Frankfurter คือบริการ API ฟรี/โอเพ่นซอร์ส สำหรับอัตราแลกเปลี่ยนที่อ้างอิงจากธนาคารกลาง/หน่วยงานไม่เชิงพาณิชย์ โดยเฉพาะ ECB Euro foreign exchange reference rates

แหล่งข้อมูล & รอบการอัปเดต

อัตราแลกเปลี่ยนมาจาก ECB Reference Rates (อัตราอ้างอิง ไม่ใช่อัตราซื้อขายธุรกรรมจริง) และโดยทั่วไป อัปเดตรอบวันเวลาประมาณ 16:00 CET ในวันทำการ ของยุโรป (TARGET days)
```