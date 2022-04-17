package jwt

import (
	"context"
	"errors"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/infra/remotecache"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"gopkg.in/square/go-jose.v2/jwt"
)

const ServiceName = "AuthService"

func ProvideService(cfg *setting.Cfg, remoteCache *remotecache.RemoteCache) (*AuthService, error) {
	s := newService(cfg, remoteCache)
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func newService(cfg *setting.Cfg, remoteCache *remotecache.RemoteCache) *AuthService {
	return &AuthService{
		Cfg:         cfg,
		RemoteCache: remoteCache,
		log:         log.New("auth.jwt"),
	}
}

func (s *AuthService) init() error {
	if !s.Cfg.JWTAuthEnabled {
		return nil
	}

	if err := s.initClaimExpectations(); err != nil {
		return err
	}
	if err := s.initKeySet(); err != nil {
		return err
	}

	return nil
}

type AuthService struct {
	Cfg         *setting.Cfg
	RemoteCache *remotecache.RemoteCache

	keySet           keySet
	log              log.Logger
	expect           map[string]interface{}
	expectRegistered jwt.Expected
}

func (s *AuthService) Verify(ctx context.Context, strToken string) (models.JWTClaims, error) {
	s.log.Debug("Parsing JSON Web Token")

	token, err := jwt.ParseSigned(strToken)
	if err != nil {
		return nil, err
	}

	keys, err := s.keySet.Key(ctx, token.Headers[0].KeyID)
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, errors.New("no keys found")
	}

	s.log.Debug("Trying to verify JSON Web Token using a key")

	var claims models.JWTClaims
	for _, key := range keys {
		if err = token.Claims(key, &claims); err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	s.log.Debug("Validating JSON Web Token claims")

	if err = s.validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}
