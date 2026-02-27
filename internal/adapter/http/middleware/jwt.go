package middleware

import (
	"context"
	"net/http"
	"strings"
	"task_vault/internal/app/auth"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

func JWTAuth(jwt *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"отсутствует токен авторизации"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				http.Error(w, `{"error":"неверный формат токена"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwt.Validate(token)
			if err != nil {
				http.Error(w, `{"error":"невалидный токен"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}
