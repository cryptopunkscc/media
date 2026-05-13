package db

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/media/obj"
	"gorm.io/gorm/clause"
)

type AudioRow struct {
	ObjectID  *astral.ObjectID `gorm:"primaryKey"`
	Format    string           `gorm:"index"`
	Duration  time.Duration
	Title     string           `gorm:"index"`
	Artist    string           `gorm:"index"`
	Album     string           `gorm:"index"`
	Genre     string           `gorm:"index"`
	Year      int              `gorm:"index"`
	PictureID *astral.ObjectID `gorm:"index"`
}

func (AudioRow) TableName() string { return obj.DBPrefix + "audio" }

func (row *AudioRow) ToAudioFile() *obj.AudioFile {
	return &obj.AudioFile{
		ObjectID:  row.ObjectID,
		Format:    astral.String8(row.Format),
		Title:     astral.String8(row.Title),
		Artist:    astral.String8(row.Artist),
		Album:     astral.String8(row.Album),
		Genre:     astral.String8(row.Genre),
		Year:      astral.Uint16(row.Year),
		PictureID: row.PictureID,
	}
}

type ObjectRow struct {
	ObjectID *astral.ObjectID `gorm:"primaryKey"`
	Audio    *AudioRow        `gorm:"foreignKey:ObjectID"`
}

func (ObjectRow) TableName() string { return obj.DBPrefix + "objects" }

func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(&AudioRow{}, &ObjectRow{})
}

func (db *DB) FindObject(objectID *astral.ObjectID) (row *ObjectRow, err error) {
	err = db.Where("object_id = ?", objectID).First(&row).Error
	return
}

func (db *DB) SaveObject(objectID *astral.ObjectID) error {
	return db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&ObjectRow{ObjectID: objectID}).
		Error
}

func (db *DB) DeleteObject(objectID *astral.ObjectID) error {
	return db.Delete(&ObjectRow{ObjectID: objectID}).Error
}

func (db *DB) FindAudio(objectID *astral.ObjectID) (row *AudioRow, err error) {
	err = db.Where("object_id = ?", objectID).First(&row).Error
	return
}

func (db *DB) SaveAudio(audio *obj.AudioFile) error {
	return db.
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&AudioRow{
			ObjectID:  audio.ObjectID,
			Format:    string(audio.Format),
			Title:     string(audio.Title),
			Artist:    string(audio.Artist),
			Album:     string(audio.Album),
			Genre:     string(audio.Genre),
			Year:      int(audio.Year),
			PictureID: audio.PictureID,
		}).Error
}

func (db *DB) DeleteAudio(objectID *astral.ObjectID) error {
	return db.Delete(&AudioRow{ObjectID: objectID}).Error
}
