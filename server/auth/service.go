package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"mig"
	"mig/repository"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	UserID          string                `json:"user_id"`
	Username        string                `json:"username"`
	Email           string                `json:"email"`
	UserRole        mig.UserRole          `json:"user_role"`
	UserFingerprint string                `json:"user_fingerprint"`
	WorkflowState   mig.UserWorkflowState `json:"workflow_state"`
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

func GetHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (a *Auther) newAccessToken(user mig.User) (string, string, *Claims, error) {
	str, err := generateRandomString(32)
	if err != nil {
		return "", "", nil, fmt.Errorf("unable to create access token: %w", err)
	}

	userFingerprint := url.QueryEscape(str)
	userFingerprintHash := GetHash(userFingerprint)
	now := time.Now()

	claims := Claims{
		UserID:          user.ID,
		Username:        user.Username,
		Email:           user.Email,
		WorkflowState:   user.WorkflowState,
		UserFingerprint: userFingerprintHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    a.issuer,
			Subject:   user.ID,
			ID:        user.ID,
			Audience:  jwt.ClaimStrings{string(user.Role)},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(a.secret)
	if err != nil {
		return "", "", nil, fmt.Errorf("unable to create access token: %w", err)
	}

	return ss, userFingerprint, &claims, nil
}

// newRefreshToken generates a base64-encoded random string of length 32 as an opaque refresh token.
func newRefreshToken() (string, error) {
	str, err := generateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("unable to create refresh token: %w", err)
	}
	return str, nil
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("unable to create random string: %w", err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func (a *Auther) verifyAccessToken(token string) (bool, *Claims, error) {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return a.secret, nil
	})

	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return false, nil, fmt.Errorf("unable to parse access token: %w", err)
	}

	if claims, ok := t.Claims.(*Claims); ok {
		return t.Valid, claims, nil
	}

	return false, nil, fmt.Errorf("invalid access token")
}

func (s *Service) VerifyAccessToken(token string) (bool, *Claims, error) {
	return s.auther.verifyAccessToken(token)
}

type LoginResponse struct {
	AccessToken     string   `json:"access_token"`
	ExpiresAt       int64    `json:"expires_at"`
	ExpiresIn       int      `json:"expires_in"`
	RefreshToken    string   `json:"refresh_token"`
	User            mig.User `json:"user"`
	UserFingerprint string   `json:"-"`
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

	accessToken, userFingerprint, claims, err := s.auther.newAccessToken(user)
	if err != nil {
		return LoginResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	refreshToken, err := newRefreshToken()
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
		AccessToken:     accessToken,
		ExpiresAt:       claims.ExpiresAt.Unix(),
		ExpiresIn:       int(claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()),
		RefreshToken:    token.Token,
		User:            user,
		UserFingerprint: userFingerprint,
	}

	return resp, nil
}

type SignupResponse struct {
	AccessToken     string   `json:"access_token"`
	ExpiresAt       int64    `json:"expires_at"`
	ExpiresIn       int      `json:"expires_in"`
	RefreshToken    string   `json:"refresh_token"`
	User            mig.User `json:"user"`
	UserFingerprint string   `json:"-"`
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
		Role:          mig.MemberUserRole,
	}

	accessToken, userFingerprint, claims, err := s.auther.newAccessToken(user)
	if err != nil {
		return SignupResponse{}, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError)
	}

	refreshToken, err := newRefreshToken()
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
		AccessToken:     accessToken,
		ExpiresAt:       claims.ExpiresAt.Unix(),
		ExpiresIn:       int(claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()),
		RefreshToken:    refreshToken,
		User:            user,
		UserFingerprint: userFingerprint,
	}

	return resp, nil
}

type RefreshTokenResponse struct {
	AccessToken     string   `json:"access_token"`
	ExpiresAt       int64    `json:"expires_at"`
	ExpiresIn       int      `json:"expires_in"`
	RefreshToken    string   `json:"refresh_token"`
	User            mig.User `json:"user"`
	UserFingerprint string
}

func (s *Service) RefreshToken(ctx context.Context, accessToken, refreshToken, fingerprint string) (RefreshTokenResponse, error) {
	_, claims, err := s.VerifyAccessToken(accessToken)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	hash := GetHash(fingerprint)

	if hash != claims.UserFingerprint {
		return RefreshTokenResponse{}, fmt.Errorf("invalid user fingerprint")
	}

	refresh, err := s.userRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	if time.Now().After(refresh.ExpiresAt) {
		return RefreshTokenResponse{}, fmt.Errorf("refresh token expired")
	}

	user, err := s.userRepo.GetUser(ctx, refresh.UserID)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	accessToken, userFingerprint, claims, err := s.auther.newAccessToken(user)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	resp := RefreshTokenResponse{
		AccessToken:     accessToken,
		ExpiresAt:       claims.ExpiresAt.Unix(),
		ExpiresIn:       int(claims.ExpiresAt.Unix() - claims.IssuedAt.Unix()),
		User:            user,
		RefreshToken:    refresh.Token,
		UserFingerprint: userFingerprint,
	}

	return resp, nil
}
