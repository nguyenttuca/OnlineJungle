package dispatcher

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

type HealthResponse struct {
	Status        string          `json:"status"`
	Languages     map[string]bool `json:"languages"`
	ActiveJobs    int             `json:"active_jobs"`
	MaxConcurrent int32           `json:"max_concurrent"`
}

func RunHealthChecker(queries *sqlcdb.Queries) {
	log.Println("Health checker started...")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		<-ticker.C
		ctx := context.Background()

		nodes, err := queries.ListJudgeNodes(ctx)
		if err != nil {
			log.Printf("Health checker: error fetching nodes: %v", err)
			continue
		}

		for _, node := range nodes {
			if !node.IsActive {
				continue
			}

			go checkNodeHealth(ctx, queries, httpClient, node)
		}
	}
}

func checkNodeHealth(ctx context.Context, queries *sqlcdb.Queries, client *http.Client, node sqlcdb.JudgeNode) {
	url := node.BaseUrl + "/health"
	resp, err := client.Get(url)

	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Health check failed for node %d (%s): %v", node.ID, url, err)

		failures := node.ConsecutiveFailures + 1
		isHealthy := true
		if failures >= 3 {
			isHealthy = false
		}

		queries.UpdateJudgeNodeHealth(ctx, sqlcdb.UpdateJudgeNodeHealthParams{
			ID:                  node.ID,
			IsHealthy:           isHealthy,
			SupportedLanguages:  node.SupportedLanguages,
			MaxConcurrent:       node.MaxConcurrent,
			LastHealthCheckAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ConsecutiveFailures: failures,
		})
		return
	}
	defer resp.Body.Close()

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		log.Printf("Health check decode error for node %d: %v", node.ID, err)
		return
	}

	langBytes, _ := json.Marshal(health.Languages)

	err = queries.UpdateJudgeNodeHealth(ctx, sqlcdb.UpdateJudgeNodeHealthParams{
		ID:                  node.ID,
		IsHealthy:           true,
		SupportedLanguages:  langBytes,
		MaxConcurrent:       health.MaxConcurrent,
		LastHealthCheckAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ConsecutiveFailures: 0,
	})

	if err != nil {
		log.Printf("Failed to update health for node %d: %v", node.ID, err)
	}
}
