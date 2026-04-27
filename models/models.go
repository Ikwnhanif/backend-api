package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `json:"-"`
	Role      string    `gorm:"type:enum('admin', 'kasir');default:'kasir'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Penjual struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	NamaPenjual  string         `gorm:"not null" json:"nama_penjual"`
	NamaWarung   string         `json:"nama_warung"`
	AlamatJualan string         `json:"alamat_jualan"`
	AlamatRumah  string         `json:"alamat_rumah"`
	NoWhatsapp   string         `json:"no_whatsapp"` // Menggantikan NomorHP
	Pin          string         `gorm:"unique;not null" json:"pin"` // Tetap ada sebagai fallback
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Lat          float64        `json:"lat"`
	Long         float64        `json:"long"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	// Relasi One-to-Many: 1 Mitra bisa punya banyak kartu RFID
	Cards        []PenjualRFID  `json:"cards" gorm:"foreignKey:PenjualID"`
}

type PenjualRFID struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PenjualID uint      `json:"penjual_id"`
	Penjual   Penjual   `gorm:"foreignKey:PenjualID" json:"-"` 
	
	RFIDTag   string    `gorm:"unique;not null" json:"rfid_tag"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
}

type Presensi struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PenjualID uint      `json:"penjual_id"`
	Penjual   Penjual   `gorm:"foreignKey:PenjualID"`
	Tanggal   time.Time `gorm:"type:date;not null" json:"tanggal"`
	JamMasuk  time.Time `gorm:"type:time;not null" json:"jam_masuk"`
}

type TransaksiMie struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	PenjualID        uint      `json:"penjual_id"`
	Penjual          Penjual   `gorm:"foreignKey:PenjualID"`
	UserID           uint      `json:"user_id"`
	JumlahKg         float64   `gorm:"type:decimal(10,2);not null" json:"jumlah_kg"`
	TanggalTransaksi time.Time `gorm:"autoCreateTime" json:"tanggal_transaksi"`
}