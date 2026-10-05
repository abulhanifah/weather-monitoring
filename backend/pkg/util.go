package pkg

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
)

// Fungsi generik 'contains' menggunakan constraint 'comparable'
// T mewakili tipe data dinamis yang akan ditentukan saat fungsi dipanggil
func Contains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func GetCryptoRandomBytes() ([]byte, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("crypto rand error: %w", err)
	}
	return bytes, nil
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
