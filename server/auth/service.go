package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"mig"
	"mig/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func (a *Auther) NewAccessToken(user mig.User, audience []JwtAudience, sessionID string) (string, Claims, error) {
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
