package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/common"
	"github.com/Fadlihardiyanto/football-api/internal/entity"
	messaging "github.com/Fadlihardiyanto/football-api/internal/gateway"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/model/converter"
	"github.com/Fadlihardiyanto/football-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthUseCase interface {
	Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error)
	Register(ctx context.Context, request *model.RegisterRequest) (*model.RegisterResponse, error)
	Refresh(ctx context.Context, request *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error)
	Logout(ctx context.Context, request *model.LogoutRequest) error
	ValidateToken(tokenString string) (*model.Auth, error)
	GetRefreshTokenExpiry() time.Duration
}

type authUseCaseImpl struct {
	UsersRepo    repository.UsersRepository
	LogsProducer *messaging.LogProducer
	DB           *gorm.DB
	Log          *logrus.Logger
	JwtConfig    *model.JWTConfig
	RedisClient  *redis.Client
}

func NewAuthUseCase(usersRepo repository.UsersRepository, logsProducer *messaging.LogProducer, db *gorm.DB, log *logrus.Logger, jwtConfig *model.JWTConfig, redisClient *redis.Client) AuthUseCase {
	return &authUseCaseImpl{
		UsersRepo:    usersRepo,
		LogsProducer: logsProducer,
		DB:           db,
		Log:          log,
		JwtConfig:    jwtConfig,
		RedisClient:  redisClient,
	}
}

func (a *authUseCaseImpl) Login(ctx context.Context, request *model.LoginRequest) (*model.LoginResponse, error) {
	tx := a.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		a.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	user := new(entity.User)

	existingUser, err := a.UsersRepo.FindByEmail(tx, request.Email)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			a.Log.Errorf("Failed to find user by email %s: %v", request.Email, err)
			tx.Rollback()
			return nil, common.ErrInternalServer("Failed to find user").WithDetail("email", request.Email)
		}
	}

	if existingUser == nil {
		a.Log.Warnf("User with email %s not found", request.Email)
		tx.Rollback()
		return nil, common.ErrNotFound("User not found").WithDetail("email", request.Email)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(request.Password)); err != nil {
		a.Log.Warnf("Invalid password for user %s: %v", request.Email, err)
		tx.Rollback()
		return nil, common.ErrUnauthorized("Invalid credentials").WithDetail("email", request.Email)
	}

	user = existingUser

	// Generate access and refresh tokens
	accessToken, refreshToken, refreshTokenID, accessExpiry, err := a.GenerateToken(existingUser.ID, existingUser.Role)
	if err != nil {
		a.Log.Warnf("Failed to generate token : %+v", err)
		return nil, common.ErrInternalServer("Failed to generate token")
	}

	// Store the refresh token in Redis
	if err := a.StoreToken(refreshToken, refreshTokenID, existingUser.ID); err != nil {
		a.Log.Errorf("Failed to store token: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to store token").WithDetail("email", existingUser.Email)
	}

	if err := tx.Commit().Error; err != nil {
		a.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction").WithDetail("email", existingUser.Email)
	}

	// Produce log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("User %s logged in successfully", user.Email),
		Service: "auth-service",
		Time:    time.Now().Format(time.RFC3339),
	}
	if err := a.LogsProducer.Send(logEvent); err != nil {
		a.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to log event").WithDetail("email", user.Email)
	}

	return converter.ToLoginResponse(user, accessToken, refreshTokenID, accessExpiry), nil
}

func (a *authUseCaseImpl) Register(ctx context.Context, request *model.RegisterRequest) (*model.RegisterResponse, error) {
	tx := a.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		a.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	// Check if user already exists
	existingUser, err := a.UsersRepo.FindByEmail(tx, request.Email)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			a.Log.Errorf("Failed to find user by email %s: %v", request.Email, err)
			tx.Rollback()
			return nil, common.ErrInternalServer("Failed to check existing user").WithDetail("email", request.Email)
		}
	}

	if existingUser != nil {
		a.Log.Warnf("User with email %s already exists", request.Email)
		tx.Rollback()
		return nil, common.ErrConflict("User already exists").WithDetail("email", request.Email)
	}

	password, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		a.Log.Warnf("Failed to generate bcrypt hash : %+v", err)
		return nil, common.ErrInternalServer("Failed to generate password hash").WithDetail("email", request.Email)
	}

	user := &entity.User{
		ID:           uuid.New().String(),
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: string(password),
		Role:         request.Role,
	}

	if err := a.UsersRepo.Create(tx, user); err != nil {
		tx.Rollback()
		a.Log.Errorf("Failed to create user: %v", err)
		return nil, common.ErrInternalServer("Failed to create user")
	}

	if err := tx.Commit().Error; err != nil {
		a.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction")
	}

	// Produce log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("User %s registered successfully", user.Email),
		Service: "auth-service",
		Time:    time.Now().Format(time.RFC3339),
	}
	if err := a.LogsProducer.Send(logEvent); err != nil {
		a.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to log event").WithDetail("email", user.Email)
	}

	return converter.ToRegisterResponse(user), nil
}

