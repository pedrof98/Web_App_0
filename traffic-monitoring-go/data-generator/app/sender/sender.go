package sender

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"traffic-monitoring-go/data-generator/app/events"
)

type Sender struct {
	siemAPIURL string
}

func New(siemAPIURL string) *Sender {
	return &Sender{
		siemAPIURL: siemAPIURL,
	}
}

func (s *Sender) IsSIEMAvailable() bool {
	resp, err := http.Get(s.siemAPIURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (s *Sender) SendEvent(event events.Event) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	resp, err := http.Post(s.siemAPIURL+"/ingest", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error sending event: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error response from SIEM: %d", resp.StatusCode)
		return
	}

	// Successful send - only log ~5% of events to avoid flooding logs
	if rand.Intn(100) < 5 {
		log.Printf("Sent %s %s event: %s", event.Severity, event.Category, event.Message)
	}
}
