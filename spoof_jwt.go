package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
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

	superSecretHex := os.Getenv("SWIZZLE_SUPER_SECRET")
	superSecret, err := hex.DecodeString(superSecretHex)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to decode super secret from hex", http.StatusInternalServerError)
		return
	}

	// Read the JSON file into a byte slice
	fileData, err := ioutil.ReadFile(SECRETS_FILE_PATH)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed reading secrets.json file", http.StatusInternalServerError)
		return
	}

	var result map[string]interface{}

	err = json.Unmarshal(fileData, &result)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to parse secret.json file", http.StatusInternalServerError)
		return
	}

	jwtSigningSecret := result["test"].(map[string]interface{})["SWIZZLE_JWT_SECRET_KEY"].(map[string]interface{})
	jwtSigningSecretCipherBase64 := jwtSigningSecret["cipher"].(string)
	jwtSigningSecretIv := jwtSigningSecret["iv"].(string)

	jwtSigningSecretCipher, err := base64.StdEncoding.DecodeString(jwtSigningSecretCipherBase64)
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to decode the JWT cipher from base64", http.StatusInternalServerError)
		return
	}

	decrypted, err := decryptAES(jwtSigningSecretCipher, superSecret, []byte(jwtSigningSecretIv))
	if err != nil {
		log.Println("Error:", err)
		http.Error(w, "Failed to decrypt the JWT signing secret from AES", http.StatusInternalServerError)
		return
	}

	jwtSecretKey := string(decrypted)

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 24).Unix(), // Expires in 24 hours
	})

	tokenString, err := token.SignedString([]byte(jwtSecretKey))

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
