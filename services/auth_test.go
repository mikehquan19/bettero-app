package services

import (
	"betterov2/models"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAuthServiceRegisterValidation(t *testing.T) {
	tests := []struct {
		name string
		body models.RegisterUserBody
	}{
		{
			name: "missing first name",
			body: models.RegisterUserBody{
				LastName: "Tester",
				Username: "tester",
				Password: "password123",
			},
		},
		{
			name: "missing last name",
			body: models.RegisterUserBody{
				FirstName: "Test",
				Username:  "tester",
				Password:  "password123",
			},
		},
		{
			name: "username is too short",
			body: models.RegisterUserBody{
				FirstName: "Test",
				LastName:  "Tester",
				Username:  "ab",
				Password:  "password123",
			},
		},
		{
			name: "email is too long",
			body: models.RegisterUserBody{
				FirstName: "Test",
				LastName:  "Tester",
				Username:  "tester",
				Email:     strings.Repeat("a", 101),
				Password:  "password123",
			},
		},
		{
			name: "password is too short",
			body: models.RegisterUserBody{
				FirstName: "Test",
				LastName:  "Tester",
				Username:  "tester",
				Password:  "short",
			},
		},
	}

	service := &AuthService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := service.Register(context.Background(), tt.body)
			if !errors.Is(err, models.ErrInvalidUserBody) {
				t.Fatalf("Register() error = %v, expected %v", err, models.ErrInvalidUserBody)
			}
		})
	}
}

func TestAuthServiceAuthenticateEmptyToken(t *testing.T) {
	service := &AuthService{}

	_, err := service.Authenticate(context.Background(), "")

	if !errors.Is(err, models.ErrNotFound) {
		t.Fatalf("Authenticate() error = %v, expected %v", err, models.ErrNotFound)
	}
}

func TestNewSessionToken(t *testing.T) {
	token, hash, err := newSessionToken()
	if err != nil {
		t.Fatalf("newSessionToken() error = %v", err)
	}
	if token == "" {
		t.Error("token is empty")
	}
	if len(hash) != 32 {
		t.Errorf("hash length = %d, expected 32", len(hash))
	}
}
