package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"mig"
	"mig/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type JwtAudience string

const (
	JwtAudienceAuthenticated JwtAudience = "authenticated"
)

type Service struct {
	userRepo repository.UserRepository
	auther   *Auther
}

func NewService(userRepo repository.UserRepository, auther *Auther) (*Service, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("missing user repository")
	}

	if auther == nil {
		return nil, fmt.Errorf("missing auther")
	}

	service := &Service{
		userRepo: userRepo,
		auther:   auther,
	}

	return service, nil
}

type Auther struct {
	secret []byte
	issuer string
}

type Claims struct {
	ID            string                `json:"id"`
	Username      string                `json:"username"`
	Email         string                `json:"email"`
	WorkflowState mig.UserWorkflowState `json:"workflow_state"`
	jwt.RegisteredClaims
}

func NewAuther(secret, issuer string) (*Auther, error) {
	if secret == "" {
		return nil, fmt.Errorf("missing jwt secret")
	}

	if issuer == "" {
		return nil, fmt.Errorf("missing jwt issuer")
	}

	auther := &Auther{
		secret: []byte(secret),
		issuer: issuer,
	}

	return auther, nil
}

func (a *Auther) NewAccessToken(user mig.User, audience []JwtAudience) (string, Claims, error) {
	aud := make([]string, len(audience))
	for _, str := range aud {
		aud = append(aud, string(str))
	}

	claims := Claims{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		WorkflowState: user.WorkflowState,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.issuer,
			Subject:   user.ID,
			ID:        user.ID,
			Audience:  aud,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(a.secret)
	if err != nil {
		return "", Claims{}, fmt.Errorf("unable to create access token: %w", err)
	}

	return ss, claims, nil
}

// NewRefreshToken generates a base64-encoded random string of length 32 as an opaque refresh token.
func (a *Auther) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("unable to create refresh token: %w", err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// parseJwtToken checks validity of token and returns user id.
func (a *Auther) parseJwtToken(token string) (string, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.secret, nil
	})

	if err != nil {
		return "", fmt.Errorf("unable to parse token: %w", err)
	}

	if !t.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if claims, ok := t.Claims.(*Claims); ok {
		return claims.ID, nil
	}

	return "", fmt.Errorf("unable to validate jwt claims")
}

// ParseJwtToken checks validity of token and returns user id.
func (s *Service) ParseJwtToken(token string) (string, error) {
	return s.auther.parseJwtToken(token)
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresAt    int64    `json:"expires_at"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	User         mig.User `json:"user"`
}

func (s *Service) Login(ctx context.Context, username, password string) (LoginResponse, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return LoginResponse{}, mig.NewError("Incorrect username or password.", err, mig.UnauthorizedError)
	}

	passwordHash, err := s.userRepo.GetPassword(ctx, user.ID)
	if err != nil {
		return LoginResponse{}, mig.NewError("Incorrect username or password.", err, mig.UnauthorizedError)
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return LoginResponse{}, mig.NewError("Incorrect username or password.", err, mig.UnauthorizedError)
	}

	accessToken, claims, err := s.auther.NewAccessToken(user, []JwtAudience{JwtAudienceAuthenticated})
	if err != nil {
		return LoginResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	refreshToken, err := s.auther.NewRefreshToken()
	if err != nil {
		return LoginResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	arg := repository.UpsertRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Revoked:   false,
	}

	token, err := s.userRepo.UpsertRefreshToken(ctx, arg)
	if err != nil {
		return LoginResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	resp := LoginResponse{
		AccessToken:  accessToken,
		ExpiresAt:    claims.ExpiresAt.Unix(),
		ExpiresIn:    int(claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()),
		RefreshToken: token.Token,
		User:         user,
	}

	return resp, nil
}

type SignupResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresAt    int64    `json:"expires_at"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	User         mig.User `json:"user"`
}

func (s *Service) Signup(ctx context.Context, username, email, password string) (SignupResponse, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err == nil {
		return SignupResponse{}, mig.NewError("User already exits.", err, mig.InvalidInputError)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	user, err = s.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return SignupResponse{}, mig.NewError("User already exits.", err, mig.InvalidInputError)
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	userID := uuid.New().String()

	user = mig.User{
		ID:            userID,
		Email:         email,
		Username:      username,
		WorkflowState: mig.UserWorkflowStateUnverified,
	}

	accessToken, claims, err := s.auther.NewAccessToken(user, []JwtAudience{JwtAudienceAuthenticated})
	if err != nil {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	refreshToken, err := s.auther.NewRefreshToken()
	if err != nil {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	arg := repository.CreateUserWithRefreshTokenParams{
		ID:            userID,
		Email:         email,
		Username:      username,
		WorkflowState: user.WorkflowState,
		Password:      string(passwordHash),
		RefreshToken:  refreshToken,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	user, _, err = s.userRepo.CreateUserWithRefreshToken(ctx, arg)
	if err != nil {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	resp := SignupResponse{
		AccessToken:  accessToken,
		ExpiresAt:    claims.ExpiresAt.Unix(),
		ExpiresIn:    int(claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()),
		RefreshToken: refreshToken,
		User:         user,
	}

	return resp, nil
}
