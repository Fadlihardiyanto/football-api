# ⚽ Football API

API backend untuk sistem manajemen tim sepak bola, lengkap dengan autentikasi, pengelolaan tim, pemain, pertandingan, dan sistem logging menggunakan Kafka.

---

## ⚙️ Setup

1. Clone repositori:
   ```bash
   git clone https://github.com/Fadlihardiyanto/football-api.git
   cd football-api
   ```

2. Copy file `.env.example` ke `.env` dan sesuaikan konfigurasi:
   ```bash
   cp .env.example .env
   ```

3. Unduh dependensi:
   ```bash
   go mod tidy
   ```

4. Build project:
   ```bash
   make build
   ```

5. Jalankan migrasi database:
   ```bash
   make migrate-up
   ```

---

## 🚀 Menjalankan Aplikasi

- Jalankan Web Server:
  ```bash
  make run-web
  ```

- Jalankan Worker (untuk Kafka consumer, log, dll.):
  ```bash
  make run-worker
  ```

- Atau jalankan keduanya sekaligus:
  ```bash
  make run-app
  ```

---

## 🧪 API Testing (Postman)

Import koleksi Postman berikut untuk mencoba semua endpoint:

🔗 [Link Postman Collection](https://www.postman.com/crimson-equinox-265805/api-public)

---

> Pastikan Web dan Worker service sudah aktif sebelum mengakses API melalui Postman.