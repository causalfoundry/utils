package util

import (
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"google.golang.org/api/option"
	"net/http"

	"github.com/rs/zerolog"
	"google.golang.org/api/idtoken"
)

type UserPayload struct {
	Username  string
	Name      string
	FirstName string
	LastName  string
}

type ExceptFn func(token string) *UserPayload

//go:generate mockgen -destination=mock_jwtparser.go -package util . JwtParser
type JwtParser interface {
	TokenToPayload(token string) (payload UserPayload, err error)
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
	return
}
