package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func spoofJwt(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		http.Error(w, "Invalid user_id specified", http.StatusBadRequest)
		return
	}

	superSecretBase64 := os.Getenv("SWIZZLE_SUPER_SECRET")
	superSecret, err := ParseBase64PrivateKey(superSecretBase64)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to read super secret", http.StatusInternalServerError)
		return
	}

	secrets, err := ReadSecretsFromFile()
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to read secrets", http.StatusInternalServerError)
		return
	}

	err = secrets.DecryptSecrets(superSecret, nil)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to decrypt secrets", http.StatusInternalServerError)
		return
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	})

	tokenString, err := token.SignedString([]byte(secrets.Test["SWIZZLE_JWT_SECRET_KEY"]))

	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to sign the JWT", http.StatusInternalServerError)
		return
	}

	data := &map[string]string{
		"jwt": tokenString,
	}

	err = WriteJSONResponse(w, data)
	if err != nil {
		http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
		return
	}
}
