package controllers

import (
	"fmt"
	"math/rand"
	"mie-supplier-api/models"
	"mie-supplier-api/repositories"
	"time"

	"github.com/gofiber/fiber/v2"
)

func AddPenjual(c *fiber.Ctx) error {
	p := new(models.Penjual)
	if err := c.BodyParser(p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Generate PIN Otomatis jika kosong
	if p.Pin == "" {
		rand.Seed(time.Now().UnixNano())
		p.Pin = fmt.Sprintf("%04d", rand.Intn(10000))
	}

	if err := repositories.CreatePenjual(p); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gagal simpan atau PIN duplikat"})
	}

	return c.Status(201).JSON(p)
}

func GetListPenjual(c *fiber.Ctx) error {
	data, _ := repositories.GetAllPenjual()
	return c.JSON(data)
}