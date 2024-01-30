Vision API , Kafka ile FOTOĞRAF ANALİZ UYGULAMASI gRPC API
Anonim bir fotoğraf yükleme uygulamasının backend servislerini yapıyor.

Fotoğraf Yükleme (UploadImage):

Kullanıcılar sisteme yeni fotoğraflar yükleyebilir.
Yüklenen fotoğrafın detayları (örneğin, tarih, kullanıcı bilgileri) kaydedilir.
Yüz analizi işlemi için bir asenkron görev başlatılır.
Yüz Analizi (AnalyzeFaces):

Asenkron olarak başlatılan yüz analizi işlemi, yüklenen fotoğrafın Vision API kullanılarak analiz edilmesini sağlar.
Yüz analizi sonuçları, fotoğrafın duygusal içeriği, yüz sayısı ve güvenilirlik puanı gibi bilgileri içerir.
Bu bilgiler Kafka üzerinden diğer sistem bileşenlerine iletilir.
Fotoğraf Detaylarını Güncelleme (UpdateImageDetail):

Yüz analizi sonuçları alındığında, bu bilgiler yüklenen fotoğrafın detaylarına eklenir.
Örneğin, fotoğraftaki yüz sayısı, duygusal içerik ve güvenilirlik puanı gibi bilgiler güncellenir.
Fotoğraf Detaylarını Alma (GetImageDetail):

Kullanıcılar belirli bir fotoğrafın detaylarını isteyebilir.
Bu istek sonucunda, fotoğrafın yüklenme tarihi, kullanıcı bilgileri ve yüz analizi sonuçları gibi detaylar geri döner.
Fotoğraf Besleme Alma (GetImageFeed):

Kullanıcılar sistemdeki fotoğrafları tarih ve yüz analizi sonuçlarına göre sıralı bir besleme şeklinde alabilirler.
Bu, kullanıcılara sistemdeki en güncel ve belirli kriterlere uygun fotoğrafları gösterme imkanı sağlar.
Bu yapı, geliştirilen uygulamanın temel işlevselliğini açıklar. Kullanıcılar fotoğraf yükleyebilir, yüklenen fotoğrafların detayları Vision API kullanılarak analiz edilir ve sonuçlar asenkron olarak sistemde güncellenir. Kullanıcılar daha sonra yüklenen fotoğrafların detaylarını ve güncel beslemeyi alabilirler. Kafka, yüz analizi sonuçlarını asenkron olarak diğer sistem bileşenlerine iletmek için kullanılır.
