# KOHA CLI (Go prototype)

Prototipe menu launcher bergaya terminal untuk Windows, ditulis dengan Go.
Satu `.exe` tanpa dependensi dan tanpa proses background: dipanggil (misalnya
lewat Lenovo Vantage atau shortcut), menampilkan menu, menjalankan satu
aksi, lalu selesai.

Ini turunan dari KOHA versi AutoHotkey (GUI native). Bedanya, menu di sini
dibaca dari file `koha.json`, jadi bisa diubah tanpa compile ulang.

## Build dan jalankan

Butuh Go 1.21 atau lebih baru.

```bash
go build -ldflags="-s -w" -o koha.exe .
./koha.exe
```

`koha.json` dicari di folder yang sama dengan `koha.exe`, bukan di folder
tempat perintah dijalankan.

## Kontrol

| Tombol | Fungsi |
|---|---|
| Panah atas / bawah, `k` / `j` | Pindah pilihan |
| Tab / Shift+Tab | Pilihan berikutnya / sebelumnya |
| Enter | Pilih (masuk submenu atau jalankan aksi) |
| Esc | Kembali satu level, atau keluar dari menu utama |
| Ctrl+C | Keluar |

## Konfigurasi (`koha.json`)

```json
{
  "theme": "amber",
  "items": [
    { "label": "Open WezTerm", "type": "exe", "target": "wezterm-gui" },
    { "label": "Sistem", "children": [
      { "label": "Lock PC", "type": "builtin", "target": "lock" }
    ]}
  ]
}
```

Field tiap item:

| Field | Keterangan |
|---|---|
| `label` | Teks yang tampil di menu |
| `type` | `exe`, `url`, `keys`, `text`, atau `builtin` |
| `target` | Isi aksi, tergantung `type` (lihat bawah) |
| `args` | Argumen tambahan (hanya `exe`) |
| `admin` | `true` = jalankan sebagai Administrator (hanya `exe`) |
| `delay_ms` | Jeda sebelum `keys`/`text` dikirim. Default 250 |
| `children` | Daftar item anak. Kalau ada, item ini jadi submenu |

### Tipe aksi

- **`exe`**: membuka program lewat `ShellExecute`. Mendukung `%VAR%`
  gaya Windows, misalnya `%LOCALAPPDATA%\\Programs\\Obsidian\\Obsidian.exe`.
- **`url`**: membuka alamat di browser default.
- **`keys`**: mengirim kombinasi tombol ke aplikasi yang sebelumnya aktif.
  Token dipisah spasi, contoh `ctrl+a ctrl+c`, `ctrl+shift+t`, `enter`.
  `wait:300` menjeda 300 ms. Tombol yang dikenal: huruf, angka, `f1`-`f12`,
  `enter`, `tab`, `esc`, `space`, `backspace`, `delete`, panah, `home`,
  `end`, `pgup`, `pgdn`. Modifier: `ctrl`, `alt`, `shift`, `win`.
- **`text`**: mengetik teks apa adanya (mendukung Unicode dan baris baru)
  tanpa memakai clipboard.
- **`builtin`**: `lock` (kunci PC) atau `sleep`.

Untuk `keys` dan `text`, menu menutup dulu, menunggu `delay_ms`, baru
mengirim input. Kalau teks tidak masuk ke aplikasi yang benar, naikkan
`delay_ms`.

### Tema

`game_boy`, `amber`, `green_term`, `catppuccin-mocha`, `gruvbox`, `vague`,
`tokyonight`. Warnanya sama dengan KOHA versi AHK. Nama tema yang tidak
dikenal jatuh ke `amber`.

## Struktur kode

| File | Isi |
|---|---|
| `main.go` | Entry point, loop menu, render, pembacaan tombol |
| `config.go` | Struct dan loader `koha.json` |
| `theme.go` | Palet warna dan escape sequence ANSI |
| `console.go` | Mode konsol, resize dan penempatan jendela |
| `actions.go` | Eksekusi aksi: ShellExecute, SendInput, lock, sleep |

## Keterbatasan

- **Hanya Windows.** `console.go` dan `actions.go` memakai Win32 API.
- **Ukuran jendela:** resize hanya berefek di conhost klasik. Di Windows
  Terminal ukuran jendela diatur oleh Terminal, bukan program ini. Untuk
  memaksa conhost, panggil `conhost.exe koha.exe`.
- Menu belum menutup otomatis saat kehilangan fokus.
- Belum ada fitur KOHA AHK lain: Search Apps, Pengeluaran, Reminder, Waktu
  Sholat, dan ganti tema dari dalam menu.
- Tanpa test otomatis. `keys` dan `text` hanya diuji manual.

## Catatan teknis

- Struct `INPUT` untuk `SendInput` harus berukuran 40 byte di x64. Padding
  di `actions.go` ada untuk itu.
- Input terminal memakai mode VT (`ENABLE_VIRTUAL_TERMINAL_INPUT`), jadi
  panah datang sebagai urutan `ESC [ A`. `Esc` dikenali sebagai tombol
  sendiri hanya kalau datang sebagai satu byte.
- Sebelum mengirim `keys`/`text`, program memanggil `FreeConsole` supaya
  jendela menu hilang dan fokus kembali ke aplikasi sebelumnya.
