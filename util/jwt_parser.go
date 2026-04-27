package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"google.golang.org/api/idtoken"
)

type UserPayload struct {
	Username  string
	Email     string
	Name      string
	FirstName string
	LastName  string
}

type ExceptFn func(token string) *UserPayload

//go:generate mockgen -destination=mock_jwtparser.go -package util . JwtParser
type JwtParser interface {
	TokenToPayload(token string) (payload UserPayload, err error)
}

// ---------------- ms -------------------
type MicroSoftJwtParser struct {
	log              zerolog.Logger
	exceptFn         ExceptFn
	jwksURL          string
	allowedAudiences map[string]struct{}
	allowedIssuers   map[string]struct{}
}

var _ JwtParser = MicroSoftJwtParser{}

type MicrosoftJwtParserConfig struct {
	AllowedAudiences []string
	AllowedIssuers   []string
	ExceptFn         ExceptFn
	JwksURL          string
}

var microsoftValidMethods = []string{"RS256"}

func NewMicrosfotJwtParser(exceptFn ExceptFn) MicroSoftJwtParser {
	return newMicrosoftJwtParser(MicrosoftJwtParserConfig{
		ExceptFn: exceptFn,
	})
}

func NewMicrosoftJwtParser(cfg MicrosoftJwtParserConfig) (MicroSoftJwtParser, error) {
	if len(cfg.AllowedAudiences) == 0 && len(cfg.AllowedIssuers) == 0 {
		return MicroSoftJwtParser{}, errors.New("microsoft jwt parser requires at least one allowed audience or issuer")
	}
	return newMicrosoftJwtParser(cfg), nil
}

const MicrosoftJwksUrl = "https://login.microsoftonline.com/common/discovery/v2.0/keys"

func newMicrosoftJwtParser(cfg MicrosoftJwtParserConfig) MicroSoftJwtParser {
	jwksURL := cfg.JwksURL
	if jwksURL == "" {
		jwksURL = MicrosoftJwksUrl
	}

	return MicroSoftJwtParser{
		log:              NewLogger("auth.microsoft-jwt-parser"),
		exceptFn:         cfg.ExceptFn,
		jwksURL:          jwksURL,
		allowedAudiences: toStringSet(cfg.AllowedAudiences),
		allowedIssuers:   toStringSet(cfg.AllowedIssuers),
	}
}

func toStringSet(values []string) map[string]struct{} {
	ret := make(map[string]struct{}, len(values))
	for _, value := range values {
		ret[value] = struct{}{}
	}
	return ret
}

func (g MicroSoftJwtParser) validateClaims(claims jwt.MapClaims) error {
	if len(g.allowedAudiences) != 0 {
		audiences, err := claims.GetAudience()
		if err != nil {
			return fmt.Errorf("error get audience claim: %w", err)
		}

		var found bool
		for _, aud := range audiences {
			if _, ok := g.allowedAudiences[aud]; ok {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unexpected audience: %v", []string(audiences))
		}
	}

	if len(g.allowedIssuers) != 0 {
		issuer, err := claims.GetIssuer()
		if err != nil {
			return fmt.Errorf("error get issuer claim: %w", err)
		}
		if _, ok := g.allowedIssuers[issuer]; !ok {
			return fmt.Errorf("unexpected issuer: %s", issuer)
		}
	}

	return nil
}

// TokenToUsername implements JwtParser
func (g MicroSoftJwtParser) TokenToPayload(token string) (ret UserPayload, err error) {
	if g.exceptFn != nil {
		if payload := g.exceptFn(token); payload != nil {
			ret = *payload
			return
		}
	}

	// Get the JSON Web Key Sets (JWKS) from Microsoft
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	funcs, err := keyfunc.NewDefaultCtx(ctx, []string{g.jwksURL})
	if err != nil {
		return
	}

	jwtToken, err := jwt.Parse(token, funcs.Keyfunc, jwt.WithValidMethods(microsoftValidMethods))
	if err != nil {
		return
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("unexpected microsoft claims type: %T", jwtToken.Claims)
		return
	}
	if err = g.validateClaims(claims); err != nil {
		err = NewErr(http.StatusUnauthorized, err.Error(), nil)
		return
	}
	ret.Email = fmt.Sprint(claims["preferred_username"])
	ret.Name = fmt.Sprint(claims["name"])
	ret.Username = fmt.Sprint(claims["preferred_username"])
	return
}

// ---------------- google -------------------
type GoogleJwtParser struct {
	log      zerolog.Logger
	ClientID string
	exceptFn ExceptFn
}

var _ JwtParser = GoogleJwtParser{}

func NewGoogleJwtParser(clientID string, exceptFn ExceptFn) GoogleJwtParser {
	return GoogleJwtParser{
		log:      NewLogger("auth.google-jwt-parser"),
		ClientID: clientID,
		exceptFn: exceptFn,
	}
}

// TokenToUsername implements JwtParser
func (g GoogleJwtParser) TokenToPayload(token string) (ret UserPayload, err error) {
	if g.exceptFn != nil {
		if payload := g.exceptFn(token); payload != nil {
			ret = *payload
			return
		}
	}

	payload, err := idtoken.Validate(context.Background(), token, g.ClientID)

	if err != nil {
		g.log.Err(err).Str("token", token).Str("client_id", g.ClientID).Msg("error validate jwt token")
		err = NewErr(http.StatusUnauthorized, err.Error(), nil)
		return
	}

	ret.Username = fmt.Sprint(payload.Claims["email"])
	ret.Name = fmt.Sprint(payload.Claims["name"])
	ret.FirstName = fmt.Sprint(payload.Claims["given_name"])
	ret.LastName = fmt.Sprint(payload.Claims["family_name"])

	return
}

// ---------------- firebase -------------------
type FirebaseJwtParser struct {
	log        zerolog.Logger
	authClient *auth.Client
	exceptFn   ExceptFn
}

var _ JwtParser = FirebaseJwtParser{}

func NewFirebaseJwtParser(jsonCredential string, exceptFn ExceptFn) FirebaseJwtParser {
	op := option.WithCredentialsJSON([]byte(jsonCredential))
	firebaseApp, err := firebase.NewApp(context.Background(), nil, op)
	Panic(err)

	client, err := firebaseApp.Auth(context.Background())
	Panic(err)

	return FirebaseJwtParser{
		log:        NewLogger("auth.firebase-jwt-parser"),
		authClient: client,
		exceptFn:   exceptFn,
	}
}

// TokenToUsername implements JwtParser
func (g FirebaseJwtParser) TokenToPayload(token string) (ret UserPayload, err error) {
	if g.exceptFn != nil {
		if payload := g.exceptFn(token); payload != nil {
			ret = *payload
			return
		}
	}

	verifiedToken, err := g.authClient.VerifyIDToken(context.Background(), token)
	if err != nil {
		g.log.Err(err).Str("token", token).Msg("error validate jwt token")
		return
	}

	ret.Username = verifiedToken.UID
	if email, ok := verifiedToken.Claims["email"]; ok {
		ret.Email = fmt.Sprint(email)
	}
	return
}
