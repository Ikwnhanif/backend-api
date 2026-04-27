package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"unique;not null" json:"username"`
	Password  string    `json:"-"` // Tidak akan muncul di JSON response
	Role      string    `gorm:"type:enum('admin', 'kasir');default:'kasir'" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Penjual struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	NamaPenjual string    `gorm:"not null" json:"nama_penjual"`
	NamaWarung  string    `json:"nama_warung"`
	NomorHP     string    `json:"nomor_hp"`
	Pin         string    `gorm:"unique;not null" json:"pin"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
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
	UserID           uint      `json:"user_id"` // Kasir yang bertugas
	JumlahKg         float64   `gorm:"type:decimal(10,2);not null" json:"jumlah_kg"`
	TanggalTransaksi time.Time `gorm:"autoCreateTime" json:"tanggal_transaksi"`
}