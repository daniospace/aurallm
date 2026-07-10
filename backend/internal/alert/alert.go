package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gateway/internal/cache"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	alertMutex sync.RWMutex
	sentAlerts = make(map[string]bool)
)

// SendBudgetAlert dispatches a Slack Webhook alert in the background.
// It uses the Cache to deduplicate alerts so Slack is never spammed.
func SendBudgetAlert(ctx context.Context, c cache.Cache, teamID string, teamName string, used float64, limit float64, thresholdPercent int) {
	if isAlertAlreadySent(teamID, thresholdPercent) {
		return
	}

	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		log.Printf("[ALERT NOTICE] Team %q has consumed %d%% of budget! Spend: $%.2f / Limit: $%.2f", teamName, thresholdPercent, used, limit)
		return
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		payload := map[string]interface{}{
			"text": fmt.Sprintf("🚨 AuraLLM FinOps Alert: Team *%s* has consumed *%d%%* of their budget! Used: $%.2f / Limit: $%.2f", teamName, thresholdPercent, used, limit),
			"blocks": []interface{}{
				map[string]interface{}{
					"type": "header",
					"text": map[string]interface{}{
						"type": "plain_text",
						"text": "🚨 AuraLLM FinOps Alert",
					},
				},
				map[string]interface{}{
					"type": "section",
					"text": map[string]interface{}{
						"type": "mrkdwn",
						"text": fmt.Sprintf("Budget threshold breached for team *%s*!\n*Utilization:* `%d%%`\n*MTD Spend:* `$%.2f` \n*Monthly Limit:* `$%.2f`", teamName, thresholdPercent, used, limit),
					},
				},
			},
		}

		bodyBytes, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(bgCtx, http.MethodPost, webhookURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Failed to dispatch Slack webhook: %v", err)
			return
		}
		defer resp.Body.Close()
		
		log.Printf("Slack Webhook Alert dispatched successfully for team %s at %d%% threshold", teamName, thresholdPercent)
	}()
}

func isAlertAlreadySent(teamID string, threshold int) bool {
	key := fmt.Sprintf("%s:%d", teamID, threshold)
	
	alertMutex.RLock()
	sent := sentAlerts[key]
	alertMutex.RUnlock()
	
	if sent {
		return true
	}
	
	alertMutex.Lock()
	sentAlerts[key] = true
	alertMutex.Unlock()
	return false
}
