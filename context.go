package televise

type ContextKey int

const (
	KeyDB ContextKey = iota
	KeyCount
	KeyTitle
	KeyRedis
)
