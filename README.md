# PS Rental Monitor

Aplikasi monitoring rental PlayStation dengan Go dan PostgreSQL.

## Fitur
- Admin dapat menambah TV/PS
- Admin dapat mengatur durasi bermain per TV
- Public dapat melihat status TV secara realtime
- Auto-update status ketika waktu habis
- WebSocket untuk update realtime

## Instalasi

1. Install dependencies:
```bash
go mod tidy
```

2. Jalankan aplikasi:
```bash
go run main.go
```

## Penggunaan

- **Admin Panel**: http://localhost:8080/admin
  - Tambah TV baru
  - Mulai/stop sesi bermain
  - Set durasi bermain

- **Public Monitor**: http://localhost:8080/
  - Lihat status semua TV
  - Monitor waktu tersisa secara realtime

## Database

Aplikasi akan otomatis membuat tabel `tvs` di database PostgreSQL yang sudah dikonfigurasi.

## API Endpoints

- `GET /api/tvs` - Get semua TV
- `POST /admin/tv` - Tambah TV baru
- `PUT /admin/tv/:id/start` - Mulai sesi bermain
- `PUT /admin/tv/:id/stop` - Stop sesi bermain
- `GET /ws` - WebSocket untuk realtime updates