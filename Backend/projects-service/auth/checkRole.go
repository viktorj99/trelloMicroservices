package auth

import "net/http"

func EnableManager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ParseToken(r)
		if err != nil || claims.Role != "Manager" {
			http.Error(w, "Forbidden: Manager role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func EnableMember(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ParseToken(r)
		if err != nil || claims.Role != "Member" {
			http.Error(w, "Forbidden: Member role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func EnableBoth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := auth.ParseToken(r)
		if err != nil || (claims.Role != "Manager" && claims.Role != "Member") {
			http.Error(w, "Forbidden: Manager or Member role required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
