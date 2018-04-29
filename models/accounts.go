package models

import (
	"regexp"
	"strings"

	"muto/hash"
	"muto/rand"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

// Errors
const (
	// ErrNotFound is returned when a resource cannot be found
	// in the database.
	ErrNotFound          modelError = "models: resource not found"
	ErrIDInvalid         modelError = "models: ID provided was invalid"
	ErrEmailRequired     modelError = "models: email address is required"
	ErrEmailInvalid      modelError = "models: email address is not valid"
	ErrEmailTaken        modelError = "models: email address is taken"
	ErrPasswordIncorrect modelError = "models: incorrect password provided"
	ErrPasswordTooShort  modelError = "models: password must be at least 8 characters long"
	ErrPasswordRequired  modelError = "models: password is required"
	ErrRememberRequired  modelError = "models: remember token is required"
	ErrRememberTooShort  modelError = "models: remember token must be at least 32 bytes"
)

// Test to verify accountGorm implements the AccountDB interface.
var _ AccountDB = &accountGorm{}

// Test to verufy userService imlements the AccountService interface
var _ AccountService = &accountService{}

// AccountService interface is a set of methods used to manipulate
// and work with the account model.
type AccountService interface {
	// Authenticate will verify the provided email
	// and password are correct.
	// If correct, the proper account will be returned.
	// If incorrect, the following errors may be returned:
	// ErrNotFound, ErrPasswordIncorrect, or related error if something else.
	// *Reference accountService "Authenticate" func for more detail on errors.
	Authenticate(email, password string) (*Account, error)
	AccountDB
}

// AccountDB is used to interact with the accounts database.
type AccountDB interface {
	// Methods for querying single accounts
	ByID(id uint) (*Account, error)
	ByEmail(email string) (*Account, error)
	ByRemember(token string) (*Account, error)

	// Methods for altering accounts
	Create(account *Account) error
	Update(account *Account) error
	Delete(id uint) error
}

type Account struct {
	gorm.Model
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// accountGorm represents our database interaction layer
// and implements the AccountDB interface fully.
type accountGorm struct {
	db *gorm.DB
}

// modelError implements a public error message.
type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

// NewAccountService
func NewAccountService(db *gorm.DB, pepper, hmacKey string) AccountService {
	ag := &accountGorm{db}
	hmac := hash.NewHMAC(hmacKey)
	av := newAccountValidator(ag, hmac, pepper)
	return &accountService{
		AccountDB: av,
		pepper:    pepper,
	}
}

type accountService struct {
	AccountDB
	pepper string
}

func newAccountValidator(adb AccountDB, hmac hash.HMAC, pepper string) *accountValidator {
	return &accountValidator{
		AccountDB: adb,
		hmac:      hmac,
		pepper:    pepper,
		emailRegex: regexp.MustCompile(
			`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
	}
}

// accountValidator is our validation layer that validates and normalizes data
// before passing it tothe AccountDB in our interface chain.
type accountValidator struct {
	AccountDB
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
	pepper     string
}

// Authenticate can be used to authenticate an account with the
// provided email address and password.
// If the email address provided is invalid, this will return
// nil, ErrNotFound
// If the password provided is invalid, this will return
// nil, ErrPasswordIncorrect
// If the email and password are both valid, this will return
// user, nil
// Otherwise if another error is encountered this will return
// nil, error
func (as *accountService) Authenticate(email, password string) (*Account, error) {
	// Locate account via email.
	foundAccount, err := as.ByEmail(email)
	if err != nil {
		return nil, err
	}

	// Compare hashed password and password provided.
	err = bcrypt.CompareHashAndPassword(
		[]byte(foundAccount.PasswordHash),
		[]byte(password+as.pepper))

	// Handle errors for each outcome.
	switch err {
	case nil:
		return foundAccount, nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return nil, ErrPasswordIncorrect
	default:
		return nil, err
	}
}

// VALIDATION - bcryptPassword will hash an accounts's password with an
// app-wide pepper and bcrypt, which salts for us.
func (av *accountValidator) bcryptPassword(account *Account) error {
	if account.Password == "" {
		// We DO NOT need to run this if the password
		// hasn't been changed.
		return nil
	}
	pwBytes := []byte(account.Password + av.pepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes,
		bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	account.PasswordHash = string(hashedBytes)
	account.Password = ""
	return nil
}

// VALIDATION - hmacRemember
func (av *accountValidator) hmacRemember(account *Account) error {
	if account.Remember == "" {
		return nil
	}
	account.RememberHash = av.hmac.Hash(account.Remember)
	return nil
}

// VALIDATION - setRememberIfUnset
func (av *accountValidator) setRememberIfUnset(account *Account) error {
	if account.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return nil
	}
	account.Remember = token
	return nil
}

// VALIDATION - idGreaterThan
func (av *accountValidator) idGreaterThan(n uint) accountValFn {
	return accountValFn(func(account *Account) error {
		if account.ID <= n {
			return ErrIDInvalid
		}
		return nil
	})
}

// VALIDATION - normalizeEmail
func (av *accountValidator) normalizeEmail(account *Account) error {
	account.Email = strings.ToLower(account.Email)
	account.Email = strings.TrimSpace(account.Email)
	return nil
}

// VALIDATION - requireEmail
func (av *accountValidator) requireEmail(account *Account) error {
	if account.Email == "" {
		return ErrEmailRequired
	}
	return nil
}

// VALIDATION - emailFormat
func (av *accountValidator) emailFormat(account *Account) error {
	if account.Email == "" {
		return nil
	}
	if !av.emailRegex.MatchString(account.Email) {
		return ErrEmailInvalid
	}
	return nil
}

// VALIDATION - emailTaken
func (av *accountValidator) emailIsAvail(account *Account) error {
	existing, err := av.ByEmail(account.Email)
	if err == ErrNotFound {
		// Email address is available if we don't find
		// an account with that email address
		return nil
	}
	// We can't continue our validation without a successful
	// query, so if we get any error other than ErrNotFound we
	// should return it.
	if err != nil {
		return err
	}
	// If we get here that means we found an account w/ this email
	// address, so we need to see if this is the same account we
	// are updating, or if we have a conflict.
	if account.ID != existing.ID {
		return ErrEmailTaken
	}
	return nil
}

// VALIDATUON - passwordMinLength
func (av *accountValidator) passwordMinLength(account *Account) error {
	if account.Password == "" {
		return nil
	}
	if len(account.Password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

// VALIDATION - passwordRequired
func (av *accountValidator) passwordRequired(account *Account) error {
	if account.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

// VALIDATION - passwordHashRequired
func (av *accountValidator) passwordHashRequried(account *Account) error {
	if account.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

// VALIDATION - rememberMinBytes
func (av *accountValidator) rememberMinBytes(account *Account) error {
	if account.Remember == "" {
		return nil
	}
	n, err := rand.NBytes(account.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return ErrRememberTooShort
	}
	return nil
}

// VALIDATION - rememberHashRequired
func (av *accountValidator) rememberHashRequired(account *Account) error {
	if account.RememberHash == "" {
		return ErrEmailRequired
	}
	return nil
}

// VALIDATION -
type accountValFn func(*Account) error

// VALIDATION -
func runAccountValFns(account *Account, fns ...accountValFn) error {
	for _, fn := range fns {
		if err := fn(account); err != nil {
			return err
		}
	}
	return nil
}

// VALIDATION - ByRemember will hash the remember token and then call ByRemember on the subsequent AccountDB layer.
func (av *accountValidator) ByRemember(token string) (*Account, error) {
	account := Account{
		Remember: token,
	}
	if err := runAccountValFns(&account, av.hmacRemember); err != nil {
		return nil, err
	}
	return av.AccountDB.ByRemember(account.RememberHash)
}

// VALIDATION - ByEmail will normalize an email address before passing it
// on to the database layer to perform the query.
func (av *accountValidator) ByEmail(email string) (*Account, error) {
	account := Account{
		Email: email,
	}
	err := runAccountValFns(&account, av.normalizeEmail)
	if err != nil {
		return nil, err
	}
	return av.AccountDB.ByEmail(account.Email)
}

// VALIDATION - Create will create the provided account and backfill data
// like the ID, CreatedAt, and UpdatedAt fields.
func (av *accountValidator) Create(account *Account) error {
	err := runAccountValFns(account,
		av.passwordRequired,
		av.passwordMinLength,
		av.bcryptPassword,
		av.passwordHashRequried,
		av.setRememberIfUnset,
		av.rememberMinBytes,
		av.hmacRemember,
		av.rememberHashRequired,
		av.normalizeEmail,
		av.requireEmail,
		av.emailFormat,
		av.emailIsAvail)
	if err != nil {
		return err
	}
	return av.AccountDB.Create(account)
}

// VALIDATION - Update will hash a remember token if provided.
func (av *accountValidator) Update(account *Account) error {
	err := runAccountValFns(account,
		av.passwordMinLength,
		av.bcryptPassword,
		av.passwordHashRequried,
		av.rememberMinBytes,
		av.hmacRemember,
		av.rememberHashRequired,
		av.normalizeEmail,
		av.requireEmail,
		av.emailFormat,
		av.emailIsAvail)
	if err != nil {
		return err
	}
	return av.AccountDB.Update(account)
}

// VALIDATION - Delete method will delete the account with the provided ID.
func (av *accountValidator) Delete(id uint) error {
	var account Account
	account.ID = id
	err := runAccountValFns(&account, av.idGreaterThan(0))
	if err != nil {
		return nil
	}
	return av.AccountDB.Delete(id)
}

// GORM - ByID will lookup an account with the provided id.
func (ag *accountGorm) ByID(id uint) (*Account, error) {
	var account Account
	db := ag.db.Where("id = ?", id)
	err := first(db, &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GORM - ByEmail will lookup an account with the provided email.
func (ag *accountGorm) ByEmail(email string) (*Account, error) {
	var account Account
	db := ag.db.Where("email = ?", email)
	err := first(db, &account)
	return &account, err
}

// GORM - ByRemember will lookup an account with the provided remember token.
// This method expects the remember token to already be hashed.
func (ag *accountGorm) ByRemember(rememberHash string) (*Account, error) {
	var account Account
	err := first(ag.db.Where("remember_hash = ?", rememberHash), &account)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// GORM - Create method produces an account and backfills the data.
func (ag *accountGorm) Create(account *Account) error {
	return ag.db.Create(account).Error
}

// GORM - Update method updates the account with data from the account object.
func (ag *accountGorm) Update(account *Account) error {
	return ag.db.Save(account).Error
}

// GORM - Delete method deletes the account with the provided ID
func (ag *accountGorm) Delete(id uint) error {
	account := Account{Model: gorm.Model{ID: id}}
	return ag.db.Delete(&account).Error
}

// First will query the gorm.DB and will return the first record and move it to dst.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
