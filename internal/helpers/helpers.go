package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/joshua-takyi/auction/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidatePassword(pass string) bool {
	// Check minimum length
	if len(pass) <= 8 {
		return false
	}

	var (
		hasLower   = false
		hasUpper   = false
		hasDigit   = false
		hasSpecial = false
	)

	for _, char := range pass {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '@' || char == '$' || char == '!' || char == '%' || char == '*' || char == '#' || char == '?' || char == '&':
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}

func ValidateProductInput(product *models.Product) error {
	if product.Title == "" {
		return errors.New("title is required")
	}
	if product.Description == "" {
		return errors.New("description is required")
	}

	if product.Price <= 0 {
		return errors.New("price must be greater than 0")
	}
	if len(product.Images) == 0 {
		return errors.New("images are required")
	}
	if len(product.Details) == 0 {
		return errors.New("details are required")
	}
	return nil
}

func GenerateSlug(title, category string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "&", "and")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "", "")
	return s
}

func GenerateCsrfToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
