package authz

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"errors"
	"fmt"
	"github.com/IndexStorm/common-go/nanoid"
	"github.com/IndexStorm/common-go/service/authz/session"
	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type jwtClaims struct {
	jwt.RegisteredClaims
	Data interface{} `json:"data,omitempty"`
}

type jwtRawClaims struct {
	jwt.RegisteredClaims
	Data json.RawMessage `json:"data"`
}

type jwtService struct {
	jwtPublicKey   crypto.PublicKey
	jwtPrivateKey  crypto.PrivateKey
	jwtParser      *jwt.Parser
	sessionService session.Service
	config         JWTServiceConfig
}

type JWTServiceConfig struct {
	Issuer   string
	Audience []string
}

func NewJWTService(keySeed []byte, sessionService session.Service, config JWTServiceConfig) Service {
	privKey := ed25519.NewKeyFromSeed(keySeed)
	pubkey := privKey.Public()
	return &jwtService{
		jwtPublicKey:   pubkey,
		jwtPrivateKey:  privKey,
		jwtParser:      jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()})),
		sessionService: sessionService,
		config:         config,
	}
}

func (s *jwtService) IssueAuthorizationToken(
	ctx context.Context, request IssueAuthorizationTokenRequest,
) (*IssueAuthorizationTokenResponse, error) {
	t := time.Now()
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        nanoid.RandomLongID(),
			Issuer:    s.config.Issuer,
			Audience:  s.config.Audience,
			ExpiresAt: jwt.NewNumericDate(t.Add(time.Hour * 24 * 7)),
			NotBefore: jwt.NewNumericDate(t.Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(t),
		},
		Data: request.Claims,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := token.SignedString(s.jwtPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}
	err = s.sessionService.InitializeSession(ctx, session.InitializeSessionRequest{
		SessionID: claims.ID,
		UserID:    request.UserID,
		ExpiresAt: claims.ExpiresAt.Time,
	})
	if err != nil {
		return nil, fmt.Errorf("init session: %w", err)
	}
	return &IssueAuthorizationTokenResponse{
		Token:     signed,
		TokenID:   claims.ID,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

func (s *jwtService) ParseAuthorizationToken(ctx context.Context, request ParseAuthorizationTokenRequest) error {
	var rawClaims jwtRawClaims
	_, err := s.jwtParser.ParseWithClaims(
		request.Token,
		&rawClaims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, errors.New("method is not jwt.SigningMethodEd25519")
			}
			return s.jwtPublicKey, nil
		},
	)
	if err != nil {
		return fmt.Errorf("parse token: %w", err)
	}
	if err = json.Unmarshal(rawClaims.Data, &request.Claims); err != nil {
		return fmt.Errorf("unmarshal claims data: %w", err)
	}
	if err = s.sessionService.ValidateSession(ctx, session.ValidateSessionRequest{SessionID: rawClaims.ID}); err != nil {
		return fmt.Errorf("validate session: %w", err)
	}
	return nil
}

func (s *jwtService) RevokeAuthorizationToken(
	ctx context.Context, request RevokeAuthorizationTokenRequest,
) error {
	return s.sessionService.InvalidateSession(ctx, session.InvalidateSessionRequest{SessionID: request.TokenID})
}
