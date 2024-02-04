package photo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	dbHost     = "localhost"
	dbPort     = 5432
	dbName     = "postgres"
	dbUser     = "postgres"
	dbPassword = "12345"
)

var db *sql.DB

// InitDB, veritabanı bağlantısını başlatır.
func InitDB() error {
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		dbHost, dbPort, dbName, dbUser, dbPassword)

	var err error
	db, err = sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
		return err
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Veritabanına bağlanılamadı: %v", err)
		return err
	}

	log.Printf("Veritabanına bağlantı başarılı")
	return nil
}

// CloseDB, veritabanı bağlantısını kapatır.
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// CreatePhotoTable, "photos" adında bir tablo oluşturur.
func CreatePhotoTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS photos (
        id SERIAL PRIMARY KEY,
        url TEXT,
        emotion TEXT,
        confidence FLOAT,
        upload_time TIMESTAMP
    )`)

	if err != nil {
		log.Printf("Tablo oluşturulamadı: %v", err)
		return err
	}

	log.Printf("Tablo oluşturuldu")
	return nil
}

// InsertPhoto, fotoğraf bilgilerini veritabanına ekler.
func InsertPhoto(photo *UploadedImage) error {
	if len(photo.FaceAnalysis) > 0 {
		_, err := db.Exec(`INSERT INTO photos (url, emotion, confidence, upload_time)
                          VALUES ($1, $2, $3, $4)`,
			photo.Url, photo.FaceAnalysis[0].Emotion, photo.FaceAnalysis[0].Confidence, now().UTC())

		if err != nil {
			log.Printf("Fotoğraf eklenemedi: %v", err)
			return err
		}
		log.Printf("Fotoğraf başarıyla eklendi")
		return nil
	} else {
		err := errors.New("analiz bilgileri bulunamadığı için fotoğraf eklenemedi")
		log.Printf("%v", err)
		return err
	}
}

// UpdatePhoto, veritabanındaki fotoğraf bilgilerini günceller.
func UpdatePhoto(img *UploadedImage) error {
	_, err := db.Exec(`
		UPDATE photos
		SET url = $2, emotion = $3, confidence = $4, upload_time = $5
		WHERE id = $1`,
		img.Id, img.Url, img.FaceAnalysis[0].Emotion, img.FaceAnalysis[0].Confidence, time.Unix(img.UploadTime, 0))

	if err != nil {
		log.Printf("Fotoğraf güncellenirken hata oluştu: %v", err)
		return err
	}

	return nil
}
