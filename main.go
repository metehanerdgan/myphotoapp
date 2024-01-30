package main

import (
	"context"
	"log"
	"net"
	"time"

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

	// Kafka üretici ve Vision API istemcisini oluşturur.
	kafkaProducer, err := photo.NewKafkaProducer()
	if err != nil {
		log.Fatalf("Kafka üretici başlatılamadı: %v", err)

		log.Printf("\n")
		log.Printf("\n")
	}
	defer kafkaProducer.Close()

	visionAPI, err := photo.NewVisionAPI(context.Background(), "config/myphotoapp-412717-7662d103da43.json")
	if err != nil {
		log.Fatalf("Vision API istemcisi oluşturulamadı: %v", err)
	}
	defer visionAPI.Close()

	// PhotoService oluşturur.
	photoService := photo.NewPhotoService(kafkaProducer, visionAPI)

	// Örnek bir fotoğraf bilgisi oluşturur.
	uploadedImage := &photo.UploadedImage{
		Id:  "1",
		Url: "https://png.pngtree.com/thumb_back/fw800/background/20230425/pngtree-woman-making-an-angry-face-with-her-eyebrows-crossed-image_2554181.jpg",
	}

	// UploadImage metodunu çağırarak fotoğrafı yükler.
	response, err := photoService.UploadImage(context.Background(), uploadedImage)
	if err != nil {
		log.Fatalf("Fotoğraf yüklenirken hata oluştu: %v", err)
	}

	log.Printf("Yüklenen fotoğraf bilgisi: %v", response)
	log.Printf("\n")
	log.Printf("\n")

	// GetImageDetail metodunu çağırarak fotoğraf detayını alır.
	detail, err := photoService.GetImageDetail(context.Background(), uploadedImage)
	if err != nil {
		log.Fatalf("Fotoğraf detayı alınırken hata oluştu: %v", err)
	}

	log.Printf("Fotoğraf detayı: %v", detail)
	log.Printf("\n")
	log.Printf("\n")
	// GetImageFeed metodunu çağırarak fotoğraf beslemesini alır.
	feedRequest := &photo.GetImageFeedRequest{
		PageSize:   10,
		PageNumber: 1,
	}
	feed, err := photoService.GetImageFeed(context.Background(), feedRequest)
	if err != nil {
		log.Fatalf("Fotoğraf beslemesi alınırken hata oluştu: %v", err)
	}

	log.Printf("Fotoğraf beslemesi: %v", feed)
	log.Printf("\n")
	log.Printf("\n")
	// Örnek bir güncellenecek fotoğraf detayı oluşturur.
	updatedDetail := &photo.UploadedImage{
		Id:  "1",
		Url: "https://i.dha.com.tr/i/dha/75/480x360/61811b5a45d2a01ab04762b9.jpg",

		UploadTime: time.Now().Unix(),
	}

	// UpdateImageDetail metodunu çağırarak fotoğraf detayını günceller.
	updatedDetailResponse, err := photoService.UpdateImageDetail(context.Background(), updatedDetail)
	if err != nil {
		log.Fatalf("Fotoğraf detayı güncellenirken hata oluştu: %v", err)
	}
	log.Printf("Güncellenmiş fotoğraf detayı: %v", updatedDetailResponse)
	log.Printf("\n")
	log.Printf("\n")

	// gRPC sunucu oluşturur.
	grpcServer := grpc.NewServer()

	// PhotoService sunucuya ekler.
	photoServiceServerImpl := &PhotoServiceServerImpl{photoService: photoService}
	photo.RegisterPhotoServiceServer(grpcServer, photoServiceServerImpl)

	log.Printf("gRPC sunucusu %s üzerinde dinleniyor", port)
	log.Printf("\n")
	log.Printf("\n")

	// Sunucuyu başlatır.
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}
