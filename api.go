// Package provides standoff function that fetches photo from Cataas Public API(https://cataas.com)
// You can visit website for more information
package catphotofetch

import (
	"io"
	"net/http"
)

const baseURL = "https://cataas.com/cat"

func FetchRandomPhoto() (*Image, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Image{
		ContentType: contentType,
		Data:        body,
	}, nil
}