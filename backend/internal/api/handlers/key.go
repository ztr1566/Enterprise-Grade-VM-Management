package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
	"go.uber.org/zap"
	"backend/internal/api"
	"backend/internal/audit"
	"backend/internal/crypto"
	"backend/internal/models"
)

// KeyHandler encapsulates SSH key vault endpoints.
type KeyHandler struct {
	DB *sql.DB
}

// CreateKeyRequest defines the payload for importing or generating an SSH key.
type CreateKeyRequest struct {
	Name       string `json:"name"`
	PrivateKey string `json:"private_key"` // Optional: PEM string. If empty, an Ed25519 key is generated.
}

// GetKeys handles listing all SSH key summaries (no private key data).
func (h *KeyHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := models.GetSSHKeys(h.DB)
	if err != nil {
		audit.Logger.Error("Failed to fetch SSH keys", zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to fetch SSH keys")
		return
	}
	if keys == nil {
		keys = []models.SSHKeySummary{} // return [] not null
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// CreateKey handles importing a PEM private key or generating a new Ed25519 key pair.
func (h *KeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		api.WriteError(w, http.StatusBadRequest, "Key name is required")
		return
	}

	var privatePEM []byte
	var publicKeyStr string

	if req.PrivateKey == "" {
		// Generate a new Ed25519 key pair
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			audit.Logger.Error("Failed to generate Ed25519 key", zap.Error(err))
			api.WriteError(w, http.StatusInternalServerError, "Failed to generate key pair")
			return
		}

		// Marshal private key to OpenSSH PEM format
		pemBlock, err := gossh.MarshalPrivateKey(privKey, "")
		if err != nil {
			audit.Logger.Error("Failed to marshal private key", zap.Error(err))
			api.WriteError(w, http.StatusInternalServerError, "Failed to encode private key")
			return
		}
		privatePEM = pem.EncodeToMemory(pemBlock)

		// Marshal public key to OpenSSH authorized_keys format
		sshPub, err := gossh.NewPublicKey(pubKey)
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, "Failed to encode public key")
			return
		}
		publicKeyStr = string(gossh.MarshalAuthorizedKey(sshPub))

	} else {
		// Import user-provided PEM
		privatePEM = []byte(req.PrivateKey)

		signer, err := gossh.ParsePrivateKey(privatePEM)
		if err != nil {
			audit.Logger.Warn("Invalid private key provided", zap.String("name", req.Name), zap.Error(err))
			api.WriteError(w, http.StatusBadRequest, "Invalid private key format: "+err.Error())
			return
		}
		publicKeyStr = string(gossh.MarshalAuthorizedKey(signer.PublicKey()))
	}

	// Encrypt the private key before persisting
	aesKey := crypto.MustGetAESKey()
	encrypted, err := crypto.Encrypt(privatePEM, aesKey)
	if err != nil {
		audit.Logger.Error("Failed to encrypt private key", zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to encrypt private key")
		return
	}

	key := models.SSHKey{
		ID:         uuid.New().String(),
		Name:       req.Name,
		PrivateKey: base64.StdEncoding.EncodeToString(encrypted),
		PublicKey:  publicKeyStr,
		CreatedAt:  time.Now().UTC(),
	}

	if err := models.CreateSSHKey(h.DB, key); err != nil {
		audit.Logger.Error("Failed to persist SSH key", zap.Error(err))
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			api.WriteError(w, http.StatusConflict, fmt.Sprintf("A key named '%s' already exists", req.Name))
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "Failed to store SSH key")
		return
	}

	audit.Logger.Info("SSH key stored in vault", zap.String("keyID", key.ID), zap.String("name", key.Name))

	// Return the summary (no private key) plus public key for convenience
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":         key.ID,
		"name":       key.Name,
		"public_key": key.PublicKey,
	})
}

// DeleteKey removes a key from the vault.
func (h *KeyHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		api.WriteError(w, http.StatusBadRequest, "Missing key ID")
		return
	}
	if err := models.DeleteSSHKey(h.DB, id); err != nil {
		audit.Logger.Error("Failed to delete SSH key", zap.String("keyID", id), zap.Error(err))
		api.WriteError(w, http.StatusInternalServerError, "Failed to delete SSH key")
		return
	}
	audit.Logger.Info("SSH key deleted from vault", zap.String("keyID", id))
	w.WriteHeader(http.StatusNoContent)
}

