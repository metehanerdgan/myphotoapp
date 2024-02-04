package photo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"
	stdtime "time"
)

var now = stdtime.Now

// Image, fotoğraf bilgilerini temsil eder.
type Image struct {
	Id           string
	Url          string
	FaceAnalysis []*FaceAnalysis
	UploadTime   int64
}

// PhotoService, fotoğraf işlemleriyle ilgili istekleri yöneten bir yapıdır.
type PhotoService struct {
	kafkaProducer  *KafkaProducer
	visionAPI      *VisionAPI
	uploadedImages []*UploadedImage // Yüklenen fotoğrafları saklamak için bir dilim
	dbImages       []*UploadedImage // Veritabanından çekilen fotoğrafları saklamak için bir dilim
	lastImageID    int              // Son atanan ID'yi takip etmek için
}

// GetHighestPhotoIDFromDB, veritabanındaki en yüksek ID değerini alır.
func GetHighestPhotoIDFromDB() (int, error) {
	var highestID int
	err := db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM photos").Scan(&highestID)
	if err != nil {
		return 0, err
	}
	return highestID, nil
}

// NewPhotoService, yeni bir PhotoService örneği oluşturur.
func NewPhotoService(kp *KafkaProducer, va *VisionAPI) *PhotoService {
	// Veritabanından en yüksek ID değerini al
	highestID, err := GetHighestPhotoIDFromDB()
	if err != nil {
		log.Printf("En yüksek ID alınamadı: %v", err)
		highestID = 0
	}

	return &PhotoService{
		kafkaProducer:  kp,
		visionAPI:      va,
		uploadedImages: make([]*UploadedImage, 0),
		dbImages:       make([]*UploadedImage, 0),
		lastImageID:    highestID,
	}
}

// GetPhotosFromDB, veritabanından tüm fotoğrafları çeker.
func GetPhotosFromDB() ([]*UploadedImage, error) {
	rows, err := db.Query(`SELECT id, url, emotion, confidence, upload_time FROM photos`)
	if err != nil {
		log.Printf("Fotoğraflar alınamadı: %v", err)
		return nil, err
	}
	defer rows.Close()

	var dbImages []*UploadedImage
	for rows.Next() {
		var uploadedImage UploadedImage
		var uploadTime time.Time

		// FaceAnalysis dilimini oluştur
		uploadedImage.FaceAnalysis = make([]*FaceAnalysis, 1)
		uploadedImage.FaceAnalysis[0] = &FaceAnalysis{}
		err := rows.Scan(&uploadedImage.Id, &uploadedImage.Url, &uploadedImage.FaceAnalysis[0].Emotion, &uploadedImage.FaceAnalysis[0].Confidence, &uploadTime)
		if err != nil {
			log.Printf("Fotoğraf alınamadı: %v", err)
			return nil, err
		}

		// uploadTime'ı int64'e dönüştürür.
		uploadedImage.UploadTime = uploadTime.Unix()

		// Boş bir dilim oluştur
		if uploadedImage.FaceAnalysis[0] == nil {
			uploadedImage.FaceAnalysis[0] = &FaceAnalysis{}
		}

		dbImages = append(dbImages, &uploadedImage)
	}

	return dbImages, nil
}

// GetPhotoByID, belirli bir ID'ye sahip fotoğrafı veritabanından çeker.
func GetPhotoByID(id string) (*UploadedImage, error) {
	row := db.QueryRow("SELECT id, url, emotion, confidence, upload_time FROM photos WHERE id = $1", id)
	return scanPhoto(row)
}

// scanPhoto, veritabanı satırındaki verileri *UploadedImage türündeki bir nesneye tarar.
func scanPhoto(row *sql.Row) (*UploadedImage, error) {
	var img UploadedImage
	var emotion sql.NullString
	var confidence sql.NullFloat64
	var uploadTime time.Time

	err := row.Scan(&img.Id, &img.Url, &emotion, &confidence, &uploadTime)
	if err != nil {
		return nil, err
	}

	// FaceAnalysis dilimini oluşturur.
	img.FaceAnalysis = make([]*FaceAnalysis, 1)

	// Eğer emotion ve confidence veritabanında NULL ise, FaceAnalysis dilimini boş bırakır.
	if emotion.Valid && confidence.Valid {
		img.FaceAnalysis[0] = &FaceAnalysis{
			Emotion:    emotion.String,
			Confidence: float32(confidence.Float64),
		}
	} else {
		img.FaceAnalysis = nil
	}

	// uploadTime'ı int64'e dönüştürür.
	img.UploadTime = uploadTime.Unix()

	return &img, nil
}

