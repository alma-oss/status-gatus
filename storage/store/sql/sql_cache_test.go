package sql

import (
	"testing"
	"time"

	"github.com/TwiN/gatus/v5/config/endpoint"
	"github.com/TwiN/gatus/v5/storage/store/common/paging"
)

// A results page larger than maximumCachedResultsPageSize must not enter the write-through cache:
// cached pages are refreshed on every insert, so a whole-history page would make every insert
// re-read the whole history until the process restarts.
func TestStore_LargeResultsPagesAreNotCached(t *testing.T) {
	store, _ := NewStore("sqlite", t.TempDir()+"/TestStore_LargeResultsPagesAreNotCached.db", true, 100000, 100)
	defer store.Close()
	ep := &endpoint.Endpoint{Name: "name", Group: "group", URL: "https://example.org"}
	for i := 0; i < 5; i++ {
		_ = store.InsertEndpointResult(ep, &endpoint.Result{Success: true, Timestamp: time.Now().Add(time.Duration(i) * time.Second)})
	}
	small := paging.NewEndpointStatusParams().WithResults(1, maximumCachedResultsPageSize)
	large := paging.NewEndpointStatusParams().WithResults(1, maximumCachedResultsPageSize+1)
	if _, err := store.GetEndpointStatusByKey(ep.Key(), small); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetEndpointStatusByKey(ep.Key(), large); err != nil {
		t.Fatal(err)
	}
	if _, exists := store.writeThroughCache.Get(generateCacheKey(ep.Key(), small)); !exists {
		t.Errorf("expected page of %d results to be cached", maximumCachedResultsPageSize)
	}
	if _, exists := store.writeThroughCache.Get(generateCacheKey(ep.Key(), large)); exists {
		t.Errorf("expected page of %d results NOT to be cached", maximumCachedResultsPageSize+1)
	}
	// The large page must still be served correctly, just uncached
	status, err := store.GetEndpointStatusByKey(ep.Key(), large)
	if err != nil || len(status.Results) != 5 {
		t.Errorf("expected 5 results from the uncached large page, got %d (err=%v)", len(status.Results), err)
	}
}
