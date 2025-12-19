package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"teralux_app/dtos"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

type tuyaErrorResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *tuyaErrorResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *tuyaErrorResponseWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func TuyaErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &tuyaErrorResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()

		responseBody := w.body.String()
		if strings.Contains(responseBody, "code: 1010") {
			utils.LogWarn("TuyaErrorMiddleware: Detected code 1010 (token invalid). Replacing response with 401.")
			// Replace with standardized 401 response
			newResponse := dtos.StandardResponse{
				Status:  false,
				Message: "Token expired. Please login or refresh the token",
				Data:    nil,
			}
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusUnauthorized)
			json.NewEncoder(w.ResponseWriter).Encode(newResponse)
		} else {
			// Write original response
			w.ResponseWriter.Write(w.body.Bytes())
		}
	}
}
