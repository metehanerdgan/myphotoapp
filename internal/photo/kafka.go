package photo

import (
	"fmt"
	"log"
	"myphotoapp/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var kafkaBrokers string

func init() {
	initKafkaConfig()
}

func initKafkaConfig() {
	configPath := "config/config.yaml"
	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Konfigürasyon dosyası okunamadı: %v", err)
	}

	kafkaBrokers = config.Kafka.Broker
}

// KafkaProducer, Kafka'ya mesaj gönderme işlemlerini yöneten bir yapıdır.
type KafkaProducer struct {
	producer *kafka.Producer
}

// NewKafkaProducer, yeni bir KafkaProducer örneği oluşturur.
func NewKafkaProducer() (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaBrokers})
	if err != nil {
		return nil, fmt.Errorf("Kafka üretici oluşturulamadı: %v", err)
	}

	log.Printf("Kafka üretici başlatıldı")

	return &KafkaProducer{producer: p}, nil
}

// ProduceMessage, Kafka'ya mesaj gönderir.
func (kp *KafkaProducer) ProduceMessage(topic string, message string) error {
	err := kp.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	if err != nil {
		return fmt.Errorf("Mesaj gönderilemedi: %v", err)
	}

	return nil
}

// Close, Kafka üreticisini kapatır.
func (kp *KafkaProducer) Close() {
	kp.producer.Close()
}
