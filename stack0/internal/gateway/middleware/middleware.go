package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/elfeddi/dxp/internal/gateway/rbac"
)

type contextKey string

const roleKey contextKey = "dxp_role"

// RoleFromContext extrait le rôle injecté par le middleware Auth.
func RoleFromContext(ctx context.Context) (rbac.Role, bool) {
	v, ok := ctx.Value(roleKey).(rbac.Role)
	return v, ok
}

// ErrorResponse est la structure JSON retournée en cas d'erreur.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
}

// WriteError écrit une réponse d'erreur JSON standardisée.
func WriteError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(code),
		Code:    code,
		Message: msg,
	})
}

// Logger logue chaque requête avec durée et status code.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start).String(),
			"remote", r.RemoteAddr,
		)
	})
}

// Recovery intercepte les panics et retourne un 500 propre.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", fmt.Sprintf("%v", rec),
					"stack", string(debug.Stack()),
				)
				WriteError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Auth vérifie le token Bearer et injecte le rôle dans le contexte.
// POC : le token EST le rôle ("Bearer admin"). Production : JWT/OIDC.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			next.ServeHTTP(w, r)
			return
		}
		token := extractBearer(r)
		if token == "" {
			WriteError(w, http.StatusUnauthorized, "missing Authorization header")
			return
		}
		role, err := rbac.RoleFromString(token)
		if err != nil {
			WriteError(w, http.StatusForbidden, fmt.Sprintf("unknown role: %s", token))
			return
		}
		ctx := context.WithValue(r.Context(), roleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS ajoute les headers CORS pour Backstage.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Chain chaîne plusieurs middlewares dans l'ordre donné.
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func extractBearer(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
