# CleanNest: Sistem Manajemen Laundry (Kiloan & Satuan)

Backend REST API untuk usaha laundry, dibangun dengan **Go** dan **Clean Architecture** (`domain → usecase → repository → delivery`, tanpa controller/handler).

CleanNest mendukung tiga jenis layanan dalam satu pesanan:

| Kategori | Contoh | Satuan harga |
|---|---|---|
| Kiloan | Cuci setrika reguler, cuci kering lipat, express | per **kg** (dengan minimum berat) |
| Satuan | Bed cover, jas / dry clean, selimut | per **pcs** |
| Khusus | Cuci sepatu, karpet, gorden | per **pcs** atau **m²** |

## Role

| Role | Hak akses |
|---|---|
| **Admin** | Kelola kasir, kelola layanan & harga, monitoring pesanan, laporan pendapatan harian/bulanan |
| **Kasir** | Daftarkan pelanggan walk-in, buat pesanan, ubah status, catat pembayaran (tunai/transfer/QRIS), batalkan pesanan |
| **Pelanggan** | Registrasi & login, riwayat pesanan sendiri, lacak cucian via kode pesanan |

## Aturan bisnis utama
- `subtotal = max(quantity, min_qty) × price_per_unit`; harga & satuan di-*snapshot* ke `order_items`.
- Alur status: `diterima → diproses → siap_diambil → selesai` (`dibatalkan` dari `diterima`/`diproses`).
- Pesanan tidak bisa `selesai` sebelum `lunas`; pembayaran bertahap didukung, kelebihan bayar ditolak.
- Pelanggan hanya bisa melihat detail pesanan miliknya sendiri.

## Struktur proyek

```text
cmd/main.go                 # entry point & dependency injection
config/                     # konfigurasi environment
pkg/security/               # hash password PBKDF2-SHA256
migrations/                 # 000001 schema (7 tabel), 000002 seed data
internal/domain/            # entitas, error, interface repository
internal/usecase/           # logika bisnis
internal/repository/        # PostgreSQL & in-memory
internal/delivery/          # router + *_route.go per fitur, response, middleware JWT
docs/                       # laporan rancangan sistem
```

## Database (7 tabel)
`users`, `customers`, `services` (category, unit, price_per_unit, min_qty), `orders`, `order_items` (quantity, unit, price_per_unit, subtotal), `payments`, `order_status_logs`. ERD dan kamus data lengkap ada di [`docs/LAPORAN_RANCANGAN_SISTEM.md`](docs/LAPORAN_RANCANGAN_SISTEM.md).

## Menjalankan

Kebutuhan: Go 1.22+. PostgreSQL bersifat opsional.

```bash
# Mode in-memory (data contoh otomatis)
go run ./cmd

# Mode PostgreSQL
createdb cleannest_db
psql -d cleannest_db -f migrations/000001_init_schema.up.sql
psql -d cleannest_db -f migrations/000002_seed_data.up.sql
USE_SQL_DB=true DB_NAME=cleannest_db go run ./cmd

# Test
go test ./...
```

Server berjalan di `http://localhost:8080`. Akun default: `admin@cleannest.com` / `password123`, `kasir@cleannest.com` / `password123`.

## Endpoint ringkas

| Method | Endpoint | Akses |
|---|---|---|
| GET | `/api/v1/health`, `/api/v1/services`, `/api/v1/orders/track?code=` | Publik |
| POST | `/api/v1/auth/register`, `/api/v1/auth/login` | Publik |
| POST/GET/DELETE | `/api/v1/admin/cashiers[/{id}]` | Admin |
| POST/PUT/DELETE | `/api/v1/admin/services[/{id}]` | Admin |
| GET | `/api/v1/admin/reports/daily`, `/monthly` | Admin |
| POST/GET | `/api/v1/customers[/{id}]` | Admin, Kasir |
| POST/GET | `/api/v1/orders` | Admin, Kasir |
| PATCH | `/api/v1/orders/{id}/status` | Admin, Kasir |
| POST | `/api/v1/orders/{id}/cancel` | Admin, Kasir |
| POST/GET | `/api/v1/payments`, `/api/v1/payments/order/{orderId}` | Admin, Kasir |
| GET | `/api/v1/customer/orders` | Pelanggan |
| GET | `/api/v1/orders/{id}` | Login (pelanggan hanya miliknya) |

Contoh membuat pesanan campuran:

```json
POST /api/v1/orders
{
  "customer_id": 1,
  "items": [
    { "service_id": 1, "quantity": 2 },
    { "service_id": 4, "quantity": 1 }
  ]
}
```

## Tim (Kelompok 3 – TI-C1)

| Anggota | Peran | Branch |
|---|---|---|
| Angelyna Rahmah Ramadhani (434241006) | Database & Domain | `feat/domain-and-repo` |
| Nasywa Ashilah Fairuz Zahra (434241010) | Usecase / Business Logic | `feat/usecase-logic` |
| Fanny Rahmadinata (434241009) | Delivery & Git Master | `feat/delivery-and-config` |
