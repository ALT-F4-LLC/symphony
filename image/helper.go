package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// Image : struct for image in postgres
type Image struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Error : struct for handling errors
type Error struct {
	Error string
}

// HandleError : handles response
func HandleError(error string) []byte {
	err := &Error{
		Error: error,
	}
	json, e := json.Marshal(err)
	if e != nil {
		panic(e)
	}
	return json
}

// GetImageByIDHandler : handles retrival of image from API request
func GetImageByIDHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		images, err := GetImageByID(db, id)
		if len(images) < 1 {
			err := HandleError("invalid_image")
			w.WriteHeader(http.StatusOK)
			w.Write(err)
			return
		}

		image := &Image{
			ID:   images[0].ID,
			Name: images[0].Name,
		}

		json, err := json.Marshal(image)
		if err != nil {
			panic(err)
		}

		v := r.URL.Query()
		download := v.Get("download")
		if download == "true" {
			path := fmt.Sprintf("data/%s", image.Name)
			f, err := os.Stat(path)
			if err != nil {
				w.WriteHeader(404)
				w.Write([]byte("404 - Something went wrong :/" + http.StatusText(404)))
				return
			}
			data, err := ioutil.ReadFile(string(path))
			if err != nil {
				w.WriteHeader(404)
				w.Write([]byte("404 - Something went wrong :/" + http.StatusText(404)))
				return
			}
			disposition := fmt.Sprintf("attachment; filename=%s", image.Name)
			size := fmt.Sprintf("%d", f.Size())
			w.Header().Set("Content-Disposition", disposition)
			w.Header().Add("Content-Length", size)
			w.Header().Add("Content-Type", "application/octet-stream")
			w.Write(data)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

// GetImageByID : gets specific virtual machine image from database
func GetImageByID(db *sql.DB, id string) ([]Image, error) {
	var images []Image
	rows, queryErr := db.Query("SELECT * FROM image WHERE id = $1", id)
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()
	for rows.Next() {
		var image Image
		if e := rows.Scan(&image.ID, &image.Name); e != nil {
			return nil, e
		}
		images = append(images, image)
	}
	return images, nil
}

// GetImages : gets virtual machine images from database
func GetImages(db *sql.DB) ([]Image, error) {
	var images []Image
	rows, queryErr := db.Query("SELECT * FROM image")
	if queryErr != nil {
		return nil, queryErr
	}
	defer rows.Close()
	for rows.Next() {
		var image Image
		if e := rows.Scan(&image.ID, &image.Name); e != nil {
			return nil, e
		}
		images = append(images, image)
	}
	return images, nil
}
