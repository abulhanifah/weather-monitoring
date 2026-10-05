package handler

import (
	"net/http"

	"github.com/abulhanifah/weather-monitoring/pkg"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	pkg.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
