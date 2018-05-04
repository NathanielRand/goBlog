package controllers

import (
	"log"
	"net/http"
	"strconv"

	"muto/context"
	"muto/models"
	"muto/views"

	"github.com/gorilla/mux"
)

const (
	IndexMicroposts = "index_microposts"
	ShowMicropost   = "show_micropost"
	EditMicropost   = "edit_micropost"
)

func NewMicroposts(ms models.MicropostService, r *mux.Router) *Microposts {
	return &Microposts{
		New:       views.NewView("materialize", "microposts/new"),
		IndexView: views.NewView("materialize", "microposts/index"),
		ShowView:  views.NewView("materialize", "microposts/show"),
		EditView:  views.NewView("materialize", "microposts/edit"),
		ms:        ms,
		r:         r,
	}
}

type Microposts struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	ms        models.MicropostService
	r         *mux.Router
}

type MicropostForm struct {
	Content string `schema:"content"`
}

// POST /microposts
func (m *Microposts) Index(w http.ResponseWriter, r *http.Request) {
	account := context.Account(r.Context())
	microposts, err := m.ms.ByAccountID(account.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = microposts
	m.IndexView.Render(w, r, vd)
}

// GET /microposts/:id
func (m *Microposts) Show(w http.ResponseWriter, r *http.Request) {
	micropost, err := m.micropostByID(w, r)
	if err != nil {
		// The micropostByID method will already render the
		// error for us, so we just need to return here.
		return
	}
	var vd views.Data
	vd.Yield = micropost
	m.ShowView.Render(w, r, vd)
}

// GET /microposts/:id/edit
func (m *Microposts) Edit(w http.ResponseWriter, r *http.Request) {
	micropost, err := m.micropostByID(w, r)
	if err != nil {
		// The micropostByID method will already render the error for us,
		// so we just need to return here.
		return
	}
	// An account needs to be logged in to access this page,
	// so we can assume that the RequireAccount middleware has run
	// and set the account for us in the request context.
	account := context.Account(r.Context())
	if micropost.AccountID != account.ID {
		http.Error(w, "You do not have permission to edit this micropost",
			http.StatusForbidden)
		return
	}
	var vd views.Data
	vd.Yield = micropost
	m.EditView.Render(w, r, vd)
}

// POST /microposts/new
func (m *Microposts) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form MicropostForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		m.New.Render(w, r, vd)
		return
	}
	account := context.Account(r.Context())
	micropost := models.Micropost{
		Content:   form.Content,
		AccountID: account.ID,
	}
	if err := m.ms.Create(&micropost); err != nil {
		vd.SetAlert(err)
		m.New.Render(w, r, vd)
		return
	}
	// Generate a URL using the router and the names ShowMicropost route.
	url, err := m.r.Get(ShowMicropost).URL("id", strconv.Itoa(int(micropost.ID)))
	// Check for errors creating the URL.
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/microposts", http.StatusNotFound)
		return
	}
	// If no errors, use the URL and redirect
	// to the path portion of that URL.
	// We do not need the entire URL in case
	// application is hosted on custom domain.
	http.Redirect(w, r, url.Path, http.StatusFound)

}

// POST /microposts/update
func (m *Microposts) Update(w http.ResponseWriter, r *http.Request) {
	micropost, err := m.micropostByID(w, r)
	if err != nil {
		return
	}
	account := context.Account(r.Context())
	if micropost.AccountID != account.ID {
		http.Error(w, "Micropost not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	vd.Yield = micropost
	var form MicropostForm
	if err := parseForm(r, &form); err != nil {
		// If there is an error we are going
		// to render the EditView again
		// but with an Alert message.
		vd.SetAlert(err)
		m.EditView.Render(w, r, vd)
		return
	}
	micropost.Content = form.Content
	err = m.ms.Update(micropost)
	// If there is an error our alert will be an error. Otherwise
	// we will still render an alert, but instead it will be
	// a success message.
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Micropost updated successfully",
		}
	}
	// Error or not, we are going to render the EditView with
	// our updated information.
	m.EditView.Render(w, r, vd)
}

// POST /micropost/:id/delete
func (m *Microposts) Delete(w http.ResponseWriter, r *http.Request) {
	// Lookup the micropost using micropostByID
	micropost, err := m.micropostByID(w, r)
	if err != nil {
		return
	}
	// Retrieve the account and verify they have permission
	// to delete this micropost. Use RequireMiddleware
	// on any routes mapped to this method.
	account := context.Account(r.Context())
	if micropost.AccountID != account.ID {
		http.Error(w, "Permisson Denied! You do not have access to this micropost.", http.StatusForbidden)
		return
	}
	var vd views.Data
	err = m.ms.Delete(micropost.ID)
	if err != nil {
		// If an error occurs, set an alert and
		// render the edit page with the error.
		// Set the Yield to micropost so that
		// the EditView is rendered correctly.
		vd.SetAlert(err)
		vd.Yield = micropost
		m.EditView.Render(w, r, vd)
		return
	}
	url, err := m.r.Get(IndexMicroposts).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// micropostByID will parse the "id" variable from the
// request path using gorilla/mux and then use that ID to
// retrieve the micropost from the MicropostService
//
// micropostByID will return an error if one occurs, but it
// will also render the error with an http.Error function
// call, so you do not need to.
func (m *Microposts) micropostByID(w http.ResponseWriter,
	r *http.Request) (*models.Micropost, error) {
	// First we get the vars like we saw earlier. We do this
	// so we can get variables from our path, like the "id"
	// variable.
	vars := mux.Vars(r)
	// Next we need to get the "id" variable from our vars.
	idStr := vars["id"]
	// Our idStr is a string, so we use the Atoi function
	// provided by the strconv package to convert it into an
	// integer. This function can also return an error, so we
	// need to check for errors and render an error.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// If there is an error we will return the StatusNotFound
		// status code, as the page requested is for an invalid
		// micropost ID, which means the page doesn't exist.
		log.Println(err)
		http.Error(w, "Invalid micropost ID", http.StatusNotFound)
		return nil, err
	}
	micropost, err := m.ms.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Micropost not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Hmmm..Something went wrong.",
				http.StatusInternalServerError)
		}
		return nil, err
	}
	return micropost, nil
}
