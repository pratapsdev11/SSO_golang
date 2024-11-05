package services

import (
	"context"
	"errors"
	"time"
	"your-project/internal/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type AuthService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAuthService(db *gorm.DB, redis *redis.Client) *AuthService {
	return &AuthService{
		db:    db,
		redis: redis,
	}
}

func (s *AuthService) AuthenticateWithMicrosoft(ctx context.Context, token string) (*models.User, string, error) {
	// Verify token with Microsoft Entra ID
	msUser, err := s.verifyMicrosoftToken(token)
	if err != nil {
		return nil, "", err
	}

	// Find or create user
	var user models.User
	result := s.db.Where("entra_id = ?", msUser.ID).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Create new user
			user = models.User{
				Email:   msUser.Email,
				Name:    msUser.Name,
				EntraID: msUser.ID,
			}
			if err := s.db.Create(&user).Error; err != nil {
				return nil, "", err
			}
		} else {
			return nil, "", result.Error
		}
	}

	// Generate session token
	sessionToken, err := s.createSession(ctx, user.ID)
	if err != nil {
		return nil, "", err
	}

	return &user, sessionToken, nil
}

func (s *AuthService) createSession(ctx context.Context, userID uint) (string, error) {
	token := generateRandomToken()

	// Store session in Redis
	err := s.redis.Set(ctx, token, userID, 24*time.Hour).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	// Get user ID from Redis
	userID, err := s.redis.Get(ctx, token).Uint64()
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// Get user from database
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
