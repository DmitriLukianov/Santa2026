package v1

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"secret-santa-backend/internal/controller/http/v1/response"
	"secret-santa-backend/internal/definitions"
	"secret-santa-backend/internal/dto"
	"secret-santa-backend/internal/helpers"
	"secret-santa-backend/internal/oauth"
	"secret-santa-backend/internal/usecase"
	authuc "secret-santa-backend/internal/usecase/auth"
)

type AuthHandler struct {
	provider    oauth.Provider
	jwt         *oauth.JWTManager
	uc          *authuc.UseCase
	userUC      usecase.UserUseCase
	frontendURL string
}

func NewAuthHandler(provider oauth.Provider, jwt *oauth.JWTManager, uc *authuc.UseCase, userUC usecase.UserUseCase, frontendURL string) *AuthHandler {
	return &AuthHandler{
		provider:    provider,
		jwt:         jwt,
		uc:          uc,
		userUC:      userUC,
		frontendURL: frontendURL,
	}
}

// generateState создаёт криптографически случайный state для OAuth CSRF-защиты.
func generateState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	// Сохраняем state в httpOnly cookie — проверим при callback
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 минут
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.provider.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем OAuth state против cookie (CSRF-защита)
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		response.WriteHTTPError(w, definitions.ErrInvalidOAuthCode)
		return
	}
	if r.URL.Query().Get("state") != stateCookie.Value {
		response.WriteHTTPError(w, definitions.ErrInvalidOAuthCode)
		return
	}
	// Удаляем использованный state-cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		response.WriteHTTPError(w, definitions.ErrInvalidOAuthCode)
		return
	}

	token, err := h.provider.Config().Exchange(ctx, code)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	userInfo, err := h.provider.GetUserInfo(ctx, token)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	userID, err := h.uc.LoginWithOAuth(ctx, userInfo)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	jwtToken, err := h.jwt.GenerateToken(userID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	redirectURL := h.frontendURL + "/auth/callback?token=" + jwtToken
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *AuthHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.SendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	if err := helpers.ValidateStruct(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	if err := h.uc.SendOTP(r.Context(), req.Email); err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Код отправлен на почту",
	})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}
	if err := helpers.ValidateStruct(&req); err != nil {
		response.WriteHTTPError(w, definitions.ErrInvalidUserInput)
		return
	}

	userID, err := h.uc.VerifyOTP(r.Context(), req.Email, req.Code, req.Name)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	jwtToken, err := h.jwt.GenerateToken(userID)
	if err != nil {
		response.WriteHTTPError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	parsedID, err := uuid.Parse(userID)
	if err == nil {
		user, err := h.userUC.GetByID(r.Context(), parsedID)
		if err == nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"token": jwtToken,
				"user":  response.UserResponse{ID: user.ID.String(), Name: user.Name, Email: user.Email},
			})
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"token": jwtToken})
}
