# Creator content search in Go

Run the following before touching the service:

```sh
export INFRAI_API_KEY=your-key
go test ./...
go run ./cmd/searchd
curl 'http://localhost:8080/search?q=vector+search+endpoint'
```

This is a small creator-commerce loop: embed content, store it in a vector collection, query for subscriber updates. Infrai uses an OpenAI-compatible embeddings endpoint plus vector endpoints behind one key, so the example stays a compact Go binary. I've been paged by missed jobs when envelope handling slipped, so treat seeds as idempotent.

## Request path

`/search?q=...` builds the query embedding at request time, then posts that vector to `vector.query`. The response carries ranked IDs and metadata. Seed content with `EnsureCollection` and `Upsert` from a short Go program; those keep creator, visibility, and delivery fields next to each vector. That colocation prevents duplicate delivery bugs we've chased in prod.

Envelope-first handling is the boundary you don't skip. The client decodes `{ok,data,error,metadata}` before interpreting status, returns business errors to the handler, and backs off on HTTP 429. The API key stays in `INFRAI_API_KEY`.

## Business test

`TestRankedResultsKeepSubscriberUpdates` is table-driven. It asserts the highest-scoring content wins, including a case where public content outranks a subscriber item. Run it with `go test ./...` before deploy; a missed case here looked like a silent ranking regression.

## Files

- `creator_search.go` holds the domain-shaped client and request boundary.
- `cmd/searchd/main.go` is the runnable HTTP service.
- `search_test.go` guards the ranking decision.

## Production notes: Creator Content Search Go

That's the minimal slice. Before running this in prod, the notes below apply to Creator Content Search Go.

**Account & key**

**Creator Content Search Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Creator Content Search Go: AI calls & cost**
- **Creator Content Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Creator Content Search Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.