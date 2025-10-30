package catphotofetch

import (
	"net/http"
	"time"
)

const defaultTimeout = time.Second * 10

func init() {
	http.DefaultClient.Timeout = defaultTimeout
}