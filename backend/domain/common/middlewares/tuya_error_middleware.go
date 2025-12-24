package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// tuyaErrorResponseWriter is a custom ResponseWriter that captures the response body.
// It allows the middleware to inspect and modify the response before sending it to the client.
type tuyaErrorResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body bytes.
//
// param b The byte slice to write.
// return int The number of bytes written.
// return error An error if the write fails.
func (w *tuyaErrorResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// WriteString captures the response body string.
//
// param s The string to write.
// return int The number of bytes written.
// return error An error if the write fails.
func (w *tuyaErrorResponseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

// TuyaErrorMiddleware inspects the response body for specific Tuya error codes (e.g., 1010).
// If a token expiration error (code 1010) is detected, it intercepts the response and returns a standardized 401 Unauthorized error.
//
// return gin.HandlerFunc The Gin middleware handler.
func TuyaErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &tuyaErrorResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		responseBody := w.body.String()
		if strings.Contains(responseBody, "code: 1010") {
			utils.LogWarn("TuyaErrorMiddleware: Detected code 1010 (token invalid). Replacing response with 401.")
			newResponse := dtos.StandardResponse{
				Status:  false,
				Message: "Token expired. Please login or refresh the token",
				Data:    nil,
			}
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusUnauthorized)
			json.NewEncoder(w.ResponseWriter).Encode(newResponse)
		} else {
			w.ResponseWriter.Write(w.body.Bytes())
		}
	}
}