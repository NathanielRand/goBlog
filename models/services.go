package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Services struct {
	Account   AccountService
	Gallery   GalleryService
	Image     ImageService
	Micropost MicropostService
	db        *gorm.DB
}

type ServicesConfig func(*Services) error

// NewServices now will accept a list of config functions to
// run. Each function will accept a pointer to the current
// Services object as its only argument and will edit that
// object inline and return an error if there is one. Once
// we have run all configs we will return the Services object.
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	// For each ServicesConfig function...
	for _, cfg := range cfgs {
		// Run the function passing in a pointer to our Services
		// object and catching any errors
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}
	// Then finally return the result
	return &s, nil
}

func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}
func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)
		return nil
	}
}

func WithAccount(pepper, hmacKey string) ServicesConfig {
	return func(s *Services) error {
		s.Account = NewAccountService(s.db, pepper, hmacKey)
		return nil
	}
}

func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}

func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService()
		return nil
	}
}

func WithMicropost() ServicesConfig {
	return func(s *Services) error {
		s.Micropost = NewMicropostService(s.db)
		return nil
	}
}

func (s *Services) Close() error {
	return s.db.Close()
}

func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&Account{}, &Gallery{}, &Micropost{}).Error
}

func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&Account{}, &Gallery{}, &Micropost{}).Error
	if err != nil {
		return err
	}
	return s.AutoMigrate()
}
