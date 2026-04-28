package main

import (
	"log"
	"mie-supplier-api/common"
	"mie-supplier-api/config"
	"mie-supplier-api/controllers"
	"mie-supplier-api/models"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment Docker")
	}

	// 2. Koneksi DB & Migrasi
	config.ConnectDB()

	// 3. Jalankan Seeder Admin
	seedAdmin()

	// 4. Inisialisasi Fiber
	app := fiber.New(fiber.Config{
		AppName: "Mie Supplier System",
	})

	// 5. Middleware Global
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, https://mie.outsys.space", 
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// 6. Setup Routes
	setupRoutes(app)

	// 7. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Auth (Public)
	api.Post("/login", controllers.Login)

	// ==========================================
	// --- JALUR KASIR ---
	// ==========================================
	// Bisa diakses Kasir & Admin
	kasir := api.Group("/kasir", common.AuthRequired(""))
	
	// Transaksi
	kasir.Post("/check-in", controllers.ProcessPresensiDanMie)
	
	// [FITUR BARU] Endpoint Verifikasi Mitra via RFID atau PIN
	kasir.Post("/verify-mitra", controllers.VerifyMitra)

	// ==========================================
	// --- JALUR ADMIN ---
	// ==========================================
	// Hanya bisa diakses Admin
	admin := api.Group("/admin", common.AuthRequired("admin"))

	// Master Penjual (CRM)
	admin.Get("/penjual", controllers.GetListPenjual)
	admin.Post("/penjual", controllers.AddPenjual)
	admin.Put("/penjual/:id", controllers.UpdatePenjual)
	
	// [FITUR BARU] Endpoint Pairing Kartu RFID ke Mitra
	admin.Post("/penjual/rfid", controllers.AddRFIDToPenjual)
	
	// [FITUR BARU] Endpoint Lihat Semua RFID Milik Mitra
	admin.Get("/penjual/:penjual_id/rfid", controllers.GetRFIDByPenjual)
	
	// [FITUR BARU] Endpoint Delete RFID (Berdasarkan ID Kartu)
	admin.Delete("/penjual/rfid/:id", controllers.DeleteRFIDFromPenjual)
	
	// [FITUR BARU] Endpoint Delete RFID (Berdasarkan RFID Tag via Query)
	// Contoh: DELETE /api/admin/penjual/rfid?rfid_tag=AB:CD:EF:12:34:56
	admin.Delete("/penjual/rfid", controllers.DeleteRFIDFromPenjual)
	// Reports & Analytics
	admin.Get("/daily-rekap", controllers.GetDailyRekap)
	admin.Get("/ranking", controllers.GetTopPenjual)
	admin.Get("/low-activity", controllers.GetInactivePenjual)
	admin.Get("/export", controllers.ExportTransaksiCSV)
	admin.Get("/history-rekap", controllers.GetHistoryRekap)

	// ==========================================
	// [MODUL BARU] ERP SEWA & ASET
	// ==========================================
	sewaRoutes := admin.Group("/sewa")
	
	// Master Katalog Barang Sewa
	sewaRoutes.Get("/katalog", controllers.GetKatalogSewa)
	sewaRoutes.Post("/katalog", controllers.AddKatalogSewa)
	sewaRoutes.Put("/katalog/:id", controllers.EditKatalogSewa)
	sewaRoutes.Delete("/katalog/:id", controllers.DeleteKatalogSewa)

	// Assign Peminjaman & Pengembalian Aset
	sewaRoutes.Post("/assign", controllers.RentAssetToMitra)
	sewaRoutes.Put("/return/:id", controllers.ReturnAssetToAdmin)
	sewaRoutes.Get("/assign", controllers.GetListSewa)

	// Manajemen Izin (Pause Tagihan)
	sewaRoutes.Get("/izin", controllers.GetIzinMitra)
	sewaRoutes.Post("/izin", controllers.AddIzinMitra)
	sewaRoutes.Put("/izin/:id", controllers.EditIzinMitra)
	sewaRoutes.Delete("/izin/:id", controllers.DeleteIzinMitra)

	sewaRoutes.Get("/invoice-preview", controllers.GetInvoicePreview)

}

func seedAdmin() {
	var count int64
	config.DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		// Gunakan Bcrypt untuk password production
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

		admin := models.User{
			Username: "admin",
			Password: string(hashedPassword),
			Role:     "admin",
		}
		config.DB.Create(&admin)

		// Buat kasir default juga untuk testing
		hashedPasswordKasir, _ := bcrypt.GenerateFromPassword([]byte("kasir123"), bcrypt.DefaultCost)
		kasir := models.User{
			Username: "kasir",
			Password: string(hashedPasswordKasir),
			Role:     "kasir",
		}
		config.DB.Create(&kasir)

		log.Println("✅ Seed: Admin (admin123) & Kasir (kasir123) berhasil dibuat")
	}
}