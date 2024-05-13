package util

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var (
	firebaseApp       *firebase.App
	firebaseClient    *auth.Client
	firebaseMsgClient *messaging.Client
)

type FirebaseAuth struct {
	CredentialJSON string `yaml:"credential_json" json:"credential_json"`
}

func (c FirebaseAuth) GetFirebaseClient() *auth.Client {
	return firebaseClient
}

func (c FirebaseAuth) GetFirebaseMsgClient() *messaging.Client {
	return firebaseMsgClient
}

func (c FirebaseAuth) InitFirebase() {
	var err error
	if firebaseApp == nil {
		op := option.WithCredentialsJSON([]byte(c.CredentialJSON))
		firebaseApp, err = firebase.NewApp(context.Background(), nil, op)
		Panic(err)
	}

	if firebaseClient == nil {
		firebaseClient, err = firebaseApp.Auth(context.Background())
		Panic(err)
	}

	if firebaseMsgClient == nil {
		firebaseMsgClient, err = firebaseApp.Messaging(context.Background())
		Panic(err)
	}
}
