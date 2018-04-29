package middleware

import (
	"net/http"
	"strings"

	"muto/context"
	"muto/models"
)

// User middleware will lookup the current user via their
// remember_token cookie using the UserService. If the user
// is found, they will be set on the request context.
// Regardless, the next handler is always called.
type Account struct {
	models.AccountService
}

func (mw *Account) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn will return an http.HandlerFunc that checks to see
// if an account is logged in. if so, we will call next (w, r)
// otherwise we will redirect the guest to the login page.
func (mw *Account) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// If account is requesting a static assets
		// or image we will not need to lookup
		// current account, so we can skip it.
		if strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}

		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}

		account, err := mw.AccountService.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}

		// Get the context from our request.
		ctx := r.Context()
		// Create a new context from the existing one
		// that has our account stored in it
		// with the private account key.
		ctx = context.WithAccount(ctx, account)
		// Create a new request from the existing one
		// with our context attached to it and assign it back to 'r'.
		r = r.WithContext(ctx)
		// Call next(w, r) with our updated context.
		next(w, r)
	})
}

// RequireUser will redirect a user to the /login page
// if they are not logged in. This middleware assumes
// that User middleware has already been run, otherwise
// it will always redirect users.
type RequireAccount struct{}

func (mw *RequireAccount) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

func (mw *RequireAccount) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account := context.Account(r.Context())
		if account == nil {
			http.Redirect(w, r, "/enter", http.StatusFound)
			return
		}
		next(w, r)
	})
}
