package handlers

import (
	"io/ioutil"
	"net/http"
	"policy-server/models"

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
		ID   string `json:"id"`
		Port int    `json:"port"`
	}
	err = h.Unmarshaler.Unmarshal(bodyBytes, &payload)
	if err != nil {
		h.ErrorResponse.BadRequest(w, err, "policies-self", "invalid values passed to API")
		return
	}

	policies := []models.Policy{
		{
			Source: models.Source{
				ID: payload.ID,
			},
			Destination: models.Destination{
				ID:       payload.ID,
				Port:     payload.Port,
				Protocol: "tcp",
			},
		},
	}

	err = h.Store.Create(policies)
	if err != nil {
		h.ErrorResponse.BadRequest(w, err, "policies-self", "creating self policy")
		return
	}

	tags, err := h.Store.Tags()
	if err != nil {
		h.ErrorResponse.BadRequest(w, err, "policies-self", "getting tags")
		return
	}

	var newTag string
	for _, t := range tags {
		if t.ID == payload.ID {
			newTag = t.Tag
			break
		}
	}

	if newTag == "" {
		h.ErrorResponse.InternalServerError(w, err, "policies-self", "no tag found")
		return
	}

	var tag struct {
		Tag string `json:"tag"`
	}
	tag.Tag = newTag
	body, err := h.Marshaler.Marshal(tag)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	return
}
