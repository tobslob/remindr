package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

func GetIDParam(r *http.Request, paramName ...string) (uuid.UUID, error) {
	name := "id"
	if len(paramName) > 0 && paramName[0] != "" {
		name = paramName[0]
	}

	id := r.PathValue(name)
	return uuid.Parse(id)
}

func GetIDsInQuery(r *http.Request) ([]uuid.UUID, error) {
	idsParam := r.URL.Query().Get("ids")

	if idsParam == "" {
		return nil, errors.New("missing ids query parameter")
	}

	idStrs := strings.Split(idsParam, ",")
	var ids []uuid.UUID
	for _, idStr := range idStrs {
		id, err := uuid.Parse(strings.TrimSpace(idStr))
		if err != nil {
			return nil, errors.New("invalid id format in ids query parameter")
		}
		ids = append(ids, id)
	}
	return ids, nil
}
