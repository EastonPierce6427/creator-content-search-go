package creatorsearch

import "testing"

func TestRankedResultsKeepSubscriberUpdates(t *testing.T) {
	cases := []struct {
		name    string
		results []SearchResult
		want    string
	}{
		{"highest score", []SearchResult{{ID: "post-1", Score: .91}, {ID: "post-2", Score: .72}}, "post-1"},
		{"subscriber filter", []SearchResult{{ID: "post-2", Score: .88, Content: map[string]any{"visibility": "subscriber"}}, {ID: "post-3", Score: .9, Content: map[string]any{"visibility": "public"}}}, "post-3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := bestResult(tc.results)
			if got.ID != tc.want {
				t.Fatalf("got %s want %s", got.ID, tc.want)
			}
		})
	}
}

func bestResult(results []SearchResult) SearchResult {
	var best SearchResult
	for _, r := range results {
		if r.Score > best.Score {
			best = r
		}
	}
	return best
}
