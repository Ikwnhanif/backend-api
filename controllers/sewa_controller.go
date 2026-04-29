package controllers

import (
	"mie-supplier-api/config"
	"mie-supplier-api/models"

	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// MASTER KATALOG SEWA (Gerobak, Mangkok, dll)
// ==========================================

func GetKatalogSewa(c *fiber.Ctx) error {
	var katalog []models.KatalogSewa
	
	// SEBELUMNYA: config.DB.Where("is_active = ?", true).Find(&katalog)
	// UBAH MENJADI SEPERTI DI BAWAH INI (Ambil semua data tanpa filter):
	config.DB.Find(&katalog)
	
	return c.JSON(katalog)
}

func AddKatalogSewa(c *fiber.Ctx) error {
	katalog := new(models.KatalogSewa)
	if err := c.BodyParser(katalog); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data salah"})
	}

	if err := config.DB.Create(katalog).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan katalog"})
	}
	return c.Status(201).JSON(katalog)
}

// ==========================================
// MANAJEMEN IZIN MITRA (Pause Tagihan)
// ==========================================

func AddIzinMitra(c *fiber.Ctx) error {
	izin := new(models.IzinMitra)
	if err := c.BodyParser(izin); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format tanggal salah"})
	}

	if err := config.DB.Create(izin).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan izin"})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Izin berhasil dicatat, tagihan sewa akan di-pause pada tanggal tersebut", "data": izin})
}

// ==========================================
// ASSIGN BARANG KE MITRA (Peminjaman)
// ==========================================

func RentAssetToMitra(c *fiber.Ctx) error {
	sewa := new(models.SewaMitra)
	if err := c.BodyParser(sewa); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data salah"})
	}

	if err := config.DB.Create(sewa).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memproses peminjaman"})
	}
	return c.Status(201).JSON(sewa)
}

// ... (Kode sebelumnya tetap ada di sini) ...

// ==========================================
// EDIT & DELETE KATALOG SEWA
// ==========================================

// EditKatalogSewa mengubah nama barang atau harga sewanya
func EditKatalogSewa(c *fiber.Ctx) error {
	id := c.Params("id")
	var katalog models.KatalogSewa

	if err := config.DB.First(&katalog, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Barang tidak ditemukan"})
	}

	if err := c.BodyParser(&katalog); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data salah"})
	}

	config.DB.Save(&katalog)
	return c.JSON(katalog)
}

// DeleteKatalogSewa menonaktifkan barang (Soft Delete) agar riwayat lama tidak error
func DeleteKatalogSewa(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Kita tidak benar-benar menghapus barisnya, kita hanya set IsActive jadi false
	if err := config.DB.Model(&models.KatalogSewa{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menonaktifkan barang"})
	}
	
	return c.JSON(fiber.Map{"message": "Barang berhasil dinonaktifkan dari katalog"})
}

// ==========================================
// EDIT & DELETE IZIN MITRA
// ==========================================

func EditIzinMitra(c *fiber.Ctx) error {
	id := c.Params("id")
	var izin models.IzinMitra

	if err := config.DB.First(&izin, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Data izin tidak ditemukan"})
	}

	if err := c.BodyParser(&izin); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data salah"})
	}

	config.DB.Save(&izin)
	return c.JSON(izin)
}

func DeleteIzinMitra(c *fiber.Ctx) error {
	id := c.Params("id")
	
	// Hapus permanen karena data izin yang salah tidak perlu disimpan
	if err := config.DB.Delete(&models.IzinMitra{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menghapus izin"})
	}
	
	return c.JSON(fiber.Map{"message": "Data izin berhasil dihapus, tagihan sewa akan kembali normal untuk tanggal tersebut"})
}

// ==========================================
// PENGEMBALIAN BARANG SEWA (RETURN ASSET)
// ==========================================

func ReturnAssetToAdmin(c *fiber.Ctx) error {
	idTransaksiSewa := c.Params("id")

	var sewa models.SewaMitra
	if err := config.DB.First(&sewa, idTransaksiSewa).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Data peminjaman tidak ditemukan"})
	}

	type ReturnPayload struct {
		TanggalKembali string `json:"tanggal_kembali"` // RFC3339 string
	}
	var payload ReturnPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data salah"})
	}

	// 1. Parsing string menjadi tipe time.Time milik Go
	parsedTime, err := time.Parse(time.RFC3339, payload.TanggalKembali)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format tanggal gagal dibaca"})
	}

	// 2. Assign ke struct dan simpan secara aman
	sewa.TanggalKembali = &parsedTime
	sewa.IsActive = false

	if err := config.DB.Save(&sewa).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan pengembalian ke database"})
	}

	return c.JSON(fiber.Map{
		"message": "Barang berhasil dikembalikan. Argometer sewa telah dihentikan.",
	})
}

