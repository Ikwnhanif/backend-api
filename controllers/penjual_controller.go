package controllers

import (
	"fmt"
	"math/rand"
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
		Type  string `json:"type"`  // isinya harus "rfid" atau "pin"
		Value string `json:"value"` // isinya kode tag RFID atau angka PIN
	}

	req := new(AuthRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Format request salah"})
	}

	var penjual models.Penjual

	// CABANG LOGIKA DUAL-AUTH
	if req.Type == "rfid" {
		
		// 1. Kasir pakai Kartu (Tapping)
		var rfidEntry models.PenjualRFID
		// Cari kartu, dan langsung tarik data profil penjualnya (Preload)
		err := config.DB.Preload("Penjual").Where("rfid_tag = ?", req.Value).First(&rfidEntry).Error
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Kartu RFID tidak terdaftar"})
		}
		penjual = rfidEntry.Penjual

	} else if req.Type == "pin" {
		
		// 2. Kasir pakai PIN Manual (Fallback jika kartu tertinggal)
		err := config.DB.Where("pin = ?", req.Value).First(&penjual).Error
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "PIN salah atau tidak ditemukan"})
		}

	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Tipe verifikasi tidak valid"})
	}

	// Validasi Terakhir: Pastikan mitra tidak sedang di-banned / nonaktif
	if !penjual.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "Mitra ini sedang dinonaktifkan. Hubungi Admin."})
	}

	// Sukses! Kembalikan data penjual ke layar kasir untuk lanjut transaksi
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