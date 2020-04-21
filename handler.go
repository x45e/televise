package televise

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gocql/gocql"
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

type sessionData struct {
	ID Snowflake `json:"id"`
}

func HandleSession(w http.ResponseWriter, r *http.Request) {
	sf := r.Context().Value(KeySnowflaker).(*snowflaker)
	db := r.Context().Value(KeyDB).(*gocql.Session)
	id := NewIdentity(sf, r)
	err := CreateSession(db, id)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(sessionData{ID: id.ID})
}

/*
func HandleInfo(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*gocql.Session)
	id := FindIdentity(r)
	go LogVisit(db, id)
	count := r.Context().Value(KeyCount)
	viewers := int64(0)
	if count != nil {
		n := r.Context().Value(KeyCount).(*int64)
		if n != nil {
			viewers = *n
		}
	}
	ctxtitle := r.Context().Value(KeyTitle)
	title := ""
	if ctxtitle != nil {
		v := r.Context().Value(KeyTitle).(*string)
		if v != nil {
			title = *v
		}
	}
	meta := make(map[string]MetadataValue)
	meta["movie"] = MetadataValue{Value: &title}
	info := struct {
		Viewers int64                    `json:"viewers"`
		Meta    map[string]MetadataValue `json:"meta"`
	}{
		Viewers: viewers,
		Meta:    meta,
	}
	json.NewEncoder(w).Encode(info)
}*/

func MetadataUpdate(w http.ResponseWriter, r *http.Request) {
	k := r.URL.Query().Get("k")
	if k == "" {
		httpError(w, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query().Get("v")
	if v == "" {
		httpError(w, nil, http.StatusBadRequest)
		return
	}
	db := r.Context().Value(KeyDB).(*gocql.Session)
	err := MetadataSet(db, k, v)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
}

func HandleManifest(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*gocql.Session)
	val, err := MetadataGet(db, "manifest")
	if err != nil {
		// silently report error by not printing any text
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, val)
}

/*
func HandleCastVote(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*sql.DB)
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		httpError(w, nil, http.StatusBadRequest)
		return
	}
	err = CastVote(db, NewIdentity(r).Key, id)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
*/

func HandleViewers(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*gocql.Session)
	n, err := VisitorCount(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, n)
}
