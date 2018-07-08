package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"coverd/context"
	"coverd/models"
	"coverd/views"

	"github.com/gorilla/mux"
)

const (
	IndexGalleries = "index_galleries"
	ShowGallery    = "show_gallery"
	EditGallery    = "edit_gallery"

	// Bit Shift
	maxMultipartMem = 1 << 20 // 1 megabyte
)

func NewGalleries(gs models.GalleryService, is models.ImageService, r *mux.Router) *Galleries {
	return &Galleries{
		New:       views.NewView("materialize", "galleries/new"),
		ShowView:  views.NewView("materialize", "galleries/show"),
		EditView:  views.NewView("materialize", "galleries/edit"),
		IndexView: views.NewView("materialize", "galleries/index"),
		gs:        gs,
		is:        is,
		r:         r,
	}
}

type Galleries struct {
	New       *views.View
	ShowView  *views.View
	EditView  *views.View
	IndexView *views.View
	gs        models.GalleryService
	is        models.ImageService
	r         *mux.Router
}

type GalleryForm struct {
	Title       string `schema:"title"`
	Reason      string `schema:"reason"`
	Doctor      string `schema:"doctor"`
	Cost        int    `schema:"cost"`
	Address     string `schema:"address"`
	City        string `schema:"city"`
	State       string `schema:"state"`
	ZipCode     int    `schema:"zipcode"`
	PhoneNumber string `schema:"phonenumber"`
}

// POST /galleries
func (g *Galleries) Index(w http.ResponseWriter, r *http.Request) {
	account := context.Account(r.Context())
	galleries, err := g.gs.ByAccountID(account.ID)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var vd views.Data
	vd.Yield = galleries
	g.IndexView.Render(w, r, vd)
}

// GET /galleries/:id
func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		// The galleryByID method will already render the
		// error for us, so we just need to return here.
		return
	}
	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, r, vd)
}

// GET /galleries/:id/edit
func (g *Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		// The galleryByID method will already render the error for us,
		// so we just need to return here.
		return
	}
	// An account needs to be logged in to access this page,
	// so we can assume that the RequireAccount middleware has run
	// and set the account for us in the request context.
	account := context.Account(r.Context())
	if gallery.AccountID != account.ID {
		http.Error(w, "You do not have permission to edit this gallery",
			http.StatusForbidden)
		return
	}
	var vd views.Data
	vd.Yield = gallery
	g.EditView.Render(w, r, vd)
}

// POST /galleries/:id/images
func (g *Galleries) ImageUpload(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	account := context.Account(r.Context())
	if gallery.AccountID != account.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	var vd views.Data
	vd.Yield = gallery
	err = r.ParseMultipartForm(maxMultipartMem)
	if err != nil {
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}

	// Iterate over uploaded files to process them.
	files := r.MultipartForm.File["images"]
	for _, f := range files {
		// Open the uploaded file with existing code
		file, err := f.Open()
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
		defer file.Close()

		// Call the ImageService's Create method.
		// Create the image
		err = g.is.Create(gallery.ID, file, f.Filename)
		if err != nil {
			vd.SetAlert(err)
			g.EditView.Render(w, r, vd)
			return
		}
	}

	url, err := g.r.Get(EditGallery).URL("id",
		fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// POST /galleries/:id/images/:filename/delete
func (g *Galleries) ImageDelete(w http.ResponseWriter, r *http.Request) {
	// Look up gallery by it's ID.
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	// Verify current account has permission to edit gallery.
	account := context.Account(r.Context())
	if gallery.AccountID != account.ID {
		http.Error(w, "You do not have permission to edit "+
			"this gallery or image", http.StatusForbidden)
		return
	}
	// Get the filename from the path.
	filename := mux.Vars(r)["filename"]
	// Build the Image model.
	i := models.Image{
		Filename:  filename,
		GalleryID: gallery.ID,
	}
	// Try to delete the image.
	err = g.is.Delete(&i)
	if err != nil {
		// Render the edit page with any errors.
		var vd views.Data
		vd.Yield = gallery
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	// If all goes well, redirect to the edit gallery page.
	url, err := g.r.Get(EditGallery).URL("id", fmt.Sprintf("%v", gallery.ID))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// POST /galleries/new
func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	account := context.Account(r.Context())
	gallery := models.Gallery{
		Reason:      form.Reason,
		Doctor:      form.Doctor,
		Cost:        form.Cost,
		Address:     form.Address,
		City:        form.City,
		State:       form.State,
		ZipCode:     form.ZipCode,
		PhoneNumber: form.PhoneNumber,
		AccountID:   account.ID,
	}
	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, r, vd)
		return
	}
	// Generate a URL using the router and the names ShowGallery route.
	url, err := g.r.Get(EditGallery).URL("id", strconv.Itoa(int(gallery.ID)))
	// Check for errors creating the URL.
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/galleries", http.StatusNotFound)
		return
	}
	// If no errors, use the URL and redirect
	// to the path portion of that URL.
	// We do not need the entire URL in case
	// application is hosted on custom domain (i.e. "www.domain.com").
	http.Redirect(w, r, url.Path, http.StatusFound)

}

