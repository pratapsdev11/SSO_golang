package services

import (
    "context"
    "errors"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/google/uuid"
    "gorm.io/gorm"
    
    "your-project/internal/models"
)

type SessionService struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewSessionService(db *gorm.DB, redis *redis.Client) *SessionService {
    return &SessionService{
        db:    db,
        redis: redis,
    }
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(ctx context.Context, userID uint) (*models.Session, error) {
    // Generate a unique session token
    token := uuid.New().String()
    
    // Set session expiration time (24 hours from now)
    expiresAt := time.Now().Add(24 * time.Hour)
    
    // Create session record in database
    session := &models.Session{
        UserID:    userID,
        Token:     token,
        ExpiresAt: expiresAt,
    }
    
    if err := s.db.Create(session).Error; err != nil {
        return nil, fmt.Errorf("failed to create session: %w", err)
    }
    
    // Store session in Redis for quick lookups
    err := s.redis.Set(ctx, 
        fmt.Sprintf("session:%s", token),
        userID,
        time.Until(expiresAt),
    ).Err()
    if err != nil {
        // If Redis fails, delete the session from database
        s.db.Delete(session)
        return nil, fmt.Errorf("failed to store session in cache: %w", err)
    }
    
    return session, nil
}

// GetSession retrieves a session by token
func (s *SessionService) GetSession(ctx context.Context, token string) (*models.Session, error) {
    // Try to get from Redis first
    userID, err := s.redis.Get(ctx, fmt.Sprintf("session:%s", token)).Uint64()
    if err == nil {
        // Found in Redis, get full session from database
        var session models.Session
        if err := s.db.Where("token = ? AND user_id = ?", token, userID).First(&session).Error; err != nil {
            return nil, err
        }
        return &session, nil
    }
    
    // If not in Redis, try database
    var session models.Session
    if err := s.db.Where("token = ?", token).First(&session).Error; err != nil {
        return nil, err
    }
    
    // If found in database but not in Redis, check if expired
    if time.Now().After(session.ExpiresAt) {
        s.DeleteSession(ctx, session.Token, session.UserID)
        return nil, errors.New("session expired")
    }
    
    // Repopulate Redis
    err = s.redis.Set(ctx,
        fmt.Sprintf("session:%s", token),
        session.UserID,
        time.Until(session.ExpiresAt),
    ).Err()
    if err != nil {
        // Log error but don't fail the request
        fmt.Printf("failed to repopulate session in cache: %v\n", err)
    }
    
    return &session, nil
}

// DeleteSession removes a session
func (s *SessionService) DeleteSession(ctx context.Context, token string, userID uint) error {
    // Delete from Redis
    err := s.redis.Del(ctx, fmt.Sprintf("session:%s", token)).Err()
    if err != nil {
        fmt.Printf("failed to delete session from cache: %v\n", err)
    }
    
    // Delete from database
    result := s.db.Where("token = ? AND user_id = ?", token, userID).Delete(&models.Session{})
    if result.Error != nil {
        return fmt.Errorf("failed to delete session: %w", result.Error)
    }
    
    if result.RowsAffected == 0 {
        return errors.New("session not found")
    }
    
    return nil
}

// CleanExpiredSessions removes all expired sessions
func (s *SessionService) CleanExpiredSessions(ctx context.Context) error {
    // Delete expired sessions from database
    result := s.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
    if result.Error != nil {
        return fmt.Errorf("failed to clean expired sessions: %w", result.Error)
    }
    
    // Note: Redis keys will expire automatically based on TTL
    
    return nil
}

// RefreshSession extends the session lifetime
func (s *SessionService) RefreshSession(ctx context.Context, token string) error {
    // Get existing session
    session, err := s.GetSession(ctx, token)
    if err != nil {
        return err
    }
    
    // Update expiration time
    newExpiresAt := time.Now().Add(24 * time.Hour)
    
    // Update database
    if err := s.db.Model(session).Update("expires_at", newExpiresAt).Error; err != nil {
        return fmt.Errorf("failed to refresh session in database: %w", err)
    }
    
    // Update Redis
    err = s.redis.Expire(ctx,
        fmt.Sprintf("session:%s", token),
        time.Until(newExpiresAt),
    ).Err()
    if err != nil {
        return fmt.Errorf("failed to refresh session in cache: %w", err)
    }
    
    return nil
}

// ValidateSession checks if a session is valid
func (s *SessionService) ValidateSession(ctx context.Context, token string) (uint, error) {
    session, err := s.GetSession(ctx, token)
    if err != nil {
        return 0, err
    }
    
    if time.Now().After(session.ExpiresAt) {
        s.DeleteSession(ctx, token, session.UserID)
        return 0, errors.New("session expired")
    }
    
    return session.UserID, nil
}