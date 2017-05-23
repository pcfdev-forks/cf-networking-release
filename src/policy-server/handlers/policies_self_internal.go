package handlers

import (
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/go-db-helpers/marshal"
	"code.cloudfoundry.org/lager"
)

type PoliciesSelfInternal struct {
	Logger        lager.Logger
	Store         store
	Marshaler     marshal.Marshaler
	Unmarshaler   marshal.Unmarshaler
	ErrorResponse errorResponse
}

func (h *PoliciesSelfInternal) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Logger.Debug("internal request made to list policies", lager.Data{"URL": req.URL, "RemoteAddr": req.RemoteAddr})

	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.ErrorResponse.BadRequest(w, err, "policies-self", "failed reading request body")
		return
	}

	var payload struct {
		ID          string `json:"id"`
		Destination string `json:"protocol"`
		Port        int    `json:"port"`
	}
	err = h.Unmarshaler.Unmarshal(bodyBytes, &payload)
	if err != nil {
		h.ErrorResponse.BadRequest(w, err, "policies-self", "invalid values passed to API")
		return
	}
	// policy := models.Policy{}
	// TODO validate policy
	// TODO create policy
	// TODO return tag
}
