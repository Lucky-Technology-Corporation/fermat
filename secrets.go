package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
)

type GetSecretsResponse struct {
	Test map[string]string `json:"test"`
	Prod map[string]bool   `json:"prod"` // True if secret is set false otherwise
}

type Secrets struct {
	Test map[string]string `json:"test"`
	Prod map[string]string `json:"prod"`
}

func GetSecrets(w http.ResponseWriter, r *http.Request) {
	secrets, err := ReadSecretsFromFile()
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed reading secrets.json", http.StatusInternalServerError)
		return
	}

	err = WriteJSONResponse(w, secrets)
	if err != nil {
		http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
		return
	}
}

func UpdateSecrets(w http.ResponseWriter, r *http.Request) {
	var newSecrets Secrets
	err := json.NewDecoder(r.Body).Decode(&newSecrets)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	file, err := os.OpenFile(SECRETS_FILE_PATH, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		http.Error(w, "Failed to open secrets.json file for writing", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	err = newSecrets.SaveSecrets(file)
	if err != nil {
		http.Error(w, "Failed to write secrets.json", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parsePrivateKeyFromParam(key string) (*rsa.PrivateKey, error) {
	privateKeyDecoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyDecoded)
	if block == nil {
		return nil, errors.New("Failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func parsePublicKeyFromParam(key string) (*rsa.PublicKey, error) {
	publicKeyDecoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(publicKeyDecoded)
	if block == nil {
		return nil, errors.New("Failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return publicKey, nil
}

func ReadSecretsFromFile() (*Secrets, error) {
	file, err := os.Open(SECRETS_FILE_PATH)
	if err != nil {
		return nil, err
	}

	secrets, err := ReadSecrets(file)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func ReadSecrets(in io.Reader) (*Secrets, error) {
	decoder := json.NewDecoder(in)

	var secrets Secrets
	if err := decoder.Decode(&secrets); err != nil {
		return nil, err
	}
	return &secrets, nil
}

func (secrets Secrets) SaveSecrets(out io.Writer) error {
	encoder := json.NewEncoder(out)

	if err := encoder.Encode(&secrets); err != nil {
		return err
	}
	return nil
}
