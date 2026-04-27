package controllers

import (
	"encoding/csv"
	"fmt"
	"mie-supplier-api/config"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Menghitung Ranking Penjual berdasarkan pengambilan mie terbanyak bulan ini
func GetTopPenjual(c *fiber.Ctx) error {
	var results []map[string]interface{}
	
	// PERBAIKAN: Nama tabel diubah menjadi penjuals dan transaksi_mies
	// Ditambahkan juga filter YEAR agar tidak tercampur data bulan yang sama di tahun berbeda
	query := `
		SELECT p.nama_penjual, p.nama_warung, SUM(t.jumlah_kg) as total_kg
		FROM penjuals p
		JOIN transaksi_mies t ON p.id = t.penjual_id
		WHERE MONTH(t.tanggal_transaksi) = MONTH(CURRENT_DATE()) 
		  AND YEAR(t.tanggal_transaksi) = YEAR(CURRENT_DATE())
		GROUP BY p.id
		ORDER BY total_kg DESC
		LIMIT 10
	`
	config.DB.Raw(query).Scan(&results)
	return c.JSON(results)
}

// Identifikasi penjual yang jarang hadir (kurang dari 10 hari dalam sebulan terakhir)
func GetInactivePenjual(c *fiber.Ctx) error {
	var results []map[string]interface{}

	// PERBAIKAN: Nama tabel diubah menjadi penjuals dan presensis
	query := `
		SELECT p.nama_penjual, p.nama_warung, COUNT(pr.id) as total_hadir
		FROM penjuals p
		LEFT JOIN presensis pr ON p.id = pr.penjual_id 
			AND pr.tanggal >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
		GROUP BY p.id
		HAVING total_hadir < 10
		ORDER BY total_hadir ASC
	`
	config.DB.Raw(query).Scan(&results)
	return c.JSON(results)
}

// Rekap Harian untuk Dashboard Admin
func GetDailyRekap(c *fiber.Ctx) error {
	var results []map[string]interface{}

	// PERBAIKAN BESAR:
	// Sekarang query mengembalikan array SEMUA penjual untuk mengisi tabel Live Status di Frontend.
	// COALESCE digunakan agar penjual yang belum ambil mie hari ini tetap muncul dengan jumlah_kg = 0 (Status: Libur)
	query := `
		SELECT p.nama_penjual, p.nama_warung, COALESCE(SUM(t.jumlah_kg), 0) as jumlah_kg
		FROM penjuals p
		LEFT JOIN transaksi_mies t ON p.id = t.penjual_id 
			AND DATE(t.tanggal_transaksi) = CURDATE()
		GROUP BY p.id
		ORDER BY jumlah_kg DESC, p.nama_penjual ASC
	`
	
	config.DB.Raw(query).Scan(&results)
	return c.JSON(results)
}

// Fungsi Export CSV yang sempat hilang
func ExportTransaksiCSV(c *fiber.Ctx) error {
	// 1. Set Header agar browser mendownload sebagai file CSV
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=laporan_transaksi_mie.csv")

	// 2. Inisialisasi CSV Writer
	writer := csv.NewWriter(c)
	defer writer.Flush()

	// 3. Tulis Header Kolom
	writer.Write([]string{"ID Transaksi", "Nama Penjual", "Nama Warung", "Jumlah (KG)", "Waktu"})

	// 4. Ambil Data dari Database dengan Join
	type ExportData struct {
		ID               uint    
		NamaPenjual      string
		NamaWarung       string
		JumlahKg         float64
		TanggalTransaksi time.Time
	}
	var data []ExportData

	// PERBAIKAN: Nama tabel menggunakan "transaksi_mies" dan "penjuals"
	config.DB.Table("transaksi_mies").
		Select("transaksi_mies.id, penjuals.nama_penjual, penjuals.nama_warung, transaksi_mies.jumlah_kg, transaksi_mies.tanggal_transaksi").
		Joins("join penjuals on penjuals.id = transaksi_mies.penjual_id").
		Order("transaksi_mies.tanggal_transaksi DESC").
		Scan(&data)

	// 5. Iterasi data ke dalam baris CSV
	for _, item := range data {
		row := []string{
			fmt.Sprintf("%d", item.ID),
			item.NamaPenjual,
			item.NamaWarung,
			fmt.Sprintf("%.2f", item.JumlahKg),
			item.TanggalTransaksi.Format("2006-01-02 15:04:05"),
		}
		writer.Write(row)
	}

	return nil
}

// ==========================================
// 5. HISTORY REKAP (Berdasarkan Tanggal)
// ==========================================
func GetHistoryRekap(c *fiber.Ctx) error {
	var results []map[string]interface{}
	
	// Ambil parameter tanggal dari URL (misal: ?date=2026-04-25)
	targetDate := c.Query("date")
	if targetDate == "" {
		// Jika kosong, gunakan hari ini sebagai default
		targetDate = time.Now().Format("2006-01-02")
	}

	query := `
		SELECT p.nama_penjual, p.nama_warung, COALESCE(SUM(t.jumlah_kg), 0) as jumlah_kg
		FROM penjuals p
		LEFT JOIN transaksi_mies t ON p.id = t.penjual_id 
			AND DATE(t.tanggal_transaksi) = ?
		GROUP BY p.id
		ORDER BY jumlah_kg DESC, p.nama_penjual ASC
	`
	
	// Inject variabel targetDate ke dalam query ?
	config.DB.Raw(query, targetDate).Scan(&results)
	return c.JSON(results)
}