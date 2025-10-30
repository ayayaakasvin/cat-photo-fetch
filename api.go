package catphotofetch

import (
	"io"
	"net/http"
)

const baseURL = "https://cataas.com/cat"

func FetchRandomPhoto() ([]byte, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}