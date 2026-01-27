package httpx

import "net/http"

// CorsMiddleware enables access from the given origin.
// This is just a common way to use it.
// For more complex options and more refined configurations, the user should define its own middleware instead.
func CorsMiddleware(origin string, next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", origin)
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func WildcardCorsMiddleware(next http.Handler) http.Handler {
	return CorsMiddleware("*", next)
}
