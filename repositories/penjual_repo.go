package repositories

import (
	"mie-supplier-api/config"
	"mie-supplier-api/models"
)

func CreatePenjual(p *models.Penjual) error {
	return config.DB.Create(p).Error
}

func GetAllPenjual() ([]models.Penjual, error) {
	var penjual []models.Penjual
	err := config.DB.Find(&penjual).Error
	return penjual, err
}

func GetPenjualByPin(pin string) (models.Penjual, error) {
	var p models.Penjual
	err := config.DB.Where("pin = ?", pin).First(&p).Error
	return p, err
}