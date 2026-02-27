package main

import (
	"log"
	"net"
)

func main() {
	
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Backend port dinleneme: %v", err)
	}
	log.Println("Backend stub :50051 portunda dinliyor...")

	// Bağlantıları kabul et (iş mantığı backend developer tarafından eklenecek)
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("Bağlantı hatası: %v", err)
			continue
		}
		conn.Close()
	}
}
