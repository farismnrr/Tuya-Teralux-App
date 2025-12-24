package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// GenerateTuyaSignature calculates the HMAC-SHA256 signature required for Tuya API requests.
// It constructs the message by concatenating clientID, accessToken, timestamp, and the stringToSign.
//
// Message Structure: clientID + accessToken + timestamp + stringToSign
//
// param clientID The Tuya Client ID.
// param clientSecret The Tuya Client Secret (used as the HMAC key).
// param accessToken The current access token (can be empty for token retrieval).
// param timestamp The current timestamp in milliseconds.
// param stringToSign The constructed string representing request details (method, hash, url).
// return string The uppercased hexadecimal signature.
func GenerateTuyaSignature(clientID, clientSecret, accessToken, timestamp, stringToSign string) string {
	// Concatenate: client_id + access_token + t + stringToSign
	message := clientID + accessToken + timestamp + stringToSign

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(clientSecret))
	h.Write([]byte(message))
	signature := h.Sum(nil)

	// Convert to uppercase hexadecimal
	return strings.ToUpper(hex.EncodeToString(signature))
}

// GenerateTuyaStringToSign creates the canonical string used as part of the signature calculation.
// It follows a specific format defined by Tuya's authentication protocol.
//
// Format: HTTPMethod + "\n" + ContentHash + "\n" + Headers + "\n" + URL
//
// param httpMethod The HTTP method (GET, POST, etc.).
// param contentHash The SHA256 hash of the request body (or empty string hash for GET).
// param headers The canonical headers string (often empty).
// param url The request URL path.
// return string The formatted string to sign.
func GenerateTuyaStringToSign(httpMethod, contentHash, headers, url string) string {
	return httpMethod + "\n" + contentHash + "\n" + headers + "\n" + url
}