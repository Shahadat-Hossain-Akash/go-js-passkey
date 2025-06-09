package data

import (
	"go-js-passkey/models"

	"github.com/go-webauthn/webauthn/webauthn"
)

type PasskeyStore interface {
	GetUserByEmail(userName string) (*models.PasskeyUser, error)
	GetUserByID(ID int) (*models.PasskeyUser, error)
	SaveUser(models.PasskeyUser)
	GenSessionID() (string, error)
	GetSession(token string) (webauthn.SessionData, bool)
	SaveSession(token string, data webauthn.SessionData)
	DeleteSession(token string)
}