// POST /galleries/update
func (g *Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	account := context.Account(r.Context())
	if gallery.AccountID != account.ID {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}
	var vd views.Data
	vd.Yield = gallery
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		// If there is an error we are going
		// to render the EditView again
		// but with an Alert message.
		vd.SetAlert(err)
		g.EditView.Render(w, r, vd)
		return
	}
	gallery.Reason = form.Reason
	gallery.Doctor = form.Doctor
	gallery.Cost = form.Cost
	gallery.Address = form.Address
	gallery.City = form.City
	gallery.State = form.State
	gallery.ZipCode = form.ZipCode
	gallery.PhoneNumber = form.PhoneNumber
	err = g.gs.Update(gallery)
	// If there is an error our alert will be an error. Otherwise
	// we will still render an alert, but instead it will be
	// a success message.
	if err != nil {
		vd.SetAlert(err)
	} else {
		vd.Alert = &views.Alert{
			Level:   views.AlertLvlSuccess,
			Message: "Gallery updated successfully",
		}
	}
	// Error or not, we are going to render the EditView with
	// our updated information.
	g.EditView.Render(w, r, vd)
}

// POST /gallery/:id/delete
func (g *Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	// Lookup the gallery using galleryByID
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}
	// Retrieve the account and verify they have permission
	// to delete this gallery. Use RequireMiddleware
	// on any routes mapped to this method.
	account := context.Account(r.Context())
	if gallery.AccountID != account.ID {
		http.Error(w, "Permisson Denied! You do not have access to this gallery.", http.StatusForbidden)
		return
	}
	var vd views.Data
	err = g.gs.Delete(gallery.ID)
	if err != nil {
		// If an error occurs, set an alert and
		// render the edit page with the error.
		// Set the Yield to gallery so that
		// the EditView is rendered correctly.
		vd.SetAlert(err)
		vd.Yield = gallery
		g.EditView.Render(w, r, vd)
		return
	}
	url, err := g.r.Get(IndexGalleries).URL()
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, url.Path, http.StatusFound)
}

// galleryByID will parse the "id" variable from the
// request path using gorilla/mux and then use that ID to
// retrieve the gallery from the GalleryService
//
// galleryByID will return an error if one occurs, but it
// will also render the error with an http.Error function
// call, so you do not need to.
func (g *Galleries) galleryByID(w http.ResponseWriter,
	r *http.Request) (*models.Gallery, error) {
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
		// gallery ID, which means the page doesn't exist.
		log.Println(err)
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.gs.ByID(uint(id))
	if err != nil {
		switch err {
		case models.ErrNotFound:
			http.Error(w, "Gallery not found", http.StatusNotFound)
		default:
			log.Println(err)
			http.Error(w, "Hmmm..Something went wrong.",
				http.StatusInternalServerError)
		}
		return nil, err
	}
	images, _ := g.is.ByGalleryID(gallery.ID)
	gallery.Images = images
	return gallery, nil
}

// Lookup galleryByCatergory
// Lookup galleryByTag
// Lookup galleryByDateCreated
