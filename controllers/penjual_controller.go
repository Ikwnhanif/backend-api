package controllers

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"mie-supplier-api/config" // Sesuaikan dengan path folder config database Anda
	"mie-supplier-api/models"

	"github.com/gofiber/fiber/v2"
)

// ==========================================
// AREA ADMIN: MANAJEMEN DATA MITRA
// ==========================================

func AddPenjual(c *fiber.Ctx) error {
	p := new(models.Penjual)
	if err := c.BodyParser(p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Fitur Lama Anda: Generate PIN Otomatis jika kosong
	if p.Pin == "" {
		rand.Seed(time.Now().UnixNano())
		p.Pin = fmt.Sprintf("%04d", rand.Intn(10000))
	}

	// Default selalu aktif saat pertama dibuat
	p.IsActive = true

	// Simpan ke database (Menggunakan config.DB agar langsung jalan)
	if err := config.DB.Create(p).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal simpan data atau PIN/Nama duplikat"})
	}

	return c.Status(201).JSON(p)
}

func GetListPenjual(c *fiber.Ctx) error {
	var penjuals []models.Penjual
	
	// Preload("Cards") berfungsi agar saat admin melihat daftar mitra, 
	// data kartu RFID yang terhubung ke masing-masing mitra ikut tampil.
	config.DB.Preload("Cards").Find(&penjuals)
	
	return c.JSON(penjuals)
}

// Tambah Kartu RFID Baru ke Mitra yang sudah ada
func AddRFIDToPenjual(c *fiber.Ctx) error {
	card := new(models.PenjualRFID)
	if err := c.BodyParser(card); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data kartu salah"})
	}

	// Simpan ke tabel PenjualRFID
	if err := config.DB.Create(card).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal menyimpan, mungkin Tag RFID sudah dipakai"})
	}

	return c.Status(201).JSON(card)
}

func DeleteRFIDFromPenjual(c *fiber.Ctx) error {
	// Bisa hapus berdasarkan ID kartu RFID atau berdasarkan RFID tag
	id := c.Params("id") // ID dari tabel PenjualRFID
	rfidTag := c.Query("rfid_tag") // Alternatif: hapus berdasarkan tag RFID

	var card models.PenjualRFID

	// Jika ada parameter rfid_tag, cari berdasarkan tag
	if rfidTag != "" {
		if err := config.DB.Where("rfid_tag = ?", rfidTag).First(&card).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Kartu RFID tidak ditemukan",
				"detail": fmt.Sprintf("Tidak ada kartu dengan tag: %s", rfidTag),
			})
		}
	} else if id != "" {
		// Jika tidak ada rfid_tag, cari berdasarkan ID
		if err := config.DB.First(&card, id).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "Kartu RFID tidak ditemukan",
				"detail": fmt.Sprintf("Tidak ada kartu dengan ID: %s", id),
			})
		}
	} else {
		return c.Status(400).JSON(fiber.Map{
			"error": "Parameter tidak valid",
			"detail": "Gunakan ID kartu di URL atau query parameter ?rfid_tag=xxx",
		})
	}

	// Ambil data penjual sebelum dihapus untuk response
	var penjual models.Penjual
	config.DB.First(&penjual, card.PenjualID)

	// Hapus kartu RFID
	if err := config.DB.Delete(&card).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menghapus kartu RFID",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Kartu RFID berhasil dihapus",
		"deleted_card": fiber.Map{
			"id":        card.ID,
			"rfid_tag":  card.RFIDTag,
			"penjual_id": card.PenjualID,
			"nama_penjual": penjual.NamaPenjual,
		},
	})
}

