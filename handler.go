package televise

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type contextKey int

func httpError(w http.ResponseWriter, err error, code int) {
	if err != nil {
		log.Println(err)
	}
	if code == 0 {
		code = http.StatusInternalServerError
	}
	http.Error(w, fmt.Sprintln(code, http.StatusText(code)), code)
}

func withContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func HandleInfo(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*sql.DB)
	id := NewIdentity(r)
	_, err := CreateOrUpdateSession(db, id)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	viewers, err := SessionCount(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	meta, err := MetadataDisplayList(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	info := struct {
		Viewers int64                    `json:"viewers"`
		Meta    map[string]MetadataValue `json:"meta"`
	}{
		Viewers: viewers,
		Meta:    meta,
	}
	json.NewEncoder(w).Encode(info)
}

func MetadataUpdate(w http.ResponseWriter, r *http.Request) {
	k := r.URL.Query().Get("k")
	if k == "" {
		httpError(w, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query().Get("v")
	var val *string
	if v == "" {
		val = nil
	} else {
		val = &v
	}
	db := r.Context().Value(KeyDB).(*sql.DB)
	err := MetadataSet(db, k, val)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
}

func HandleManifest(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*sql.DB)
	val, err := MetadataGet(db, "manifest")
	if err != nil {
		if err == sql.ErrNoRows {
			httpError(w, nil, http.StatusNotFound)
			return
		}
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	if val == nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	fmt.Fprint(w, *val)
}
