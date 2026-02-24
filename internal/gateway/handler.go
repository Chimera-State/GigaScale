package gateway

import (
	"encoding/json"
	"net/http"
)

func HandleReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Sadece POST istekleri kabul edilir.", http.StatusMethodNotAllowed)
		return
	}

	var req ReserveHTTPRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.printf("JSON Decoding Error %v", err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

}
