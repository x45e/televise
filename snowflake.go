package televise

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

const NilSnowflake Snowflake = 0

// Snowflake is used for unique identifiers.
type Snowflake int64

const (
	epoch int64 = 1420070400000
)

type snowflaker struct {
	pid int
	wid int
	mu  sync.Mutex
	seq int
}

func (r *snowflaker) next() (s Snowflake) {
	r.mu.Lock()
	defer r.mu.Unlock()
	dt := time.Now().UnixNano()/int64(time.Millisecond) - epoch
	r.seq++
	if r.seq > 1<<12-1 {
		r.seq = 0
	}
	s = Snowflake(dt << 22)
	s |= Snowflake((r.wid & 0x3E) << 17)
	s |= Snowflake((r.pid & 0x1F) << 17)
	s |= Snowflake(r.seq & 0xFFF)
	return s
}

var flaker *snowflaker

func init() {
	flaker = &snowflaker{
		pid: os.Getpid(),
	}
}

// GenerateSnowflake creates a new unique snowflake
func GenerateSnowflake() Snowflake {
	return flaker.next()
}

func ParseSnowflake(s string) (sf Snowflake) {
	(&sf).UnmarshalText([]byte(s))
	return sf
}

// Time returns when the snowflake was created
func (s Snowflake) Time() time.Time {
	ms := int64(s)>>22 + epoch
	sec := ms / 1000
	ns := ms % sec
	return time.Unix(sec, ns)
}

// WorkerID is not used
func (s Snowflake) WorkerID() int {
	return int(int64(s&0x3E0000) >> 17)
}

// ProcessID is the process id that generated the snowflake
func (s Snowflake) ProcessID() int {
	return int(int64(s&0x1F000) >> 12)
}

// Increment increases every time a new snowflake is generated
func (s Snowflake) Increment() int {
	return int(s & 0xFFF)
}

func (s Snowflake) String() string {
	return strconv.FormatInt(int64(s), 16)
}

// MarshalText implements encoding.TextMarshaler
func (s Snowflake) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (s *Snowflake) UnmarshalText(text []byte) error {
	n, err := strconv.ParseInt(string(text), 16, 64)
	if err != nil {
		return err
	}
	*s = Snowflake(n)
	return nil
}

// MarshalJSON implements json.Marshaler
func (s Snowflake) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (s *Snowflake) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return errors.New("invalid input")
	}
	n, err := strconv.ParseInt(string(b[1:len(b)-1]), 16, 64)
	if err != nil {
		return err
	}
	*s = Snowflake(n)
	return nil
}

// MarshalCQL implements gocql.Marshaler
func (s Snowflake) MarshalCQL(info gocql.TypeInfo) (b []byte, err error) {
	b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(s))
	return b, nil
}

// UnmarshalCQL implements gocql.Unmarshaler
func (s *Snowflake) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	n := binary.BigEndian.Uint64(data)
	*s = Snowflake(n)
	return nil
}
