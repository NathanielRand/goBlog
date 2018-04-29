package models

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

type ImageService interface {
	Create(galleryID uint, r io.Reader, filename string) error
	ByGalleryID(galleryID uint) ([]Image, error)
	Delete(i *Image) error
}

// Image is used to represent images stored in a Gallery
// Image is NOT stored in the database, and
// instead references data stored on disk.
type Image struct {
	GalleryID uint
	Filename  string
}

type imageService struct{}

func NewImageService() ImageService {
	return &imageService{}
}

func (is *imageService) Create(galleryID uint,
	r io.Reader, filename string) error {
	path, err := is.mkImageDir(galleryID)
	if err != nil {
		return err
	}
	// Create a destination file
	dst, err := os.Create(filepath.Join(path, filename))
	if err != nil {
		return err
	}
	defer dst.Close()
	// Copy reader data to the destination file
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (is *imageService) Delete(i *Image) error {
	return os.Remove(i.RelativePath())
}

func (is *imageService) ByGalleryID(galleryID uint) ([]Image, error) {
	path := is.imageDir(galleryID)
	strings, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, err
	}
	// Setup the Image slice we are returning
	ret := make([]Image, len(strings))
	for i, imgStr := range strings {
		ret[i] = Image{
			Filename:  filepath.Base(imgStr),
			GalleryID: galleryID,
		}
	}
	return ret, nil
}

// Use this function when we know it is already made.
func (is *imageService) imageDir(galleryID uint) string {
	// Create the directory to contain our images
	// filepath.Join will return a path like:
	// images/galleries/123
	// We use filepath.Join instead of building the path
	// manually because the slashes and other characters
	// could vary between operating systems.
	return filepath.Join("images", "galleries",
		fmt.Sprintf("%v", galleryID))
}

func (is *imageService) mkImageDir(galleryID uint) (string, error) {
	galleryPath := is.imageDir(galleryID)
	// Create our directory (and any necessary parent dirs)
	// using 0755 permissions.
	err := os.MkdirAll(galleryPath, 0755)
	if err != nil {
		return "", err
	}
	return galleryPath, nil
}

// Path is used to build the absolute path
// used to reference this image
// via a web request.
func (i *Image) Path() string {
	temp := url.URL{
		Path: "/" + i.RelativePath(),
	}
	return temp.String()
}

// RelativePath is used to build the path to this image
// on our local disk, relative to where our
// Go application is run from.
func (i *Image) RelativePath() string {
	// Convert the gallery ID to a string
	galleryID := fmt.Sprintf("%v", i.GalleryID)
	return filepath.ToSlash(filepath.Join("images", "galleries", galleryID, i.Filename))
}
