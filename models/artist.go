package models

import (
	"gorm.io/gorm"
)

type Artist struct {
	ID			uint	`gorm:"primaryKey"`
    Name    	string  `gorm:"size:255"`
    Country 	string  `gorm:"size:255"`

	MusicFiles 	[]MusicFile `gorm:"foreignKey:ArtistID"`
}

func MigrateArtist(db *gorm.DB) error {
	return db.AutoMigrate(&Artist{})
}