// Mengambil daftar barang yang sedang aktif dipinjam
func GetListSewa(c *fiber.Ctx) error {
	var sewa []models.SewaMitra
	// Preload digunakan agar data nama mitra dan nama barang ikut terbawa ke frontend
	config.DB.Preload("Penjual").Preload("KatalogSewa").Where("is_active = ?", true).Find(&sewa)
	return c.JSON(sewa)
}

// Mengambil daftar izin/cuti mitra
func GetIzinMitra(c *fiber.Ctx) error {
	var izin []models.IzinMitra
	// Preload Penjual agar nama mitra ikut terkirim ke frontend
	if err := config.DB.Preload("Penjual").Order("created_at desc").Find(&izin).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal memuat data izin"})
	}
	return c.JSON(izin)
}

// Menghitung tagihan sewa mitra untuk bulan tertentu
func GetInvoicePreview(c *fiber.Ctx) error {
	bulan := c.QueryInt("bulan")
	tahun := c.QueryInt("tahun")

	if bulan == 0 || tahun == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Bulan dan Tahun harus diisi"})
	}

	// 1. Tentukan batas awal dan akhir bulan yang dicari
	startOfMonth := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1) // Tanggal terakhir di bulan tersebut

	var penjualList []models.Penjual
	config.DB.Where("is_active = ?", true).Find(&penjualList)

	type InvoiceResponse struct {
		PenjualID       uint    `json:"penjual_id"`
		NamaPenjual     string  `json:"nama_penjual"`
		NamaWarung      string  `json:"nama_warung"`
		TotalSewaNormal float64 `json:"total_sewa_normal"`
		PotonganIzin    float64 `json:"potongan_izin"`
		GrandTotal      float64 `json:"grand_total"`
		ItemDetails     string  `json:"item_details"`
	}

	results := []InvoiceResponse{}

	for _, p := range penjualList {
		var assets []models.SewaMitra
		config.DB.Preload("KatalogSewa").Where("penjual_id = ? AND is_active = ?", p.ID, true).Find(&assets)

		if len(assets) == 0 { continue }

		var totalNormal, totalPotongan float64
		var itemNames []string

		// Hitung jumlah hari dalam bulan tersebut (bisa 28, 29, 30, atau 31)
		hariDalamBulan := float64(endOfMonth.Day())

		for _, asset := range assets {
			totalNormal += asset.KatalogSewa.HargaHari * hariDalamBulan
			itemNames = append(itemNames, asset.KatalogSewa.NamaAset)

			// 2. Ambil semua izin yang BERIRISAN dengan bulan ini
			var izins []models.IzinMitra
			config.DB.Where("penjual_id = ? AND tanggal_mulai <= ? AND tanggal_akhir >= ?", 
				p.ID, endOfMonth, startOfMonth).Find(&izins)

			for _, iz := range izins {
				// 3. Tentukan titik potong izin terhadap bulan ini
				// Ambil mana yang lebih akhir antara (Awal Bulan vs Awal Izin)
				hitungMulai := startOfMonth
				if iz.TanggalMulai.After(startOfMonth) {
					hitungMulai = iz.TanggalMulai
				}

				// Ambil mana yang lebih awal antara (Akhir Bulan vs Akhir Izin)
				hitungAkhir := endOfMonth
				if iz.TanggalAkhir.Before(endOfMonth) {
					hitungAkhir = iz.TanggalAkhir
				}

				// Hitung selisih harinya
				selisih := hitungAkhir.Sub(hitungMulai).Hours() / 24
				hariTerpotong := selisih + 1 // Inklusif

				if hariTerpotong > 0 {
					totalPotongan += asset.KatalogSewa.HargaHari * hariTerpotong
				}
			}
		}

		results = append(results, InvoiceResponse{
			PenjualID:       p.ID,
			NamaPenjual:     p.NamaPenjual,
			NamaWarung:      p.NamaWarung,
			TotalSewaNormal: totalNormal,
			PotonganIzin:    totalPotongan,
			GrandTotal:      totalNormal - totalPotongan,
			ItemDetails:     strings.Join(itemNames, ", "),
		})
	}

	return c.JSON(results)
}