func (a *authUseCaseImpl) Refresh(ctx context.Context, request *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	tx := a.DB.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := common.ValidateStruct(request); err != nil {
		a.Log.Warnf("Invalid request body : %+v", err)
		return nil, err
	}

	// check if the refresh token exists in Redis
	isUsed, err := a.IsTokenUsed(request.ID)
	if err != nil {
		a.Log.Warnf("Failed to check if token is used : %+v", err)
		return nil, common.ErrInternalServer("Failed to check if token is used")
	}
	if isUsed {
		a.Log.Warnf("Token %s is already used", request.ID)
		tx.Rollback()
		return nil, common.ErrUnauthorized("Token is already used").WithDetail("id", request.ID)
	}

	// Get the refresh token from Redis
	token, err := a.GetToken(request.ID)
	if err != nil {
		a.Log.Warnf("Failed to get token : %+v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to get token").WithDetail("id", request.ID)
	}

	// check ttl of the token
	if ttl := a.GetTTLOfToken(request.ID); ttl <= 0 {
		a.Log.Warnf("Token expired")
		return nil, common.ErrUnauthorized("Token expired").WithDetail("id", request.ID)
	}

	// validate token
	auth, err := a.ValidateToken(token)
	if err != nil {
		a.Log.Warnf("Failed to validate token : %+v", err)
		tx.Rollback()
		return nil, common.ErrUnauthorized("Invalid token").WithDetail("id", request.ID)
	}

	// Generate new access and refresh tokens
	accessToken, refreshToken, refreshTokenID, accessExpiry, err := a.GenerateToken(auth.ID, auth.Role)
	if err != nil {
		a.Log.Warnf("Failed to generate token : %+v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to generate token")
	}

	// mark the old token as used
	if err := a.MarkTokenAsUsed(request.ID); err != nil {
		a.Log.Warnf("Failed to mark token as used : %+v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to mark token as used").WithDetail("id", request.ID)
	}

	// Store the new refresh token in Redis
	if err := a.StoreToken(refreshToken, refreshTokenID, auth.ID); err != nil {
		a.Log.Errorf("Failed to store token: %v", err)
		tx.Rollback()
		return nil, common.ErrInternalServer("Failed to store token").WithDetail("id", request.ID)
	}

	if err := tx.Commit().Error; err != nil {
		a.Log.Errorf("Failed to commit transaction: %v", err)
		return nil, common.ErrInternalServer("Failed to commit transaction").WithDetail("id", request.ID)
	}

	// Produce log event
	logEvent := &model.LogEvent{
		Level:   "info",
		Message: fmt.Sprintf("User %s refreshed tokens successfully", auth.ID),
		Service: "auth-service",
		Time:    time.Now().Format(time.RFC3339),
	}
	if err := a.LogsProducer.Send(logEvent); err != nil {
		a.Log.Errorf("Failed to send log event: %v", err)
		return nil, common.ErrInternalServer("Failed to log event").WithDetail("id", request.ID)
	}

	// Return the new access and refresh tokens
	return converter.ToRefreshTokenResponse(refreshTokenID, accessToken, accessExpiry), nil

}

func (a *authUseCaseImpl) Logout(ctx context.Context, request *model.LogoutRequest) error {
	tokenKey := fmt.Sprintf("token:%s", request.RefreshTokenID)
	usedKey := fmt.Sprintf("token:%s:used", request.RefreshTokenID)

	if err := a.RedisClient.Del(ctx, tokenKey).Err(); err != nil {
		a.Log.Warnf("Failed to delete token from Redis: %v", err)
		return err
	}

	_ = a.RedisClient.Del(ctx, usedKey).Err()

	return nil
}

