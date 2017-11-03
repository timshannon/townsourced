// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package web

import (
	"net/http"

	"git.townsourced.com/townsourced/townsourced/app"
	"git.townsourced.com/townsourced/townsourced/data"
	"git.townsourced.com/townsourced/townsourced/fail"
)

func imageGet(w http.ResponseWriter, r *http.Request, c context) {
	imageKey := data.ToUUID(c.params.ByName("image"))

	// ?thumb
	// ?placeholder
	values := r.URL.Query()

	var i *app.Image
	var err error

	if _, ok := values["placeholder"]; ok {
		i, err = app.ImageGetPlaceholder(imageKey)
		if errHandled(err, w, r, c) {
			return
		}
	} else if _, ok := values["thumb"]; ok {
		i, err = app.ImageGetThumb(imageKey)
		if errHandled(err, w, r, c) {
			return
		}
	} else {
		i, err = app.ImageGet(imageKey)
		if errHandled(err, w, r, c) {
			return
		}
	}

	serveImage(w, r, i)
}

func serveImage(w http.ResponseWriter, r *http.Request, image *app.Image) {
	w.Header().Set("Content-Type", image.ContentType)
	w.Header().Set("ETag", image.Etag())

	http.ServeContent(w, r, string(image.Key), image.Updated, image.ReadSeeker())
}

func imagePost(w http.ResponseWriter, r *http.Request, c context) {
	// new image
	if c.session == nil {
		errHandled(&fail.Fail{
			Message:    "You must be logged in to upload an image",
			HTTPStatus: http.StatusUnauthorized,
		}, w, r, c)
		return
	}

	u, err := c.session.User()
	if errHandled(err, w, r, c) {
		return
	}

	images, err := imagesFromForm(u, r)
	if errHandled(err, w, r, c) {
		return
	}

	respondJsendCode(w, &JSend{
		Status: statusSuccess,
		Data:   images,
	}, http.StatusCreated)
}

func imagesFromForm(u *app.User, r *http.Request) ([]*app.Image, error) {

	var images []*app.Image

	err := r.ParseMultipartForm(maxUploadMemory)
	if err != nil {
		return nil, err
	}

	if len(r.MultipartForm.File) > app.PostMaxImages {
		return nil, app.ErrPostTooManyImages
	}

	for _, files := range r.MultipartForm.File {
		if len(files) > app.PostMaxImages {
			return nil, app.ErrPostTooManyImages
		}

		for i := range files {
			file, err := files[i].Open()
			if err != nil {
				return nil, err
			}

			i, err := app.ImageNew(u, files[i].Header.Get("Content-Type"), file)
			if err != nil {
				return nil, err
			}
			images = append(images, i)

			if len(images) >= app.PostMaxImages {
				return images, nil
			}
		}
	}

	return images, nil
}
