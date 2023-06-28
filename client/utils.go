package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hatchify/errors"
)

func handleError(res *http.Response) (err error) {
	var resp apiResp
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		err = fmt.Errorf("error decoding response: %v", err)
		return
	}

	return resp.Errors.Err()
}

type apiResp struct {
	Data   interface{}      `json:"data"`
	Errors errors.ErrorList `json:"errors"`
}
