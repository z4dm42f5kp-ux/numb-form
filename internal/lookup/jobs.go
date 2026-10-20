package lookup

import (
	"context"
	"sync"
	"time"
)

// firstVerified reads candidate websites, three at a time, and returns
// the first (in candidate order) that names the company. A site the user
// typed is accepted even when unverified.
func firstVerified(ctx context.Context, c *Client, q Query, cands []string, typed bool) (SiteResult, bool) {
	if len(cands) == 0 {
		return SiteResult{}, false
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	results := make([]*SiteResult, len(cands))
	done := make([]chan struct{}, len(cands))
	sem := make(chan struct{}, 3)
	for i := range cands {
		done[i] = make(chan struct{})
		go func(i int) {
			defer close(done[i])
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			r := ReadSite(ctx, c, cands[i], q, typed)
			results[i] = &r
		}(i)
	}
	var notes []string
	for i := range cands {
		<-done[i]
		r := results[i]
		if r == nil {
			continue
		}
		if r.Verified || typed {
			cancel()
			r.Notes = append(notes, r.Notes...)
			return *r, r.Verified || typed
		}
		notes = append(notes, r.Notes...)
	}
	return SiteResult{Result: Result{Notes: notes}}, false
}

func runJobs(ctx context.Context, jobs []func(context.Context) (Result, error)) Result {
	results := make([]Result, len(jobs))
	var wg sync.WaitGroup
	for i := range jobs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			jctx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			r, err := jobs[i](jctx)
			if err != nil {
				r.Notes = append(r.Notes, sanitizeErr(err))
			}
			results[i] = r
		}(i)
	}
	wg.Wait()
	var all Result
	for _, r := range results {
		all.merge(r)
	}
	return all
}
