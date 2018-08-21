package controllers

import (
	"fmt"
	"net/http"
	"time"

	"coverd/context"
	"coverd/models"
	"coverd/rand"
	"coverd/views"
)

func NewAccounts(as models.AccountService) *Accounts {
	return &Accounts{
		NewView:   views.NewView("materialize", "accounts/new"),
		LoginView: views.NewView("materialize", "accounts/enter"),
		as:        as,
	}
}

type Accounts struct {
	NewView   *views.View
	LoginView *views.View
	as        models.AccountService
}

type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

type RegisterForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// New is used to render the account registration form.
// GET /register
func (a *Accounts) New(w http.ResponseWriter, r *http.Request) {
	a.NewView.Render(w, r, nil)
}

// Create is used to process the form when creating a new account
// POST /register
func (a *Accounts) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form RegisterForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		a.NewView.Render(w, r, vd)
		return
	}

	account := models.Account{
		Email:    form.Email,
		Password: form.Password,
	}

	if err := a.as.Create(&account); err != nil {
		vd.SetAlert(err)
		a.NewView.Render(w, r, vd)
		return
	}

	err := a.signIn(w, &account)
	if err != nil {
		http.Redirect(w, r, "/enter", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// *Implement account index for public view in the pool.
// func (a *Accounts) Index(w http.ResponseWriter, r *http.Request) {
//
// }

// CookieTest displays cookie info on the screen to the current account.
func (a *Accounts) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	account, err := a.as.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, account)
}

// signIn is ised to sign the given account in via cookie.
func (a *Accounts) signIn(w http.ResponseWriter, account *models.Account) error {
	if account.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}

		account.Remember = token
		err = a.as.Update(account)
		if err != nil {
			return err
		}
	}

	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    account.Remember,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}

// Login is used to process the form
// when accessing an existing account.
// POST/enter
func (a *Accounts) Login(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form LoginForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		a.LoginView.Render(w, r, vd)
		return
	}

	account, err := a.as.Authenticate(form.Email, form.Password)
	if err != nil {
		switch err {
		case models.ErrNotFound:
			vd.AlertError("No account exists with that email address")
		default:
			vd.SetAlert(err)
		}
		a.LoginView.Render(w, r, vd)
		return
	}

	err = a.signIn(w, account)
	if err != nil {
		vd.SetAlert(err)
		a.LoginView.Render(w, r, vd)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// Logout is used to expire an account's session token,
// and update the account's remember_me token with a new random value
func (a *Accounts) Logout(w http.ResponseWriter, r *http.Request) {
	// Expire the accounts cookie.
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    "",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	// Update the account w/ the new remember token.
	account := context.Account(r.Context())
	// Ignore errors for now. Even if they do occur,
	// we cant recover now that the account doesn't have a valid cookie.
	token, _ := rand.RememberToken()
	account.Remember = token
	a.as.Update(account)
	// Redirect the account to the landing page.
	http.Redirect(w, r, "/", http.StatusFound)
}
