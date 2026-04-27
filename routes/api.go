package routes

import (
	"mie-supplier-api/common" // Import folder middleware tadi
	"mie-supplier-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Public
	api.Post("/login", controllers.Login)

	// Kasir Group (Hanya Kasir & Admin yang bisa akses)
	kasir := api.Group("/kasir", common.AuthRequired("")) 
	kasir.Post("/check-in", controllers.ProcessPresensiDanMie)

	// Admin Group (Hanya Admin yang bisa akses)
	admin := api.Group("/admin", common.AuthRequired("admin"))
	admin.Get("/penjual", controllers.GetListPenjual)
	admin.Post("/penjual", controllers.AddPenjual)
	admin.Get("/daily-rekap", controllers.GetDailyRekap)
	admin.Get("/ranking", controllers.GetTopPenjual)
	admin.Get("/low-activity", controllers.GetInactivePenjual)
	admin.Get("/export", controllers.ExportTransaksiCSV)
	admin.Get("/history-rekap", controllers.GetHistoryRekap)
}