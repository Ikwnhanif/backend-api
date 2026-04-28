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

// ==========================================
// MODUL SEWA & ASET MITRA
// ==========================================

// 1. Master Barang Sewa (Katalog)
type KatalogSewa struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NamaAset  string    `gorm:"not null" json:"nama_aset"` // Misal: Gerobak, Dandang, Mangkok
	HargaHari float64   `gorm:"not null" json:"harga_hari"` // Harga sewa per hari
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (KatalogSewa) TableName() string { return "katalog_sewa" }

// 2. Pencatatan Barang yang Disewa Mitra
type SewaMitra struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	PenjualID     uint        `json:"penjual_id"`
	// 👇 UBAH json:"-" MENJADI json:"penjual" 👇
	Penjual       Penjual     `gorm:"foreignKey:PenjualID" json:"penjual"`
	KatalogSewaID uint        `json:"katalog_sewa_id"`
	KatalogSewa   KatalogSewa `gorm:"foreignKey:KatalogSewaID" json:"katalog_sewa"`
	TanggalMulai  time.Time   `gorm:"type:date;not null" json:"tanggal_mulai"`
	TanggalKembali *time.Time  `gorm:"type:date" json:"tanggal_kembali"` 
	IsActive      bool        `gorm:"default:true" json:"is_active"` 
}

func (SewaMitra) TableName() string { return "sewa_mitra" }

// 3. Pencatatan Izin (Libur Jualan = Bebas Sewa)
type IzinMitra struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	PenjualID    uint      `json:"penjual_id"`
	// 👇 UBAH json:"-" MENJADI json:"penjual" 👇
	Penjual      Penjual   `gorm:"foreignKey:PenjualID" json:"penjual"`
	TanggalMulai time.Time `gorm:"type:date;not null" json:"tanggal_mulai"`
	TanggalAkhir time.Time `gorm:"type:date;not null" json:"tanggal_akhir"`
	Keterangan   string    `json:"keterangan"`
	CreatedAt    time.Time `json:"created_at"`
}

func (IzinMitra) TableName() string { return "izin_mitra" }

// 4. Tagihan Invoice Bulanan/Mingguan
type InvoiceSewa struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PenjualID      uint      `json:"penjual_id"`
	// 👇 UBAH json:"-" MENJADI json:"penjual" 👇
	Penjual        Penjual   `gorm:"foreignKey:PenjualID" json:"penjual"`
	NoInvoice      string    `gorm:"unique;not null" json:"no_invoice"` 
	Bulan          int       `json:"bulan"`
	Tahun          int       `json:"tahun"`
	TotalSewa      float64   `json:"total_sewa"` 
	TotalDiskonIzin float64  `json:"total_diskon_izin"` 
	GrandTotal     float64   `json:"grand_total"` 
	StatusLunas    bool      `gorm:"default:false" json:"status_lunas"`
	CreatedAt      time.Time `json:"created_at"`
}

func (InvoiceSewa) TableName() string { return "invoice_sewa" }
