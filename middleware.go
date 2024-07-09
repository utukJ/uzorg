package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

// LoggingMiddleware logs the details of incoming requests and their processing time
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Check if the request method is POST
		if r.Method == "POST" {
			// Read the body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
				http.Error(w, "Error reading request body", http.StatusInternalServerError)
				return
			}

			// Log the body
			log.Printf("Request Body: %s", string(bodyBytes))

			// Since the body has been read, we need to recreate the io.Reader for further processing
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Log the incoming request details
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		// Process the request
		next.ServeHTTP(w, r)

		// Log the completion of request processing
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // No "Bearer " prefix found
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("UZORG_JWT_SECRET")), nil
		})

		if err != nil {
			log.Printf("JWT Token Parse error: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
			// Check for token expiry
			if err := claims.Valid(); err != nil {
				http.Error(w, "Expired token", http.StatusUnauthorized)
				return
			}

			// Token is valid and not expired. You can access claims like claims.Subject
			userID := claims.Subject

			ctx := context.WithValue(r.Context(), "userId", userID)
			r = r.WithContext(ctx)
			// Proceed with the next handler
			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
	})
}

func CMW(
	handler http.Handler,
	middlewares ...func(http.Handler) http.Handler,
) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
