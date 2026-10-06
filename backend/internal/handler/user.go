package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/abulhanifah/weather-monitoring/internal/config"
	"github.com/abulhanifah/weather-monitoring/internal/models"
	"github.com/abulhanifah/weather-monitoring/internal/services"
	"github.com/abulhanifah/weather-monitoring/pkg"
	"github.com/golang-jwt/jwt/v5"
)

type UserHandler struct {
	cfg *config.Config
	svc *services.UserService
}

func NewUserHandler(cfg *config.Config, svc *services.UserService) *UserHandler {
	return &UserHandler{cfg: cfg, svc: svc}
}

// Login godoc
// @Summary Login dan dapatkan JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorEnvelope
// @Failure 401 {object} models.ErrorEnvelope
// @Router /api/v1/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Invalid payload",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		pkg.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": "Email and password are required",
		})
		return
	}

	user, err := h.svc.CheckLogin(r.Context(), req.Email, req.Password)
	if err != nil {
		pkg.WriteJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"error": "Invalid email or password",
		})
		return
	}

	claims := models.Claims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.cfg.AuthSecret))
	if err != nil {
		pkg.WriteJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to generate token",
		})
		return
	}

	pkg.WriteJSON(w, http.StatusOK, models.LoginResponse{Token: tokenString})
}
