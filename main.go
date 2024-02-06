package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"myphotoapp/internal/photo"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// PhotoServiceServerImpl struct'ını tanımlar.
type PhotoServiceServerImpl struct {
	photo.UnimplementedPhotoServiceServer
	photoService *photo.PhotoService
}

func main() {
	// gRPC sunucu dinleyiciyi oluşturur.
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Dinleme başarısız: %v", err)
	}

	// Veritabanı bağlantısını başlatır.
	err = photo.InitDB()
	if err != nil {
		log.Fatalf("Veritabanı bağlantısı başlatılamadı: %v", err)
	}
	defer func() {
		if cerr := photo.CloseDB(); cerr != nil {
			log.Fatalf("Veritabanı bağlantısı kapatılamadı: %v", cerr)
		}
	}()

	// "photos" tablosunu oluştur.
	err = photo.CreatePhotoTable()
	if err != nil {
		log.Fatalf("Tablo oluşturulamadı: %v", err)
	}

	// Kafka üretici ve Vision API istemcisini oluşturur.
	kafkaProducer, err := photo.NewKafkaProducer()
	if err != nil {
		log.Fatalf("Kafka üretici başlatılamadı: %v", err)
	}
	defer kafkaProducer.Close()

	visionAPI, err := photo.NewVisionAPI(context.Background(), "config/myphotoapp-412717-7662d103da43.json")
	if err != nil {
		log.Fatalf("Vision API istemcisi oluşturulamadı: %v", err)
	}
	defer visionAPI.Close()

	// PhotoService oluşturur.
	photoService := photo.NewPhotoService(kafkaProducer, visionAPI)

	// UploadImage örneği 1
	image1 := &photo.UploadedImage{
		Url: "https://png.pngtree.com/thumb_back/fw800/background/20230425/pngtree-woman-making-an-angry-face-with-her-eyebrows-crossed-image_2554181.jpg",
	}

	// UploadImage metodunu kullanarak fotoğrafı yükler.
	_, err = photoService.UploadImage(context.Background(), image1)
	if err != nil {
		log.Fatal(err)
	}

	// UploadImage örneği 2
	image2 := &photo.UploadedImage{
		Url: "https://www.aljazeera.com.tr/sites/default/files/styles/aljazeera_article_main_image/public/2014/04/16/face_shutter_main.jpg?itok=ED671aXO",
	}

	// UploadImage metodunu kullanarak fotoğrafı yükler.
	_, err = photoService.UploadImage(context.Background(), image2)
	if err != nil {
		log.Fatal(err)
	}

	// UploadImage örneği 3
	image3 := &photo.UploadedImage{
		Url: "https://img3.stockfresh.com/files/k/kurhan/m/59/1185098_stock-photo-man.jpg",
	}

	// UploadImage metodunu kullanarak fotoğrafı yükler.
	_, err = photoService.UploadImage(context.Background(), image3)
	if err != nil {
		log.Fatal(err)
	}

	// GetImageFeed metodunu kullanarak yüklenen fotoğrafları listeler.
	imageFeedResponse, err := photoService.GetImageFeed(context.Background(), &photo.GetImageFeedRequest{
		PageSize:   10,
		PageNumber: 1,
	})
	if err != nil {
		log.Fatal(err)
	}

	// // Liste sonuçlarını yazdırır.
	fmt.Println("Fotoğraf Listesi:")
	for _, img := range imageFeedResponse.Images {
		fmt.Printf("ID: %s, URL: %s\n", img.Id, img.Url)

		// 	// Fotoğrafın analiz bilgilerini de yazdır.
		for _, faceAnalysis := range img.FaceAnalysis {
			fmt.Printf("  Analiz: %s, Güvenirlik Oranı: %f\n", faceAnalysis.Emotion, faceAnalysis.Confidence)
		}
	}

	// Kullanıcıdan yeni bilgileri alır.
	newImageDetails := &photo.UploadedImage{
		Id:  "2",                                                                                                                        // Güncellenecek fotoğrafın ID'si
		Url: "https://st3.depositphotos.com/1258191/17024/i/950/depositphotos_170241044-stock-photo-aggressive-angry-woman-yelling.jpg", // Yeni URL
	}

	// Güncellenmiş fotoğraf detayını alır.
	updatedDetail, err := photoService.UpdateImageDetail(context.Background(), newImageDetails)
	if err != nil {
		log.Fatalf("Fotoğraf detayı güncellenirken hata oluştu: %v", err)
	}

	log.Printf("Güncellenmiş fotoğraf detayı: %v", updatedDetail)

	// gRPC sunucu oluşturur.
	grpcServer := grpc.NewServer()

	// PhotoService sunucuya ekler.
	photoServiceServerImpl := &PhotoServiceServerImpl{photoService: photoService}
	photo.RegisterPhotoServiceServer(grpcServer, photoServiceServerImpl)

	log.Printf("gRPC sunucusu %s üzerinde dinleniyor", port)

	// Sunucuyu başlatır.
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
