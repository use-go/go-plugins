package consul

import "context"

func watchStale(ctx context.Context) bool {
	if ctx == nil {
		return true
	}

	stale, ok := ctx.Value(watchStaleKey{}).(bool)
	if !ok {
		return true
	}
	return stale
}
