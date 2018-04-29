package models

import (
	"github.com/jinzhu/gorm"
)

// GALLERY - ERRORS
const (
	ErrAccountIDRequired modelError = "models: account ID is required"
	ErrTitleRequired     modelError = "models: title is required"
)

var _ GalleryDB = &galleryGorm{}

type Gallery struct {
	gorm.Model
	AccountID uint    `gorm:"not_null;index"`
	Title     string  `gorm:"not_null"`
	Images    []Image `gorm:"-"`
}

type GalleryService interface {
	GalleryDB
}

type GalleryDB interface {
	ByID(id uint) (*Gallery, error)
	ByAccountID(accountID uint) ([]Gallery, error)
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error
}

// GALLERY - SERVICE
type galleryService struct {
	GalleryDB
}

// GALLERY - VALIDATION
type galleryValidator struct {
	GalleryDB
}

// GALLERY - GORM
type galleryGorm struct {
	db *gorm.DB
}

type galleryValFn func(*Gallery) error

// GALLERY - VALIDATION
func runGalleryValFns(gallery *Gallery, fns ...galleryValFn) error {
	for _, fn := range fns {
		if err := fn(gallery); err != nil {
			return err
		}
	}
	return nil
}

// GALLERY - VALIDATION
func (mv *galleryValidator) accountIDRequired(m *Gallery) error {
	if m.AccountID <= 0 {
		return ErrAccountIDRequired
	}
	return nil
}

// GALLERY - VALIDATION
func (mv *galleryValidator) titleRequired(m *Gallery) error {
	if m.Title == "" {
		return ErrTitleRequired
	}
	return nil
}

// GALLERY - VALIDATION
// // categoryRequired
// // imageRequired

// GALLERY - VALIDATION - Create
func (mv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFns(gallery,
		mv.accountIDRequired,
		mv.titleRequired)
	if err != nil {
		return err
	}
	return mv.GalleryDB.Create(gallery)
}

// GALLERY - VALIDATION - Update
func (mv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFns(gallery,
		mv.accountIDRequired,
		mv.titleRequired)
	if err != nil {
		return err
	}
	return mv.GalleryDB.Update(gallery)
}

// GALLERY - VALIDATION - nonZeroID
func (mv *galleryValidator) nonZeroID(gallery *Gallery) error {
	if gallery.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

// GALLERY - VALIDATION - Delete
func (mv *galleryValidator) Delete(id uint) error {
	var gallery Gallery
	gallery.ID = id
	if err := runGalleryValFns(&gallery, mv.nonZeroID); err != nil {
		return err
	}
	return mv.GalleryDB.Delete(gallery.ID)
}

// // GALLERY - VALIDATION - categoryTattoo
// func (mv *galleryValidator) categoryTattoo(gallery *Gallery) error {
// 	if gallery.Category == "tattoo" {
// 		return ErrIDInvalid
// 	}
// 	return nil
// }

// GALLERY - GORM
func (mg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := mg.db.Where("id = ?", id)
	err := first(db, &gallery)
	if err != nil {
		return nil, err
	}
	return &gallery, nil
}

// GALLERY - GORM
func (mg *galleryGorm) ByAccountID(accountID uint) ([]Gallery, error) {
	var galleries []Gallery
	db := mg.db.Where("account_id = ?", accountID)
	if err := db.Find(&galleries).Error; err != nil {
		return nil, err
	}
	return galleries, nil
}

// GALLERY - GORM
// // ByCategory
// // ByTag
// // ByDateCreated

// GALLERY - GORM
func (mg *galleryGorm) Create(gallery *Gallery) error {
	return mg.db.Create(gallery).Error
}

// GALLERY - GORM
func (mg *galleryGorm) Update(gallery *Gallery) error {
	return mg.db.Save(gallery).Error
}

// GALLERY - GORM
func (mg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return mg.db.Delete(&gallery).Error
}

// GALLERY - SERVICE
func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			GalleryDB: &galleryGorm{
				db: db,
			},
		},
	}
}

// Render Image columns easier
func (m *Gallery) ImagesSplitN(n int) [][]Image {
	// Create out 2D slice
	ret := make([][]Image, n)
	// Create the inner slices - we need N of them,
	// and we will start them with a size of 0.
	for i := 0; i < n; i++ {
		ret[i] = make([]Image, 0)
	}
	// Iterate over our images, using the index % n
	// to determine which of the slices in ret to add the image to.
	for i, img := range m.Images {
		// % is the remainder operator in Go
		// eg:
		// 0%3 = 0
		// 1%3 = 1
		// 2%3 = 2
		// 3%3 = 0
		// 4%3 = 1
		// 5%3 = 2
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}
	return ret
}
