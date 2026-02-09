package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RequireAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//fmt.Println("Executing the middleware with request data ", r)

		tokenString, err := r.Cookie("access_token")

		if err != nil {
			log.Println(err)
			http.Error(w, "Cookie value missing or added incorrectly. The key should be access_token", http.StatusBadRequest)
			return
		}

		token, err := jwt.Parse(tokenString.Value, func(t *jwt.Token) (interface{}, error) {

			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_KEY")), nil
		})

		if err != nil {
			log.Println(err)

			http.Error(w, "Token could not be parsed", http.StatusBadRequest)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(w, "invalid token claims", http.StatusUnauthorized)
			return
		}

		if float64(time.Now().Unix()) > claims["exp"].(float64) {

			http.Error(w, "Token is expired. Please try again with new token", http.StatusBadRequest)
			return

		}
		username := claims["sub"].(string)

		ctx := context.WithValue(r.Context(), "username", username)

		next.ServeHTTP(w, r.WithContext(ctx))

	})

}
