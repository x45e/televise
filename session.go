package televise

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	redisKeySession = "session"
	redisKeyVisit   = "visit"
	redisTimeout    = time.Second

	sessionCountPeriod  = 10 * time.Second
	maxSessionRetention = 30 * time.Second
)

type Identity struct {
	Key       string `json:"key"`
	Addr      string `json:"-"`
	UserAgent string `json:"-"`
}

func NewIdentity(r *http.Request) *Identity {
	addr := r.Header.Get("X-Forwarded-For")
	if addr == "" {
		addr = r.RemoteAddr
	}
	// ignore port
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	h := sha1.New()
	h.Write([]byte(addr))
	h.Write([]byte(r.UserAgent()))
	return &Identity{
		Key:       hex.EncodeToString(h.Sum(nil)),
		Addr:      host,
		UserAgent: r.UserAgent(),
	}
}

func (id Identity) Register(c *redis.Client) error {
	if c == nil {
		return errors.New("client nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	return c.HSet(ctx, redisKeySession, id.Key, id.Addr+"|"+id.UserAgent).Err()
}

func (id Identity) LogVisit(c *redis.Client) error {
	if c == nil {
		return errors.New("client nil")
	}
	if id.Key == "" || len(id.Key) > 64 {
		return errors.New("key invalid")
	}
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	err := c.ZAdd(ctx, redisKeyVisit, &redis.Z{Score: float64(now.UnixNano()), Member: id.Key}).Err()
	if err != nil {
		return err
	}
	return nil
}

// SessionCount returns the current number of active connections.
func SessionCount(c *redis.Client) (n int64, err error) {
	if c == nil {
		return -1, errors.New("client nil")
	}
	start := time.Now().Add(-sessionCountPeriod).UnixNano()
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	n, err = c.ZCount(ctx, redisKeyVisit, strconv.Itoa(int(start)), "+inf").Result()
	if err != nil {
		return -1, err
	}
	return n, nil
}

func PruneSessions(c *redis.Client) error {
	if c == nil {
		return errors.New("client nil")
	}
	start := time.Now().Add(-maxSessionRetention).UnixNano()
	ctx, cancel := context.WithTimeout(context.Background(), redisTimeout)
	defer cancel()
	return c.ZRemRangeByScore(ctx, redisKeyVisit, "-inf", strconv.FormatInt(start, 10)).Err()
}
