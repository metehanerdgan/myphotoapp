package photo

import (
	"context"
	"log"

	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"google.golang.org/api/option"
)

// VisionAPI, görüntü analizi işlemlerini yöneten bir yapıdır.
type VisionAPI struct {
	client *vision.ImageAnnotatorClient
}

// NewVisionAPI, yeni bir VisionAPI örneği oluşturur.
func NewVisionAPI(ctx context.Context, credentialsFile string) (*VisionAPI, error) {
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Fatalf("Vision API istemcisi oluşturulamadı: %v", err)
		return nil, err
	}

	return &VisionAPI{client: client}, nil
}

// AnalyzeFaces, Vision API kullanarak bir görüntüdeki yüzleri analiz eder.
func (v *VisionAPI) AnalyzeFaces(ctx context.Context, imageURI string) ([]*FaceAnalysisResult, error) {
	// Vision API kullanarak yüz analizi işlemini burada gerçekleştirinr.
	annotations, err := v.client.AnnotateImage(ctx, &visionpb.AnnotateImageRequest{
		Image: &visionpb.Image{
			Source: &visionpb.ImageSource{
				ImageUri: imageURI,
			},
		},
		Features: []*visionpb.Feature{
			{
				Type: visionpb.Feature_FACE_DETECTION,
			},
		},
	})
	if err != nil {
		log.Printf("Yüz analizi başarısız: %v", err)
		return nil, err
	}

	// Yüz analizi sonuçlarını tutmak için bir dilim oluştur.
	var results []*FaceAnalysisResult

	// Tüm yüzleri döngü ile işler.
	for _, faceAnnotation := range annotations.FaceAnnotations {
		// Yüz analizi sonuçlarına göre hissiyatı ve güvenilirlik puanını belirler.
		emotion := determineEmotion(faceAnnotation)
		confidence := calculateConfidence(faceAnnotation)

		// Her bir yüz için bir FaceAnalysisResult oluşturur.
		result := &FaceAnalysisResult{
			Emotion:    emotion,
			Confidence: confidence,
		}

		// Oluşturulan FaceAnalysisResult'ı results dilimine ekler.
		results = append(results, result)
	}

	return results, nil
}

func determineEmotion(faceAnnotation *visionpb.FaceAnnotation) string {
	// Yüz analizi sonuçlarına göre duyguyu belirler.
	// Örnek olarak, Joy, Sorrow, Anger, Surprise, vb. gibi duyguları belirlemek için bir algoritma kullanır.
	// Kullanılacak algoritma, API'nin dönüş değerlerine göre günceller.

	// Örnek bir algoritma: Yüzün gülümsüyüp gülümsmediğini kontrol eder.
	if faceAnnotation.JoyLikelihood == visionpb.Likelihood_VERY_LIKELY || faceAnnotation.JoyLikelihood == visionpb.Likelihood_LIKELY {
		// Yüz gülümsüyorsa, "Joy" yani sevinç döndürür.
		return "Joy"
	}

	// Örnek bir algoritma: Gözlerin kapanıp kapanmadığını kontrol eder.
	if faceAnnotation.SorrowLikelihood == visionpb.Likelihood_VERY_LIKELY || faceAnnotation.SorrowLikelihood == visionpb.Likelihood_LIKELY {
		// Gözler kapatılmışsa, "Sorrow" yani üzüntü döndürür.
		return "Sorrow"
	}

	// Örnek bir algoritma: Kaşların çatılıp çatılmadığını kontrol eder.
	if faceAnnotation.AngerLikelihood == visionpb.Likelihood_VERY_LIKELY || faceAnnotation.AngerLikelihood == visionpb.Likelihood_LIKELY {
		// Kaşlar çatılmışsa, "Anger" yani öfke döndürür.
		return "Anger"
	}

	// Örnek bir algoritma: Şaşkın ifadeyi kontrol eder.
	if faceAnnotation.SurpriseLikelihood == visionpb.Likelihood_VERY_LIKELY || faceAnnotation.SurpriseLikelihood == visionpb.Likelihood_LIKELY {
		// Şaşkın ifade varsa, "Surprise" yani şaşkınlık döndürür.
		return "Surprise"
	}
	return "Unknown" // Eğer belirlenebilmiş bir duygu yoksa "Unknown" yani bilinmeyen olarak döndür.
}

// calculateConfidence, yüz analizi sonuçlarından güvenilirlik puanını hesaplar.
func calculateConfidence(faceAnnotation *visionpb.FaceAnnotation) float64 {
	// Burada yüz analizi sonuçlarından güvenilirlik puanını hesaplar.

	// Eğer yüz analizi sonuçları yoksa veya güvenilirlik puanı belirlenmemişse, varsayılan bir değer döndürür.
	if faceAnnotation.DetectionConfidence == 0 {
		return 0.0
	}

	// Vision API tarafından sağlanan güvenilirlik puanını döndürür.
	return float64(faceAnnotation.DetectionConfidence)
}

// FaceAnalysisResult, yüz analizi sonuçlarını temsil eder.
type FaceAnalysisResult struct {
	Emotion    string  // Algılanan duygu (örneğin, Joy, Sorrow, Anger, Surprise vb.)
	Confidence float64 // Duygu algısının güvenilirlik puanı
}

// Close, Vision API istemcisini kapatır.
func (v *VisionAPI) Close() {
	if v.client != nil {
		if err := v.client.Close(); err != nil {
			log.Printf("Vision API istemcisi kapatılamadı: %v", err)
		}
	}
}
