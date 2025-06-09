package data

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-js-passkey/logger"
	"go-js-passkey/models"
	"strconv"

	"github.com/go-webauthn/webauthn/webauthn"
)

type PassKeyRepository struct {
	db       *sql.DB
	sessions map[string]webauthn.SessionData
	log      logger.Logger
}

func NewPasskeyRepository(db *sql.DB, log logger.Logger) *PassKeyRepository {
	return &PassKeyRepository{
		db:       db,
		sessions: make(map[string]webauthn.SessionData),
		log:      log,
	}
}

func (r *PassKeyRepository) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", nil
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (r *PassKeyRepository) GetSession(token string) (webauthn.SessionData, bool) {
	r.log.Info(fmt.Sprintf("Get Session: %v", r.sessions[token]))
	val, ok := r.sessions[token]
	return val, ok
}

func (r *PassKeyRepository) SaveSession(token string, data webauthn.SessionData) {
	r.log.Info(fmt.Sprintf("SaveSession: %s - %v", token, data))
	r.sessions[token] = data
}

func (r *PassKeyRepository) DeleteSession(token string) {
	r.log.Info(fmt.Sprintf("DeleteSession: %v", token))
	delete(r.sessions, token)
}

func (r *PassKeyRepository) GetUserByEmail(email string) (*models.PasskeyUser, error) {
	r.log.Info(fmt.Sprintf("Get User: %v", email))

	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		r.log.Error("Failed to find new user", err)
		return nil, err
	} else if err != nil {
		r.log.Error("Failed to query user", err)
		return nil, err
	}
	rows, err := r.db.Query("SELECT keys from passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query passkeys", err)
		return nil, err
	}
	defer rows.Close()
	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			return nil, err
		}
		cred, err := deserializeCredential(keys)
		if err != nil {
			r.log.Error("Failed to deserialize credential", err)
			continue // Skip invalid credentials
		}
		credentials = append(credentials, cred)
	}
	user := models.PasskeyUser{
		ID:          []byte(strconv.Itoa(userID)),
		DisplayName: email,
		Name:        email,
		Credentials: credentials,
	}
	return &user, nil

}

func (r *PassKeyRepository) GetUserByID(id int) (*models.PasskeyUser, error) {
	r.log.Info(fmt.Sprintf("Get User: %v", id))

	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&userID)
	if err == sql.ErrNoRows {
		r.log.Error("Failed to find new user", err)
		return nil, err
	} else if err != nil {
		r.log.Error("Failed to query user", err)
		return nil, err
	}

	rows, err := r.db.Query("SELECT keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query passkeys", err)
		return nil, err
	}
	defer rows.Close()

	var email string
	err = r.db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Error("Failed to find user email", err)
			return nil, err
		}
		r.log.Error("Failed to query user email", err)
		return nil, err
	}

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			return nil, err
		}
		cred, err := deserializeCredential(keys)
		if err != nil {
			r.log.Error("Failed to deserialize credential", err)
			continue // Skip invalid credentials
		}
		credentials = append(credentials, cred)
	}

	user := models.PasskeyUser{
		ID:          []byte(strconv.Itoa(userID)),
		Name:        email,
		DisplayName: email,
		Credentials: credentials,
	}
	return &user, nil
}

func (r *PassKeyRepository) SaveUser(user models.PasskeyUser) {
	r.log.Info(fmt.Sprintf("SaveUser: %v", user.WebAuthnName()))
	
	userID, err := strconv.Atoi(string(user.ID))
	if err != nil {
		r.log.Error("Invalid user ID", err)
		return
	}

	// Insert new credentials
	for _, cred := range user.Credentials {
		keys, err := serializeCredential(cred)
		if err != nil {
			r.log.Error("Failed to serialize credential", err)
			continue
		}
		// Check if the key already exists in the database
		var exists bool
		err = r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM passkeys WHERE user_id = $1 AND keys = $2)", userID, keys).Scan(&exists)
		if err != nil {
			r.log.Error("Failed to check if passkey exists", err)
			continue
		}

		// Insert the key only if it does not already exist
		if !exists {
			_, err = r.db.Exec("INSERT INTO passkeys (user_id, keys) VALUES ($1, $2)", userID, keys)
			if err != nil {
				r.log.Error("Failed to insert passkey", err)
			}
		} else {
			r.log.Info(fmt.Sprintf("Passkey already exists for user_id: %d", userID))
		}
	}
}

func serializeCredential(credential webauthn.Credential) (string, error) {
	data, err := json.Marshal(credential)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}
	return string(data), nil
}

func deserializeCredential(credential string) (webauthn.Credential, error) {
	var cred webauthn.Credential
	err := json.Unmarshal([]byte(credential), &cred)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("failed to unmarshal credential: %w", err)
	}
	return cred, nil
}
