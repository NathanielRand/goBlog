package main

import (
	"flag"
	"fmt"
	"net/http"

	"muto/controllers"
	"muto/middleware"
	"muto/models"
	"muto/rand"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func main() {
	// Production flag for ensuring .config file is provided.
	boolPtr := flag.Bool("prod", false, "Provide this flag "+
		"in production. This ensures that a .config file is "+
		"provided before the application starts.")
	flag.Parse()
	// Configuartion Information.
	// boolPtr is a pointer to a boolean, so we need to use
	// *boolPtr to get the boolean value and pass it
	// into our LoadConfig function.
	cfg := LoadConfig(*boolPtr)
	dbCfg := cfg.Database
	// Create a service, check for errors,
	// defer close until main func exits,
	// and call the AutoMigrate function.
	services, err := models.NewServices(
		models.WithGorm(dbCfg.Dialect(), dbCfg.ConnectionInfo()),
		// Only log when not in prod
		models.WithLogMode(!cfg.IsProd()),
		models.WithAccount(cfg.Pepper, cfg.HMACKey),
		models.WithGallery(),
		models.WithImage(),
		models.WithMicropost(),
	)

	if err != nil {
		panic(err)
	}

	defer services.Close()
	services.AutoMigrate()

	// Controllers
	r := mux.NewRouter()
	staticC := controllers.NewStatic()
	accountsC := controllers.NewAccounts(services.Account)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)
	micropostsC := controllers.NewMicroposts(services.Micropost, r)

	// Middleware - Check Account Logged In
	AccountMw := middleware.Account{
		AccountService: services.Account,
	}

	// Middleware - Require Account Logged In
	requireAccountMw := middleware.RequireAccount{}

	// Asset Routes
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	// Image Routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Static Routes
	r.Handle("/", staticC.LandingView).Methods("GET")
	r.Handle("/contact", staticC.ContactView).Methods("GET")
	r.Handle("/faq", staticC.FaqView).Methods("GET")
	r.Handle("/faq-question", staticC.FaqQuestionView).Methods("GET")
	r.Handle("/dashboard", staticC.DashboardView).Methods("GET")

	// Account Routes
	r.HandleFunc("/register", accountsC.New).Methods("GET")
	r.HandleFunc("/register", accountsC.Create).Methods("POST")
	r.Handle("/enter", accountsC.LoginView).Methods("GET")
	r.HandleFunc("/enter", accountsC.Login).Methods("POST")
	r.HandleFunc("/cookietest", accountsC.CookieTest).Methods("GET")

	// Gallery Routes
	r.Handle("/galleries/new",
		requireAccountMw.Apply(galleriesC.New)).
		Methods("GET")
	r.Handle("/galleries",
		requireAccountMw.ApplyFn(galleriesC.Create)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}",
		galleriesC.Show).
		Methods("GET").
		Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit",
		requireAccountMw.ApplyFn(galleriesC.Edit)).
		Methods("GET").
		Name(controllers.EditGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/update",
		requireAccountMw.ApplyFn(galleriesC.Update)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete",
		requireAccountMw.ApplyFn(galleriesC.Delete)).
		Methods("POST")
	r.HandleFunc("/galleries",
		requireAccountMw.ApplyFn(galleriesC.Index)).
		Methods("GET").
		Name(controllers.IndexGalleries)
	r.HandleFunc("/galleries/{id:[0-9]+}/images",
		requireAccountMw.ApplyFn(galleriesC.ImageUpload)).
		Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete",
		requireAccountMw.ApplyFn(galleriesC.ImageDelete)).
		Methods("POST")

	// Micropost Routes
	r.Handle("/microposts/new",
		requireAccountMw.Apply(micropostsC.New)).
		Methods("GET")
	r.Handle("/microposts",
		requireAccountMw.ApplyFn(micropostsC.Create)).
		Methods("POST")
	r.HandleFunc("/microposts/{id:[0-9]+}",
		micropostsC.Show).
		Methods("GET").
		Name(controllers.ShowMicropost)
	r.HandleFunc("/microposts/{id:[0-9]+}/edit",
		requireAccountMw.ApplyFn(micropostsC.Edit)).
		Methods("GET").
		Name(controllers.EditMicropost)
	r.HandleFunc("/microposts/{id:[0-9]+}/update",
		requireAccountMw.ApplyFn(micropostsC.Update)).
		Methods("POST")
	r.HandleFunc("/microposts/{id:[0-9]+}/delete",
		requireAccountMw.ApplyFn(micropostsC.Delete)).
		Methods("POST")
	r.HandleFunc("/microposts",
		requireAccountMw.ApplyFn(micropostsC.Index)).
		Methods("GET")

	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}

	// CSRF Protection
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))

	// Server Messages / Listen & Serve
	fmt.Printf("Starting the server on Port:%d...\n", cfg.Port)
	fmt.Println("Success! Application Compiled.")
	fmt.Println("Application Running...")
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(AccountMw.Apply(r)))
}
