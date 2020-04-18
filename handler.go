package televise

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	count := r.Context().Value(KeyCount)
	viewers := int64(0)
	if count != nil {
		n := r.Context().Value(KeyCount).(*int64)
		if n != nil {
			viewers = *n
		}
	}
	/*viewers, err := SessionCount(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}*/
	/*meta, err := MetadataDisplayList(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}*/
	meta := make(map[string]MetadataValue)
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
	if r.URL.Query().Get("poll") == "true" && val != nil {
		id, err := InsertOption(db, *val)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
		val := strconv.FormatInt(id, 10)
		err = MetadataSet(db, k+"_id", &val)
		if err != nil {
			httpError(w, err, http.StatusInternalServerError)
			return
		}
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

func HandleViewers(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*sql.DB)
	n, err := SessionCount(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, n)
}

func HandleLastResults(w http.ResponseWriter, r *http.Request) {
	db := r.Context().Value(KeyDB).(*sql.DB)
	title, votes, err := LastVote(db)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(struct {
		Title string `json:"title"`
		Votes int64  `json:"votes"`
	}{
		Title: title,
		Votes: votes,
	})
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
}