// UploadImage, yeni bir fotoğrafı sisteme yükleyen işlemi gerçekleştirir.
func (s *PhotoService) UploadImage(ctx context.Context, image *UploadedImage) (*UploadedImage, error) {

	// ID'yi bir artırarak yeni bir ID oluşturur.
	s.lastImageID++
	newImageID := s.lastImageID
	// Yüz analizi sonuçlarını alır.
	faceAnalysisResult, err := s.visionAPI.AnalyzeFaces(ctx, image.Url)
	if err != nil {
		return nil, fmt.Errorf("Yüz analizi yapılırken hata oluştu: %v", err)
	}

	// Yüz analizi sonuçları diliminin boş olup olmadığını kontrol eder.
	if len(faceAnalysisResult) == 0 {
		return nil, fmt.Errorf("Yüz analizi sonuçları bulunamadı")
	}
	// Yüklenen fotoğrafı oluşturur.
	uploadedImage := &UploadedImage{
		Id:  strconv.Itoa(newImageID),
		Url: image.Url,
		FaceAnalysis: []*FaceAnalysis{
			{
				Emotion:    faceAnalysisResult[0].Emotion,
				Confidence: float32(faceAnalysisResult[0].Confidence),
			},
		},
		UploadTime: now().Unix(),
	}
	s.uploadedImages = append(s.uploadedImages, uploadedImage)

	// Veritabanına fotoğrafı ekler.
	if err := InsertPhoto(uploadedImage); err != nil {
		// Veritabanına ekleme hatası
		log.Printf("Veritabanına fotoğraf eklenirken hata oluştu: %v", err)
		return nil, err
	}

	return uploadedImage, nil

	// Kafka'ya asenkron bir şekilde Vision API için mesaj gönderir.
	err = s.kafkaProducer.ProduceMessage("image-upload-topic", "Fotoğraf Yüklendi: "+uploadedImage.Id)
	if err != nil {
		log.Printf("Kafka'ya mesaj gönderirken hata oluştu: %v", err)
	}

	return uploadedImage, nil
}

// GetImageDetail, belli bir fotoğrafın yüz analizi bilgilerini verir.
func (s *PhotoService) GetImageDetail(ctx context.Context, req *UploadedImage) (*UploadedImage, error) {
	// Fotoğrafı veritabanından çeker.
	dbImage, err := GetPhotoByID(req.Id)
	if err != nil {
		return nil, fmt.Errorf("Fotoğraf bulunamadı: %v", err)
	}

	// Yüz analizi sonuçlarını alır.
	faceAnalysisResult, err := s.visionAPI.AnalyzeFaces(ctx, dbImage.Url)
	if err != nil {
		return nil, fmt.Errorf("Yüz analizi yapılırken hata oluştu: %v", err)
	}

	// Yüz analizi sonuçları diliminin boş olup olmadığını kontrol eder.
	if len(faceAnalysisResult) == 0 {
		return nil, fmt.Errorf("Yüz analizi sonuçları bulunamadı")
	}

	// Yüz analizi sonuçlarını fotoğraf detayına ekler.
	dbImage.FaceAnalysis = []*FaceAnalysis{
		{
			Emotion:    faceAnalysisResult[0].Emotion,
			Confidence: float32(faceAnalysisResult[0].Confidence),
		},
	}
	return dbImage, nil
}

