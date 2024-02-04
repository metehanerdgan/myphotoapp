Merhaba! Aşağıda, bir fotoğraf yükleme uygulamasının backend servisini oluşturan bir Go projesinin özetini bulacaksınız. Bu projede gRPC protokolü kullanılarak bir dizi hizmet sunulmaktadır. İşte projenin ana özellikleri:

1. gRPC Hizmetleri: photo_upload.proto dosyasında tanımlanan gRPC hizmetleri, fotoğraf yükleme, detayları alma, besleme alma ve detay güncelleme işlemlerini içermektedir.

2. Veritabanı Bağlantısı: db.go dosyasında, PostgreSQL veritabanına başarılı bir şekilde bağlantı kurulur ve gerekli tablo oluşturulur.

3. Kafka ile Asenkron İşlemler: kafka.go dosyasında, Kafka mesaj kuyruğuna mesaj gönderme işlemleri yapılmaktadır.

4. Fotoğraf Servisi: service.go dosyasında tanımlanan PhotoService yapısı, temel fotoğraf işleme ve yönetme fonksiyonlarını gerçekleştirir.

5. Yüz Analizi: Fotoğraf yükleme işlemi sırasında, VisionAPI kullanılarak fotoğraf içindeki yüzlerin analizi yapılmaktadır.

6. Konfigürasyon: config.go dosyasında, YAML formatında bulunan konfigürasyon dosyasından gerekli bilgiler okunmaktadır.

7. Ana Program: main.go dosyasında, gRPC sunucusu başlatılır ve örnek fotoğraflar yüklenerek servislerin kullanımı simüle edilir.

Bu projenin amacı, kullanıcıların fotoğraf yüklemelerini yönetmek ve bu yüklemeler üzerinde çeşitli işlemler gerçekleştirmektir. Duygu analizi vb. projenin farklı bölümleri arasında etkileşim, asenkron mesajlaşma ve dış servis entegrasyonları gibi pek çok önemli özellik bulunmaktadır.
