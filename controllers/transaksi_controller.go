package controllers

import (
	"mie-supplier-api/config"
	"mie-supplier-api/models"
	"mie-supplier-api/repositories"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AttendanceRequest struct {
	Pin      string  `json:"pin"`
	JumlahKg float64 `json:"jumlah_kg"`
	UserID   uint    `json:"user_id"` // ID Kasir
}

func ProcessPresensiDanMie(c *fiber.Ctx) error {
	req := new(AttendanceRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Cari Penjual berdasarkan PIN
	penjual, err := repositories.GetPenjualByPin(req.Pin)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "PIN tidak terdaftar"})
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	// 2. Cek apakah sudah presensi hari ini
	var existing models.Presensi
	config.DB.Where("penjual_id = ? AND tanggal = ?", penjual.ID, today).First(&existing)

	// Mulai Transaksi Database (agar konsisten)
	tx := config.DB.Begin()

	if existing.ID == 0 {
		// Jika belum hadir, buat record presensi
		presensi := models.Presensi{
			PenjualID: penjual.ID,
			Tanggal:   now,
			JamMasuk:  now,
		}
		if err := tx.Create(&presensi).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Gagal catat presensi"})
		}
	}

	// 3. Catat Transaksi Mie
	if req.JumlahKg > 0 {
		transaksi := models.TransaksiMie{
			PenjualID: penjual.ID,
			UserID:    req.UserID,
			JumlahKg:  req.JumlahKg,
		}
		if err := tx.Create(&transaksi).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Gagal catat pembelian"})
		}
	}

	tx.Commit()
	return c.JSON(fiber.Map{
		"message": "Sukses!",
		"nama":    penjual.NamaPenjual,
		"warung":  penjual.NamaWarung,
	})
}