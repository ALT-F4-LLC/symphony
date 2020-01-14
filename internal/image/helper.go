package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Error : struct for handling errors
type Error struct {
	Error string
}

// Image : struct for image in postgres
type Image struct {
	Filename string `json:"filename"`
	ID       string `json:"id"`
	Name     string `json:"name"`
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

// PostImageHandler : handles retrival of image from API request
func PostImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(5000 << 20)

		name := r.FormValue("name")
		if name == "" {
			w.WriteHeader(500)
			w.Write([]byte("Invalid name :/" + http.StatusText(500)))
			return
		}

		file, handler, err := r.FormFile("image")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error retrieving image :/" + http.StatusText(500)))
			return
		}
		defer file.Close()

		fmt.Printf("Uploaded File: %+v\n", handler.Filename)
		fmt.Printf("File Size: %+v\n", handler.Size)
		fmt.Printf("MIME Header: %+v\n", handler.Header)

		id := uuid.New()
		f, err := os.OpenFile("data/"+id.String(), os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("Error saving image :/" + http.StatusText(500)))
			return
		}
		defer f.Close()
		io.Copy(f, file)

		sqlStatement := `INSERT INTO image(filename, id, name) VALUES ($1, $2, $3) RETURNING id`
		err = db.QueryRow(sqlStatement, handler.Filename, id, name).Scan(&id)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, id.String())
	}
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
			Filename: images[0].Filename,
			ID:       images[0].ID,
			Name:     images[0].Name,
		}

		json, err := json.Marshal(image)
		if err != nil {
			panic(err)
		}

		v := r.URL.Query()
		download := v.Get("download")
		if download == "true" {
			path := fmt.Sprintf("data/%s", image.Filename)
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
			disposition := fmt.Sprintf("attachment; filename=%s", image.Filename)
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

// GetImageHandler : handles retrival of all images from API request
func GetImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		images, err := GetImages(db)
		if err != nil {
			panic(err)
		}
		json, err := json.Marshal(images)
		if err != nil {
			panic(err)
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
		if e := rows.Scan(&image.Filename, &image.ID, &image.Name); e != nil {
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
		if e := rows.Scan(&image.Filename, &image.ID, &image.Name); e != nil {
			return nil, e
		}
		images = append(images, image)
	}
	return images, nil
}
