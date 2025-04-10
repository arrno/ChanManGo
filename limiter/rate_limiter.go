package limiter

type LimitProvider interface {
	Allow(identity string) bool
}
