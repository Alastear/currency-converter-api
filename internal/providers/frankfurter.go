package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type FrankfurterLatest struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

func FetchFrankfurter(base string) (map[string]string, time.Time, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("https://api.frankfurter.dev/v1/latest?base=%s", base)
	resp, err := client.Get(url)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, time.Time{}, fmt.Errorf("frankfurter Status %d", resp.StatusCode)
	}
	var payload FrankfurterLatest
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, time.Time{}, err
	}
	out := map[string]string{}
	for k, v := range payload.Rates {
		out[k] = fmt.Sprintf("%f", v)
	}
	t, _ := time.Parse("2006-01-02", payload.Date)
	if t.IsZero() {
		t = time.Now()
	}
	return out, t, nil
}
