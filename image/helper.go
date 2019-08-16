package main

import (
	"database/sql"
)

// Image : struct for image in postgres
type Image struct {
	id   string
	name string
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
		if e := rows.Scan(&image.id, &image.name); e != nil {
			return nil, e
		}
		images = append(images, image)
	}
	return images, nil
}