// GetRFIDByPenjual untuk melihat semua kartu RFID milik mitra tertentu
func GetRFIDByPenjual(c *fiber.Ctx) error {
	penjualID := c.Params("penjual_id")
	
	// Cek apakah mitra ada
	var penjual models.Penjual
	if err := config.DB.First(&penjual, penjualID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Mitra tidak ditemukan"})
	}

	// Ambil semua kartu RFID milik mitra
	var cards []models.PenjualRFID
	config.DB.Where("penjual_id = ?", penjualID).Find(&cards)

	return c.JSON(fiber.Map{
		"penjual": penjual,
		"cards":   cards,
		"total":   len(cards),
	})
}

// ==========================================
// AREA KASIR: VERIFIKASI TRANSAKSI
// ==========================================

func VerifyMitra(c *fiber.Ctx) error {
	type AuthRequest struct {
		Type  string `json:"type"`  // rfid, pin, atau search
		Value string `json:"value"` // kode rfid, angka pin, atau potongan nama/alamat
	}

	req := new(AuthRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request salah"})
	}

	cleanValue := strings.TrimSpace(req.Value)
	var penjual models.Penjual

	switch req.Type {
	case "rfid":
		var rfidEntry models.PenjualRFID
		err := config.DB.Preload("Penjual").Where("rfid_tag = ?", cleanValue).First(&rfidEntry).Error
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Kartu RFID tidak terdaftar"})
		}
		penjual = rfidEntry.Penjual

	case "pin":
		err := config.DB.Where("pin = ?", cleanValue).First(&penjual).Error
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "PIN salah atau tidak ditemukan"})
		}

	case "search":
		// Mencari mitra berdasarkan Nama ATAU Alamat
		// Menggunakan ILIKE (Postgres) atau LIKE (MySQL) untuk pencarian parsial
		query := "%" + cleanValue + "%"
		err := config.DB.Where("(nama_penjual LIKE ? OR alamat_jualan LIKE ?) AND is_active = ?", query, query, true).First(&penjual).Error
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Mitra dengan nama/alamat tersebut tidak ditemukan"})
		}

	default:
		return c.Status(400).JSON(fiber.Map{"error": "Tipe verifikasi tidak valid"})
	}

	if !penjual.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Mitra ini sedang dinonaktifkan"})
	}

	return c.JSON(fiber.Map{
		"message": "Verifikasi Sukses",
		"data":    penjual,
	})
}

// UpdatePenjual untuk mengedit data mitra yang sudah ada
func UpdatePenjual(c *fiber.Ctx) error {
	id := c.Params("id")
	var penjual models.Penjual

	// Cari data berdasarkan ID
	if err := config.DB.First(&penjual, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Data mitra tidak ditemukan"})
	}

	// Timpa data lama dengan data baru dari form
	if err := c.BodyParser(&penjual); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format data salah"})
	}

	// Simpan perubahan ke database
	config.DB.Save(&penjual)

	return c.JSON(penjual)
}

// Tambahkan fungsi baru ini
func GetListPenjualAktif(c *fiber.Ctx) error {
	var penjuals []models.Penjual
	
	// Kasir HANYA butuh melihat penjual yang aktif dan tidak butuh data detail (seperti omset)
	// Kita ambil field yang penting saja (opsional) atau ambil semua yang aktif
	err := config.DB.Where("is_active = ?", true).Find(&penjuals).Error
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil data mitra"})
	}
	
	return c.JSON(penjuals)
}

// GetRiwayatKasir - Mengambil data transaksi hari ini (Terbaru di atas)
func GetRiwayatKasir(c *fiber.Ctx) error {
	var riwayat []models.TransaksiMie

	// Ambil tanggal hari ini (format YYYY-MM-DD)
	today := time.Now().Format("2006-01-02")

	// Preload "Penjual" agar kita dapat nama mitranya
	// Filter berdasarkan hari ini, urutkan dari jam terbaru (desc)
	err := config.DB.Preload("Penjual").
		Where("DATE(tanggal_transaksi) = ?", today).
		Order("tanggal_transaksi desc").
		Limit(50).
		Find(&riwayat).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal mengambil riwayat transaksi"})
	}

	return c.JSON(riwayat)
}