// GetImageFeed, yüklenen fotoğrafları yüklenme tarihine ve analiz değerlerine göre sıralayarak sayfalandıran işlemi gerçekleştirir.
func (s *PhotoService) GetImageFeed(ctx context.Context, req *GetImageFeedRequest) (*GetImageFeedResponse, error) {

	// Örnek bir sayfa büyüklüğü...

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	// Örnek bir sayfa numarası belirler.
	pageNumber := req.PageNumber
	if pageNumber <= 0 {
		pageNumber = 1
	}
	// Yüklenen fotoğrafları veritabanından çeker.
	var err error
	s.dbImages, err = GetPhotosFromDB()
	if err != nil {
		return nil, fmt.Errorf("Veritabanından fotoğraflar alınamadı: %v", err)
	}

	// Yüklenen fotoğrafları yüklenme tarihine ve analiz değerlerine göre sıralar.
	sort.Slice(s.dbImages, func(i, j int) bool {
		// İlk olarak yüklenme tarihine göre sıralar.
		timeI := time.Unix(s.dbImages[i].UploadTime, 0)
		timeJ := time.Unix(s.dbImages[j].UploadTime, 0)

		if timeI.Equal(timeJ) {
			// Eğer yüklenme tarihleri aynıysa, analiz ortalamalarına göre sıralar.
			avgEmotion1 := calculateAverageEmotion(s.dbImages[i].FaceAnalysis)
			avgEmotion2 := calculateAverageEmotion(s.dbImages[j].FaceAnalysis)

			return avgEmotion1 > avgEmotion2 // Büyükten küçüğe sıralar. (yüksek ortalama önce gelsin)
		}

		// Eğer yüklenme tarihleri farklıysa, tarih sıralamasını kullanır.
		return timeI.After(timeJ)
	})
	// Sayfalama hesaplamalarını yapar.
	startIndex := int((pageNumber - 1) * pageSize)
	endIndex := int(pageNumber * pageSize)

	// Başlangıç ve bitiş indekslerini kontrol eder.
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > len(s.dbImages) {
		endIndex = len(s.dbImages)
	}
	// Sayfa boyunca olan fotoğrafları alır.
	var pageImages []*Image
	for _, img := range s.dbImages[startIndex:endIndex] {
		if len(img.FaceAnalysis) > 0 {
			pageImages = append(pageImages, &Image{
				Id:  img.Id,
				Url: img.Url,
				FaceAnalysis: []*FaceAnalysis{
					{Emotion: img.FaceAnalysis[0].Emotion, Confidence: img.FaceAnalysis[0].Confidence},
				},
				UploadTime: img.UploadTime,
			})
		} else {
			pageImages = append(pageImages, &Image{
				Id:           img.Id,
				Url:          img.Url,
				FaceAnalysis: nil,
				UploadTime:   img.UploadTime,
			})
		}
	}

	// Sayfalama sonuçlarını oluşturur.
	response := &GetImageFeedResponse{
		Images: s.dbImages,
	}

	return response, nil
}

// UpdateImageDetail, fotoğraf detaylarını günceller.
func (s *PhotoService) UpdateImageDetail(ctx context.Context, req *UploadedImage) (*UploadedImage, error) {
	// Fotoğrafı veritabanından çeker.
	dbImage, err := GetPhotoByID(req.Id)
	if err != nil {
		return nil, fmt.Errorf("Fotoğraf bulunamadı: %v", err)
	}

	// Yeni URL için Vision API'yi kullanarak yüz analizi yapar.
	faceAnalysisResult, err := s.visionAPI.AnalyzeFaces(ctx, req.Url)
	if err != nil {
		return nil, fmt.Errorf("Yüz analizi yapılırken hata oluştu: %v", err)
	}

	// Yüz analizi sonuçları diliminin boş olup olmadığını kontrol eder.
	if len(faceAnalysisResult) == 0 {
		return nil, fmt.Errorf("Yüz analizi sonuçları bulunamadı")
	}

	// Güncelleme işlemi
	dbImage.Url = req.Url
	dbImage.FaceAnalysis = []*FaceAnalysis{
		{
			Emotion:    faceAnalysisResult[0].Emotion,
			Confidence: float32(faceAnalysisResult[0].Confidence),
		},
	}
	dbImage.UploadTime = time.Now().Unix()

	// UpdatePhoto fonksiyonunu kullanarak veritabanında güncelleme yapar.
	err = UpdatePhoto(dbImage)
	if err != nil {
		return nil, fmt.Errorf("Fotoğraf veritabanında güncellenemedi: %v", err)
	}

	return dbImage, nil
}

// calculateAverageEmotion, yüz analizi sonuçlarının ortalamasını hesaplar.
func calculateAverageEmotion(faceAnalysis []*FaceAnalysis) float32 {

	var totalConfidence float32
	for _, analysis := range faceAnalysis {
		totalConfidence += analysis.Confidence
	}

	return totalConfidence / float32(len(faceAnalysis))
}
