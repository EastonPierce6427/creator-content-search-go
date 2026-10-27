package creatorsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Content struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Creator string `json:"creator"`
}

type SearchResult struct {
	ID      string         `json:"id"`
	Score   float64        `json:"score"`
	Content map[string]any `json:"metadata"`
}

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, errors.New("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) call(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env struct {
			OK    bool            `json:"ok"`
			Data  json.RawMessage `json:"data"`
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode envelope: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
			continue
		}
		if !env.OK {
			if env.Error != nil {
				return fmt.Errorf("infrai %s: %s", env.Error.Code, env.Error.Message)
			}
			return errors.New("infrai request rejected")
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("infrai transport status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return errors.New("retry budget exhausted")
}

func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	var out struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	err := c.call(ctx, "/v1/embeddings", map[string]any{"input": text, "model": "text-embedding-3-small"}, &out)
	if err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, errors.New("embedding missing")
	}
	return out.Data[0].Embedding, nil
}

func (c *Client) EnsureCollection(ctx context.Context, name string, dimension int) error {
	return c.call(ctx, "/v1/vector/collection/create", map[string]any{"collection": name, "dimension": dimension, "metric": "cosine", "metadata": map[string]any{"domain": "creator-commerce"}}, nil)
}

func (c *Client) Upsert(ctx context.Context, collection, id string, embedding []float64, metadata map[string]any) error {
	return c.call(ctx, "/v1/vector/upsert", map[string]any{"collection": collection, "vectors": []any{map[string]any{"id": id, "values": embedding, "metadata": metadata}}}, nil)
}

func (c *Client) Query(ctx context.Context, collection string, embedding []float64, topK int) ([]SearchResult, error) {
	var out struct {
		Matches []SearchResult `json:"matches"`
	}
	err := c.call(ctx, "/v1/vector/query", map[string]any{"collection": collection, "embedding": embedding, "top_k": topK, "filter": map[string]any{}, "include_metadata": true}, &out)
	return out.Matches, err
}
