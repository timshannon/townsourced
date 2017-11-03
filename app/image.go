// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"git.townsourced.com/townsourced/goexif/exif"
	"git.townsourced.com/townsourced/imaging"
	log "git.townsourced.com/townsourced/logrus"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

const (
	imageMaxSize              = int64(10 << 20) //10MB
	imageMaxWidth             = 2048
	imageMaxHeight            = 2048
	imageMaxThumbHeight       = 300
	imageMaxThumbWidth        = 300
	imageMaxPlaceholderWidth  = 30
	imageMaxPlaceholderHeight = 30
)

var (
	// ErrImageNotFound is the error returned when an image is not found
	ErrImageNotFound = &fail.Fail{
		Message:    "Image not found",
		HTTPStatus: http.StatusNotFound,
	}
	// ErrImageTooLarge is returned with the passed in image is too large
	ErrImageTooLarge = &fail.Fail{
		Message:    "The uploaded image is too large.  The max size is 5 MB",
		HTTPStatus: http.StatusRequestEntityTooLarge,
	}

	// ErrImageInvalidType is when the content type of the image is not supported
	ErrImageInvalidType = fail.New("Unsupported image type, please use a jpeg, png, or gif")

	// ErrImageNotOwner is when someone who is not the owner of the image is trying to modify it
	ErrImageNotOwner = fail.New("You cannot update an image you are not the owner of")

	// ErrImageDecodeError is when the image data can't be decoded properly
	ErrImageDecodeError = fail.New("The uploaded image cannot be decoded properly based on the content type")
)

var imageValidContentTypes = []string{"image/gif", "image/jpeg", "image/png"}

// Image is a user uploaded image for use in posts, or other items
type Image struct {
	Key data.UUID `json:"key,omitempty" gorethink:",omitempty"`

	OwnerKey    data.Key `json:"ownerKey,omitempty" gorethink:",omitempty"`
	ContentType string   `json:"contentType,omitempty" gorethink:",omitempty"`

	// whether or not the image is used by a post or elsewhere
	// images not in use will be cleaned up by a task
	InUse bool `json:"-"`
	data.Version

	Data            []byte `json:"-" gorethink:",omitempty"` // full image
	ThumbData       []byte `json:"-" gorethink:",omitempty"` // thumbnail image
	PlaceholderData []byte `json:"-" gorethink:",omitempty"` // Placeholder image small, and downloads quick

	image image.Image
}

