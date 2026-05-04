package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "sequi/internal/db/gen"
)

const pollInterval = 15 * time.Second

// srsStreamsResponse is the JSON structure returned by SRS /api/v1/streams.
type srsStreamsResponse struct {
	Code    int         `json:"code"`
	Streams []srsStream `json:"streams"`
}

type srsStream struct {
	Name    string `json:"name"` // stream key
	Clients int32  `json:"clients"`
}

// ViewerPoller periodically fetches viewer counts from SRS and updates the DB.
type ViewerPoller struct {
	queries    *db.Queries
	srsAPIURL  string
	httpClient *http.Client
}

func NewViewerPoller(pool *pgxpool.Pool, srsAPIURL string) *ViewerPoller {
	return &ViewerPoller{
		queries:   db.New(pool),
		srsAPIURL: srsAPIURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Run starts the polling loop. It blocks until ctx is cancelled.
func (p *ViewerPoller) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				log.Printf("viewer poller: %v", err)
			}
		}
	}
}

func (p *ViewerPoller) poll(ctx context.Context) error {
	srsStreams, err := p.fetchSRSStreams()
	if err != nil {
		return fmt.Errorf("fetch SRS streams: %w", err)
	}

	// Build a map: stream_key → client count
	counts := make(map[string]int32, len(srsStreams))
	for _, s := range srsStreams {
		counts[s.Name] = s.Clients
	}

	// For each active stream key, update viewer count in DB
	for key, clientCount := range counts {
		sk, err := p.queries.GetStreamKeyByKey(ctx, key)
		if err != nil {
			continue // key not in our DB, skip
		}
		stream, err := p.queries.GetStreamByUserID(ctx, sk.UserID)
		if err != nil {
			continue
		}
		if err := p.queries.UpdateStreamViewerCount(ctx, db.UpdateStreamViewerCountParams{
			ID:          stream.ID,
			ViewerCount: clientCount,
		}); err != nil {
			log.Printf("viewer poller: update stream %s: %v", key, err)
		}
	}

	// Reset viewer count for live streams that SRS no longer reports
	liveStreams, err := p.queries.ListLiveStreams(ctx)
	if err != nil {
		return fmt.Errorf("list live streams: %w", err)
	}
	for _, s := range liveStreams {
		sk, err := p.queries.GetActiveStreamKey(ctx, s.UserID)
		if err != nil {
			continue
		}
		if _, found := counts[sk.Key]; !found {
			_ = p.queries.UpdateStreamViewerCount(ctx, db.UpdateStreamViewerCountParams{
				ID:          s.ID,
				ViewerCount: 0,
			})
		}
	}

	return nil
}

func (p *ViewerPoller) fetchSRSStreams() ([]srsStream, error) {
	resp, err := p.httpClient.Get(p.srsAPIURL + "/streams")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result srsStreamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("SRS API error code %d", result.Code)
	}
	return result.Streams, nil
}
