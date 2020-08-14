package consul

import "context"

func watchStale(ctx context.Context) bool {
	if ctx == nil {
		return true
	}

	stale, ok := ctx.Value("consul_allow_watch_stale").(bool)
	if !ok {
		return true
	}
	return stale
}
