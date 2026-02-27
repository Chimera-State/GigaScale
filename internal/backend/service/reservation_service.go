package service

import (
	"context"

	"github.com/Chimera-State/GigaScale/internal/backend/reservationv1"
)

// ReservationService, gRPC servisimizi uygulayan struct'tır.
type ReservationService struct {
	// İleriye dönük uyumluluk (forward compatibility) için protoc tarafından
	// üretilen bu struct'ı gömmemiz zorunludur.
	reservationv1.UnimplementedReservationServiceServer
}

// NewReservationService, yeni bir ReservationService oluşturup döndürür.
func NewReservationService() *ReservationService {
	return &ReservationService{}
}

// ReserveSeat, protobuf'ta tanımladığımız RPC metodunun gerçek uygulamasıdır.
func (s *ReservationService) ReserveSeat(ctx context.Context, req *reservationv1.ReserveSeatRequest) (*reservationv1.ReserveSeatResponse, error) {
	// 1. Gelen İstekteki Verileri Okuma
	// İstemci (client) bu fonksiyonu çağırdığında, gönderdiği veriler `req` nesnesinin içindedir.
	// Proto dosyasındaki tanımlarınıza göre bu verileri `Get...()` fonksiyonlarıyla alırız.
	userID := req.GetUserId()
	tripID := req.GetTripId()
	seatID := req.GetSeatId()
	idempotencyKey := req.GetIdempotencyKey()

	// Şimdilik sadece log'a yazdıralım.
	// (Gerçekte burada Redis/Veritabanı işlemi yapacaksınız)
	println("Gelen Rezervasyon İsteği:")
	println("- Kullanıcı ID:", userID)
	println("- Trip ID:", tripID)
	println("- Koltuk ID:", seatID)
	println("- Idempotency Key:", idempotencyKey)

	// 2. İş Mantığınızı (Business Logic) Uygulama
	// TODO: İleride burada veritabanına ya da Redis'e bağlanıp, "bu koltuk boş mu?" diye kontrol edeceğiz (SETNX işlemi vb.).
	// Şimdilik sahte (dummy) bir yanıt dönüyoruz, yani her işlem her zaman başarılıymış gibi davranıyoruz.
	isSuccess := true
	responseMessage := "İşlem Başarılı"

	// 3. İstemciye Yanıt (Response) Döndürme
	// Proto dosyanızdaki `ReserveSeatResponse` mesajına (struct'a) karşılık gelen
	// verileri istemciye gRPC üzerinden dönüyoruz.
	return &reservationv1.ReserveSeatResponse{
		Success: isSuccess,
		Message: responseMessage,
	}, nil
}
