package config

import (
	"fmt"
	"log"
	"mie-supplier-api/models"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal koneksi ke database:", err)
	}

	// Auto Migrate
	// Pastikan &models.PenjualRFID{} ada di dalam kurung ini:
	err = db.AutoMigrate(&models.User{}, &models.Penjual{}, &models.PenjualRFID{}, &models.Presensi{}, &models.TransaksiMie{})
	if err != nil {
		log.Fatal("Gagal migrasi database:", err)
	}

	DB = db
	fmt.Println("Database terkoneksi & migrasi berhasil!")
}