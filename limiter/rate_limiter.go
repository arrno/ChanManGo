package limiter

type LimitProvider interface {
	Allow(ip string) bool
}
