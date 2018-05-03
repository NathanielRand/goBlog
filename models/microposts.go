package models

import (
	"github.com/jinzhu/gorm"
)

// Micropost - Error Constants
const (
	ErrAccountIDMicropostRequired modelError = "models: account ID is required"
	ErrContentRequired            modelError = "models: content is required"
)

var _ MicropostDB = &micropostGorm{}

type Micropost struct {
	gorm.Model
	AccountID uint   `gorm:"not_null;index"`
	Content   string `gorm:"not_null"`
}

type MicropostService interface {
	MicropostDB
}

type MicropostDB interface {
	ByID(id uint) (*Micropost, error)
	ByAccountID(accountID uint) ([]Micropost, error)
	Create(micropost *Micropost) error
	Update(micropost *Micropost) error
	Delete(id uint) error
}

// Micropost - Service
type micropostService struct {
	MicropostDB
}

// Micropost - Validation
type micropostValidator struct {
	MicropostDB
}

// Micropost - GORM
type micropostGorm struct {
	db *gorm.DB
}

type micropostValFn func(*Micropost) error

// Micropost - Validation
func runMicropostValFns(micropost *Micropost, fns ...micropostValFn) error {
	for _, fn := range fns {
		if err := fn(micropost); err != nil {
			return err
		}
	}
	return nil
}

// Micropost - Validation
func (mv *micropostValidator) accountIDRequired(m *Micropost) error {
	if m.AccountID <= 0 {
		return ErrAccountIDMicropostRequired
	}
	return nil
}

// Micropost - Validation
func (mv *micropostValidator) contentRequired(m *Micropost) error {
	if m.Content == "" {
		return ErrContentRequired
	}
	return nil
}

// Micropost - Validation - nonZeroID
func (mv *micropostValidator) nonZeroID(micropost *Micropost) error {
	if micropost.ID <= 0 {
		return ErrIDInvalid
	}
	return nil
}

// Micropost - Validation - Create
func (mv *micropostValidator) Create(micropost *Micropost) error {
	err := runMicropostValFns(micropost,
		mv.accountIDRequired,
		mv.contentRequired)
	if err != nil {
		return err
	}
	return mv.MicropostDB.Create(micropost)
}

// Micropost - Validation - Update
func (mv *micropostValidator) Update(micropost *Micropost) error {
	err := runMicropostValFns(micropost,
		mv.accountIDRequired,
		mv.contentRequired)
	if err != nil {
		return err
	}
	return mv.MicropostDB.Update(micropost)
}

// Micropost - Validation - Delete
func (mv *micropostValidator) Delete(id uint) error {
	var micropost Micropost
	micropost.ID = id
	if err := runMicropostValFns(&micropost, mv.nonZeroID); err != nil {
		return err
	}
	return mv.MicropostDB.Delete(micropost.ID)
}

// Micropost - GORM - ByID
func (mg *micropostGorm) ByID(id uint) (*Micropost, error) {
	var micropost Micropost
	db := mg.db.Where("id = ?", id)
	err := first(db, &micropost)
	if err != nil {
		return nil, err
	}
	return &micropost, nil
}

// Micropost - GORM - ByAccountID
func (mg *micropostGorm) ByAccountID(accountID uint) ([]Micropost, error) {
	var microposts []Micropost
	db := mg.db.Where("account_id = ?", accountID)
	if err := db.Find(&microposts).Error; err != nil {
		return nil, err
	}
	return microposts, nil
}

// Micropost - GORM - Create
func (mg *micropostGorm) Create(micropost *Micropost) error {
	return mg.db.Create(micropost).Error
}

// Micropost - GORM - Update
func (mg *micropostGorm) Update(micropost *Micropost) error {
	return mg.db.Save(micropost).Error
}

// Micropost - GORM - Delete
func (mg *micropostGorm) Delete(id uint) error {
	micropost := Micropost{Model: gorm.Model{ID: id}}
	return mg.db.Delete(&micropost).Error
}

// Micropost - GORM - Service
func NewMicropostService(db *gorm.DB) MicropostService {
	return &micropostService{
		MicropostDB: &micropostValidator{
			MicropostDB: &micropostGorm{
				db: db,
			},
		},
	}
}