// ImageNew inserts a new image into the database
// closes the reader when finished
func ImageNew(owner *User, contentType string, reader io.ReadCloser) (img *Image, err error) {
	defer func() {
		if cerr := reader.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	lr := &io.LimitedReader{R: reader, N: (imageMaxSize + 1)}
	buff, err := ioutil.ReadAll(lr)

	if err != nil {
		return nil, err
	}
	if lr.N == 0 {
		return nil, ErrImageTooLarge
	}

	img, err = imageNew(owner, contentType, buff, false)
	return
}

func imageNew(owner *User, contentType string, imgData []byte, inUse bool) (*Image, error) {
	i := &Image{
		OwnerKey:    owner.Username,
		ContentType: contentType,
		Data:        imgData,
		InUse:       inUse,
	}

	err := i.validate()
	if err != nil {
		return nil, err
	}

	err = i.prep()
	if err != nil {
		return nil, err
	}

	i.Rev()
	key, err := data.ImageInsert(i)
	if err != nil {
		return nil, err
	}
	i.Key = key
	return i, nil
}

// ImageGet retrieves an image by it's key
func ImageGet(key data.UUID) (*Image, error) {
	return imageGet(key, false, false)
}

// ImageGetThumb retrieves an image by it's key
func ImageGetThumb(key data.UUID) (*Image, error) {
	return imageGet(key, true, false)
}

// ImageGetPlaceholder retrieves an image by it's key
func ImageGetPlaceholder(key data.UUID) (*Image, error) {
	return imageGet(key, false, true)
}

func imageGet(key data.UUID, thumb, placeholder bool) (*Image, error) {
	i := &Image{}
	err := data.ImageGet(i, key, thumb, placeholder)
	if err == data.ErrNotFound {
		return nil, ErrImageNotFound
	}
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Etag returns an appropriate string to use as an HTTP etag, or database version
func (i *Image) Etag() string {
	return i.Version.Ver()
}

func (i *Image) update() error {
	if i.image != nil {
		err := i.encode()
		if err != nil {
			return err
		}
	}
	return data.ImageUpdate(i, i.Key)
}

func (i *Image) delete() error {
	return data.ImageDelete(i.Key)
}

//imageSetUsed updates an image as inUse and updates it in the
// database
func imageSetUsed(imageKey data.UUID) error {
	return data.ImageUpdate(struct {
		InUse bool
	}{
		InUse: true,
	}, imageKey)

}

func (i *Image) validate() error {
	if len(i.Data) == 0 {
		return errors.New("Image not loaded properly")
	}
	//check content type
	found := false
	for j := range imageValidContentTypes {
		if i.ContentType == imageValidContentTypes[j] {
			found = true
			break
		}
	}
	if !found {
		return ErrImageInvalidType
	}

	return nil
}

// fit the image into the proper size and build thumbnail and placeholders
func (i *Image) prep() error {
	orientation := 1
	if i.ContentType == "image/jpeg" {
		// rotate based on exif data if applicable
		ex, err := exif.Decode(bytes.NewReader(i.Data))
		// if at any point we can't parse exif data properly, then we stick to default orientation
		if err == nil {
			//lookup orientation
			tag, err := ex.Get(exif.Orientation)
			if err == nil {
				orientation, _ = tag.Int(0)
			}
		}

	}
	//resize to max allowable image size
	err := i.decode()
	if err != nil {
		return err
	}

	//0 will preserve the ratio
	newWidth := 0
	newHeight := 0

	if i.image.Bounds().Dx() > imageMaxWidth {
		newWidth = imageMaxWidth
	} else if i.image.Bounds().Dy() > imageMaxHeight {
		newHeight = imageMaxHeight
	}

	if newWidth != 0 || newHeight != 0 {
		err = i.resize(&User{Username: i.OwnerKey}, newWidth, newHeight)
		if err != nil {
			return err
		}
	}

	switch orientation {
	case 2:
		i.image = imaging.FlipH(i.image)
	case 3:
		i.image = imaging.Rotate180(i.image)
	case 4:
		i.image = imaging.FlipV(i.image)
	case 5:
		i.image = imaging.Rotate90(imaging.FlipH(i.image))
	case 6:
		i.image = imaging.Rotate270(i.image)
	case 7:
		i.image = imaging.FlipH(imaging.Rotate90(i.image))
	case 8:
		i.image = imaging.Rotate90(i.image)
	}

	err = i.encode()
	if err != nil {
		return err
	}
	return nil
}

// decode decodes the image data into the go Image format for processing
func (i *Image) decode() error {
	if i.image != nil {
		//image is already decoded
		return nil
	}

	if len(i.Data) == 0 {
		return errors.New("Image not loaded properly.")
	}

	var err error
	buffer := bytes.NewBuffer(i.Data)

	switch i.ContentType {
	case "image/gif":
		i.image, err = gif.Decode(buffer)
	case "image/jpeg":
		i.image, err = jpeg.Decode(buffer)
	case "image/png":
		i.image, err = png.Decode(buffer)
	default:
		return ErrImageInvalidType
	}

	if err != nil {
		log.WithField("imagekey", i.Key).Error(err)
		return err
	}

	i.Data = nil

	return nil
}

// encode encodes the image format data into image data
func (i *Image) encode() error {
	if i.image == nil {
		// image is already encoded
		return nil
	}

	buffer := bytes.NewBuffer(i.Data)

	err := imageEncode(i.image, i.ContentType, buffer)

	if err != nil {
		return err
	}

	i.Data = buffer.Bytes()
	err = i.buildThumbAndPlaceholder()
	if err != nil {
		return err
	}
	i.image = nil

	return nil
}

func imageEncode(image image.Image, contentType string, result *bytes.Buffer) error {
	switch contentType {
	case "image/gif":
		return gif.Encode(result, image, nil)
	case "image/jpeg":
		return jpeg.Encode(result, image, nil)
	case "image/png":
		return png.Encode(result, image)
	default:
		return ErrImageInvalidType
	}
}

func (i *Image) buildThumbAndPlaceholder() error {
	//0 will preserve the ratio
	newWidth := 0
	newHeight := 0
	var err error
	// create thumb and placeholder seqeuentially to benefit from
	// increasingly smaller images

	//first thumb

	//thumbnail rules are a bit different, resize so that the smallest
	// dimension is resized instead of the largest, this is to prevent
	// cases where really tall or really wide images create poor
	// thumbnails
	if i.image.Bounds().Dx() > i.image.Bounds().Dy() {
		//wider than tall
		newHeight = imageMaxThumbHeight
	} else {
		//taller than wide or same
		newWidth = imageMaxThumbWidth
	}

	if newWidth != 0 || newHeight != 0 {
		err = i.resize(&User{Username: i.OwnerKey}, newWidth, newHeight)
		if err != nil {
			return err
		}
	}

	buffer := bytes.NewBuffer(i.ThumbData)
	err = imageEncode(i.image, i.ContentType, buffer)
	if err != nil {
		return err
	}

	i.ThumbData = buffer.Bytes()

	// then placeholder
	newWidth = 0
	newHeight = 0

	if i.image.Bounds().Dx() > imageMaxPlaceholderWidth {
		newWidth = imageMaxPlaceholderWidth
	} else if i.image.Bounds().Dy() > imageMaxPlaceholderHeight {
		newHeight = imageMaxPlaceholderHeight
	}

	if newWidth != 0 || newHeight != 0 {
		err = i.resize(&User{Username: i.OwnerKey}, newWidth, newHeight)
		if err != nil {
			return err
		}
	}

	buffer = bytes.NewBuffer(i.PlaceholderData)
	err = imageEncode(i.image, i.ContentType, buffer)
	if err != nil {
		return err
	}

	i.PlaceholderData = buffer.Bytes()

	return nil
}

func (i *Image) resize(who *User, width, height int) error {
	if i.OwnerKey != who.Username {
		return ErrImageNotOwner
	}

	err := i.decode()
	if err != nil {
		return err
	}

	i.image = imaging.Resize(i.image, width, height, imaging.Linear)
	return nil
}

func (i *Image) crop(who *User, x0, y0, x1, y1 int) error {
	if i.OwnerKey != who.Username {
		return ErrImageNotOwner
	}
	err := i.decode()
	if err != nil {
		return err
	}

	i.image = imaging.Crop(i.image, image.Rect(x0, y0, x1, y1))
	return nil
}

func (i *Image) cropCenter(who *User, width, height int) error {
	if i.OwnerKey != who.Username {
		return ErrImageNotOwner
	}
	err := i.decode()
	if err != nil {
		return err
	}

	i.image = imaging.CropCenter(i.image, width, height)
	return nil
}

// ReadSeeker returns a ReadSeeker for the available image data
// uses the highest quality image data available
func (i *Image) ReadSeeker() io.ReadSeeker {
	if len(i.Data) != 0 {
		return bytes.NewReader(i.Data)
	}

	if len(i.ThumbData) != 0 {
		return bytes.NewReader(i.ThumbData)
	}
	if len(i.PlaceholderData) != 0 {
		return bytes.NewReader(i.PlaceholderData)
	}

	// No data available?  Shouldn't happen
	return bytes.NewReader([]byte{})
}

//Delete Unused Image Tasker

type taskerUnusedImages struct{}

func (d *taskerUnusedImages) Type() string       { return "DeleteUnusedImages" }
func (d *taskerUnusedImages) Priority() uint     { return priorityLow }
func (d *taskerUnusedImages) NextRun() time.Time { return time.Now().Add(15 * time.Minute) }
func (d *taskerUnusedImages) Retry() int         { return -1 }
func (d *taskerUnusedImages) Do(variables ...interface{}) error {
	// delete all images that aren't in use an hour after they were last updated
	err := data.ImageDeleteOrphans(time.Now().Add(-1 * time.Hour))
	if err == data.ErrNotFound {
		return nil
	}
	return err
}
