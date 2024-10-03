package models

import (
	"time"

	"gorm.io/gorm"
)

type MusicFile  struct {
	ID			uint		`gorm:"primaryKey"`
	Title		string		`gorm:"size:255"`
	ArtistID	uint	
	Album		string		`gorm:"size:255"`
	AlbumCover	string		`gorm:"size:255"`
	AlbumColors string		`gorm:"type:text"`
	FilePath	string		`gorm:"size:255"`
	UploadedAt	time.Time	`gorm:"autoCreateTime"`
}

func MigrateMusicFile(db *gorm.DB) error {
	return db.AutoMigrate(&MusicFile{})
}