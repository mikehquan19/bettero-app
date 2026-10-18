package controllers

import (
	"betterov2/models"
	"betterov2/services"
	"context"
	"errors"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const sessionCookieName = "bettero_session"

// AuthController handles registration, login, and logout requests.
type AuthController struct {
	service *services.AuthService
}

// NewAuthController creates an AuthController backed by the given auth service.
func NewAuthController(service *services.AuthService) *AuthController {
	return &AuthController{
		service: service,
	}
}

// Register creates a user and starts a session by setting the session cookie.
func (a *AuthController) Register(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.RegisterUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	user, token, err := a.service.Register(ctx, body)
	if err != nil {
		if errors.Is(err, models.ErrInvalidUserBody) {
			respondError(c, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, models.ErrUserExists) {
			respondError(c, http.StatusConflict, err)
			return
		}
		respondAuthInternal(c, err)
		return
	}

	a.setCookie(c, token)
	respondSuccess(c, http.StatusCreated, user)
}

// Login verifies the user's credentials and sets a session cookie.
func (a *AuthController) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var body models.LoginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		respondError(c, http.StatusBadRequest, err)
		return
	}

	user, token, err := a.service.Login(ctx, body)
	if err != nil {
		if isAuthenticationError(err) {
			respondError(c, http.StatusUnauthorized, models.ErrInvalidCredentials)
			return
		}
		respondAuthInternal(c, err)
		return
	}

	a.setCookie(c, token)
	respondSuccess(c, http.StatusOK, user)
}

// Logout ends the current session and clears the session cookie.
func (a *AuthController) Logout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	token, _ := c.Cookie(sessionCookieName)
	if err := a.service.Logout(ctx, token); err != nil {
		respondAuthInternal(c, err)
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	respondSuccess(c, http.StatusOK, gin.H{"status": "logged_out"})
}

func (a *AuthController) setCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(services.SessionDuration.Seconds()),
		Expires:  time.Now().Add(services.SessionDuration),
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

func secureCookies() bool {
	value := os.Getenv("SESSION_COOKIE_SECURE")
	if value == "" {
		return true
	}
	secure, err := strconv.ParseBool(value)
	return err != nil || secure
}

// AuthMiddleware authenticates requests using the session cookie.
// Health and auth endpoints are public; other requests receive the
// authenticated user's ID in the Gin context.
func AuthMiddleware(service *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		validEndpoints := []string{"/health", "/auth/register", "/auth/login", "/auth/logout"}
		if slices.Contains(validEndpoints, c.Request.URL.Path) {
			c.Next()
			return
		}

		token, err := c.Cookie(sessionCookieName)
		if err != nil {
			abortAuthentication(c, models.ErrInvalidCredentials)
			return
		}

		user, err := service.Authenticate(ctx, token)
		if err != nil {
			abortAuthentication(c, err)
			return
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}
