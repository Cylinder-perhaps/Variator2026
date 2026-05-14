package middleware

import (
	"encoding/json"
	"net/http"
)

// RequireRole проверяет, что роль пользователя входит в список разрешённых.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())

			for _, role := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)

			resp := map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "FORBIDDEN",
					"message": "Недостаточно прав для выполнения операции",
				},
			}
			json.NewEncoder(w).Encode(resp)
		})
	}
}
