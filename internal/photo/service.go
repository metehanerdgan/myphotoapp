package photo

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"
)

// Image, fotoğraf bilgilerini temsil eder.
type Image struct {
	Id           string
	Url          string
	FaceAnalysis []*FaceAnalysis
	UploadTime   time.Time
}

// PhotoService, fotoğraf işlemleriyle ilgili istekleri yöneten bir yapıdır.
type PhotoService struct {
	kafkaProducer  *KafkaProducer
	visionAPI      *VisionAPI
	uploadedImages []*UploadedImage // Yüklenen fotoğrafları saklamak için bir dilim
}

// NewPhotoService, yeni bir PhotoService örneği oluşturur.
func NewPhotoService(kp *KafkaProducer, va *VisionAPI) *PhotoService {
	return &PhotoService{
		kafkaProducer:  kp,
		visionAPI:      va,
		uploadedImages: make([]*UploadedImage, 0), // Boş bir dilim oluşturur
	}
}

// UploadImage, yeni bir fotoğrafı sisteme yükleyen işlemi gerçekleştirir.
func (s *PhotoService) UploadImage(ctx context.Context, image *UploadedImage) (*UploadedImage, error) {

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
		Id:  "1",
		Url: image.Url,
		FaceAnalysis: []*FaceAnalysis{
			{
				Emotion:    faceAnalysisResult[0].Emotion,
				Confidence: float32(faceAnalysisResult[0].Confidence),
			},
		},
		UploadTime: time.Now().Unix(),
	}

	s.uploadedImages = append(s.uploadedImages, uploadedImage)

	// Kafka'ya asenkron bir şekilde Vision API için mesaj gönderir.
	err = s.kafkaProducer.ProduceMessage("image-upload-topic", "Image Uploaded: "+uploadedImage.Id)
	if err != nil {
		log.Printf("Kafka'ya mesaj gönderirken hata oluştu: %v", err)
	}

	return uploadedImage, nil
}

// GetImageDetail, belli bir fotoğrafın yüz analizi bilgilerini verir.
func (s *PhotoService) GetImageDetail(ctx context.Context, req *UploadedImage) (*UploadedImage, error) {
	// Fotoğrafı oluşturur.
	imageDetail := &UploadedImage{
		Id:         req.Id,
		Url:        req.Url,
		UploadTime: time.Now().Unix(),
	}
	// Yüz analizi sonuçlarını alır.
	faceAnalysisResult, err := s.visionAPI.AnalyzeFaces(ctx, req.Url)
	if err != nil {
		return nil, fmt.Errorf("Yüz analizi yapılırken hata oluştu: %v", err)
	}

	// Yüz analizi sonuçları diliminin boş olup olmadığını kontrol eder.
	if len(faceAnalysisResult) == 0 {
		return nil, fmt.Errorf("Yüz analizi sonuçları bulunamadı")
	}

	// Yüz analizi sonuçlarını fotoğraf detayına ekler.
	imageDetail.FaceAnalysis = []*FaceAnalysis{
		{
			Emotion:    faceAnalysisResult[0].Emotion,
			Confidence: float32(faceAnalysisResult[0].Confidence),
		},
	}
	return imageDetail, nil
}

// GetImageFeed, fotoğrafları yüklenme tarihine ve analiz değerlerine göre sıralayarak sayfalandıran işlemi gerçekleştirir.
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

	// Tüm fotoğrafları almak (örnek olarak iki fotoğraf ekledim.)
	images := []*Image{
		{
			Id:           "1",
			Url:          "https://example.com/image1.jpg",
			FaceAnalysis: []*FaceAnalysis{{Emotion: "Joy", Confidence: 0.95}},
			UploadTime:   time.Now().Add(-time.Hour * 24 * 3), // Örnek bir yükleme tarihi
		},
		{
			Id:           "2",
			Url:          "https://example.com/image2.jpg",
			FaceAnalysis: []*FaceAnalysis{{Emotion: "Sorrow", Confidence: 0.75}},
			UploadTime:   time.Now().Add(-time.Hour * 24 * 2),
		},
	}

	// Fotoğrafları yüklenme tarihine ve analiz değerlerine göre sıralar.
	sort.Slice(images, func(i, j int) bool {
		// İlk olarak yüklenme tarihine göre sıralar.
		if images[i].UploadTime.Equal(images[j].UploadTime) {
			// Eğer yüklenme tarihleri aynıysa, analiz ortalamalarına göre sıralar.
			avgEmotion1 := calculateAverageEmotion(images[i].FaceAnalysis)
			avgEmotion2 := calculateAverageEmotion(images[j].FaceAnalysis)

			return avgEmotion1 > avgEmotion2 // Büyükten küçüğe sıralar. (yüksek ortalama önce gelsin)
		}

		// Eğer yüklenme tarihleri farklıysa, tarih sıralamasını kullanır.
		return images[i].UploadTime.After(images[j].UploadTime)
	})

	// Sayfalama hesaplamalarını yapar.
	startIndex := int((pageNumber - 1) * pageSize)
	endIndex := int(pageNumber * pageSize)

	// Başlangıç ve bitiş indekslerini kontrol eder.
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > len(images) {
		endIndex = len(images)
	}

	// Sayfa boyunca olan fotoğrafları alır.
	var pageImages []*Image
	for _, img := range images[startIndex:endIndex] {
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
	// Yüklenen fotoğrafları kullanın
	var uploadPageImages []*UploadedImage
	for _, img := range s.uploadedImages {
		uploadPageImages = append(uploadPageImages, &UploadedImage{
			Id:  img.Id,
			Url: img.Url,
			FaceAnalysis: []*FaceAnalysis{
				{Emotion: img.FaceAnalysis[0].Emotion, Confidence: img.FaceAnalysis[0].Confidence},
			},
			UploadTime: img.UploadTime,
		})
	}

	// Sayfalama sonuçlarını oluşturur.
	response := &GetImageFeedResponse{
		Images: uploadPageImages,
	}

	return response, nil
}

// UpdateImageDetail, fotoğraf detaylarını günceller.
func (s *PhotoService) UpdateImageDetail(ctx context.Context, req *UploadedImage) (*UploadedImage, error) {
	// Sadece eklerken değil güncellerken de yüz analizi sonuçlarını alır ve fotoğraf yolunu değiştirir sadece.
	faceAnalysisResult, err := s.visionAPI.AnalyzeFaces(ctx, req.Url)
	if err != nil {
		return nil, fmt.Errorf("Yüz analizi yapılırken hata oluştu: %v", err)
	}

	// Yüz analizi sonuçları diliminin boş olup olmadığını kontrol eder.
	if len(faceAnalysisResult) == 0 {
		return nil, fmt.Errorf("Yüz analizi sonuçları bulunamadı")
	}
	// Yüklenen fotoğrafı oluşturur.
	updatedImage := &UploadedImage{
		Id:  "1",
		Url: req.Url,
		FaceAnalysis: []*FaceAnalysis{
			{
				Emotion:    faceAnalysisResult[0].Emotion,
				Confidence: float32(faceAnalysisResult[0].Confidence),
			},
		},
		UploadTime: time.Now().Unix(),
	}
	return updatedImage, nil
}

// calculateAverageEmotion, yüz analizi sonuçlarının ortalamasını hesaplar.
func calculateAverageEmotion(faceAnalysis []*FaceAnalysis) float32 {
	if len(faceAnalysis) == 0 {
		return 0.0
	}

	var totalConfidence float32
	for _, analysis := range faceAnalysis {
		totalConfidence += analysis.Confidence
	}

	return totalConfidence / float32(len(faceAnalysis))
}
