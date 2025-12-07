package websocket

import (
	"encoding/json"
	"time"
)

// WSEventType represents different types of WebSocket events
type WSEventType string

const (
	EventMatchUpdate      WSEventType = "match_update"
	EventPredictionUpdate WSEventType = "prediction_update"
	EventLiveScore        WSEventType = "live_score"
	EventNewPrediction    WSEventType = "new_prediction"
	EventError            WSEventType = "error"
	EventSubscribed       WSEventType = "subscribed"
	EventUnsubscribed     WSEventType = "unsubscribed"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      WSEventType     `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// SubscribeMessage represents a subscription request
type SubscribeMessage struct {
	Room string `json:"room"` // e.g., "match:123", "competition:456"
}

// UnsubscribeMessage represents an unsubscription request
type UnsubscribeMessage struct {
	Room string `json:"room"`
}

// MatchUpdatePayload represents match update data
type MatchUpdatePayload struct {
	MatchID int         `json:"matchId"`
	Status  string      `json:"status"`
	Score   interface{} `json:"score"`
}

// PredictionUpdatePayload represents prediction update data
type PredictionUpdatePayload struct {
	PredictionID string  `json:"predictionId"`
	MatchID      int     `json:"matchId"`
	Status       string  `json:"status"`
	HomeWinProb  float64 `json:"homeWinProb"`
	DrawProb     float64 `json:"drawProb"`
	AwayWinProb  float64 `json:"awayWinProb"`
	Confidence   float64 `json:"confidence"`
}

// LiveScorePayload represents live score update data
type LiveScorePayload struct {
	MatchID   int    `json:"matchId"`
	HomeScore int    `json:"homeScore"`
	AwayScore int    `json:"awayScore"`
	Minute    int    `json:"minute"`
	Status    string `json:"status"`
}

// NewMessage creates a new WebSocket message
func NewMessage(eventType WSEventType, payload interface{}) (*WSMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &WSMessage{
		Type:      eventType,
		Payload:   payloadBytes,
		Timestamp: time.Now(),
	}, nil
}