func (a *authUseCaseImpl) GenerateToken(id string, role string) (string, string, string, time.Time, error) {

	tokenID := uuid.New().String()
	claims := &model.UserClaims{
		ID:      id,
		Role:    role,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.JwtConfig.AccessExpiry)),
			Issuer:    a.JwtConfig.Issuer,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := accessToken.SignedString([]byte(a.JwtConfig.SecretKey))
	if err != nil {
		a.Log.Warnf("Failed to generate access token : %+v", err)
		return "", "", "", time.Time{}, common.ErrInternalServer("Failed to generate access token")
	}

	expiresAccessToken := claims.ExpiresAt.Time

	refreshClaims := &model.UserClaims{
		ID:      id,
		Role:    role,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.JwtConfig.RefreshExpiry)),
			Issuer:    a.JwtConfig.Issuer,
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(a.JwtConfig.SecretKey))
	if err != nil {
		a.Log.Warnf("Failed to generate refresh token : %+v", err)
		return "", "", "", time.Time{}, common.ErrInternalServer("Failed to generate refresh token")
	}

	return accessTokenString, refreshTokenString, tokenID, expiresAccessToken, nil
}

func (c *authUseCaseImpl) ValidateToken(tokenString string) (*model.Auth, error) {
	var auth *model.Auth

	log.Println("tokenString: ", tokenString)

	token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(c.JwtConfig.SecretKey), nil
	})
	if err != nil {
		c.Log.Warnf("Failed to validate token : %+v", err)
		return nil, common.ErrUnauthorized("Invalid token").WithDetail("token", tokenString)
	}

	claims, ok := token.Claims.(*model.UserClaims)
	if !ok || !token.Valid {
		c.Log.Warnf("Failed to validate token : %+v", jwt.ErrInvalidKey)
		return nil, common.ErrUnauthorized("Invalid token").WithDetail("token", tokenString)

	}

	auth = &model.Auth{
		ID:   claims.ID,
		Role: claims.Role,
	}

	return auth, nil
}

func (c *authUseCaseImpl) GetToken(id string) (string, error) {
	ctx := context.Background()
	tokenKey := fmt.Sprintf("token:%s", id)
	token, err := c.RedisClient.Get(ctx, tokenKey).Result()
	if err != nil {
		c.Log.Warnf("Failed to get token : %+v", err)
		return "", err
	}

	return token, nil
}

func (c *authUseCaseImpl) GetTTLOfToken(id string) time.Duration {
	ctx := context.Background()
	tokenKey := fmt.Sprintf("token:%s", id)
	ttl, err := c.RedisClient.TTL(ctx, tokenKey).Result()
	if err != nil || ttl <= 0 {
		c.Log.Warnf("Failed to get token ttl or token expired: %v", err)
		return 0
	}
	return ttl
}

func (c *authUseCaseImpl) StoreToken(refreshToken string, refreshTokenID, userID string) error {
	ctx := context.Background()

	// store token to redis with expiry time
	tokenKey := fmt.Sprintf("token:%s", refreshTokenID)
	err := c.RedisClient.Set(ctx, tokenKey, refreshToken, c.JwtConfig.RefreshExpiry).Err()
	if err != nil {
		c.Log.Warnf("Failed to store token: %+v", err)
		return err
	}

	// save token to user token set
	userTokenSetKey := fmt.Sprintf("user:%s:tokens", userID)
	err = c.RedisClient.SAdd(ctx, userTokenSetKey, refreshTokenID).Err()
	if err != nil {
		c.Log.Warnf("Failed to add token to user set: %+v", err)
		return err
	}

	// set expiry for user token set
	_ = c.RedisClient.Expire(ctx, userTokenSetKey, c.JwtConfig.RefreshExpiry).Err()

	return nil
}

func (c *authUseCaseImpl) DeleteToken(token string) error {
	ctx := context.Background()
	tokenKey := fmt.Sprintf("token:%s", token)
	err := c.RedisClient.Del(ctx, tokenKey).Err()
	if err != nil {
		c.Log.Warnf("Failed to delete token : %+v", err)
		return err
	}

	return nil
}

func (c *authUseCaseImpl) IsTokenUsed(tokenID string) (bool, error) {
	ctx := context.Background()
	usedKey := fmt.Sprintf("token:%s:used", tokenID)
	exists, err := c.RedisClient.Exists(ctx, usedKey).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}

func (c *authUseCaseImpl) MarkTokenAsUsed(tokenID string) error {
	ctx := context.Background()
	// Set a key in Redis to mark the token as used
	usedKey := fmt.Sprintf("token:%s:used", tokenID)
	return c.RedisClient.Set(ctx, usedKey, true, 5*time.Minute).Err()
}

func (c *authUseCaseImpl) GetRefreshTokenExpiry() time.Duration {
	return c.JwtConfig.RefreshExpiry
}
