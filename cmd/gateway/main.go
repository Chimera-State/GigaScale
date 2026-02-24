package main

import (
	"net/http"

	"github.com/Chimera-State/GigaScale/internal/gateway"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("api/v1/reserve", gateway.HandleReserve)

	http.ListenAndServe(":8080", mux)
}
