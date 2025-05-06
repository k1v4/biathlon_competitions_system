package main

import (
	"os"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"Milliseconds only", time.Millisecond * 123, "00:00:00.123"},
		{"Seconds and milliseconds", time.Second*1 + time.Millisecond*50, "00:00:01.050"},
		{"Minutes, seconds, milliseconds", time.Minute*2 + time.Second*3 + time.Millisecond*4, "00:02:03.004"},
		{"Hours, minutes, seconds, milliseconds", time.Hour*1 + time.Minute*2 + time.Second*3 + time.Millisecond*400, "01:02:03.400"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDuration(%v) = %v; want %v", tt.duration, got, tt.expected)
			}
		})
	}
}

func TestTruncateFloat(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		decimals int
		expected float64
	}{
		{"Pi 3 decimals", 3.14159, 3, 3.141},
		{"Almost 2, 2 decimals", 1.9999, 2, 1.99},
		{"Whole number", 123.456, 0, 123.0},
		{"Small number", 0.00999, 2, 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateFloat(tt.value, tt.decimals)
			if got != tt.expected {
				t.Errorf("truncateFloat(%v, %d) = %v; want %v", tt.value, tt.decimals, got, tt.expected)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected Config
	}{
		{
			name: "Basic config",
			content: `{
				"laps": 2,
				"lapLen": 4000,
				"penaltyLen": 150,
				"firingLines": 2,
				"start": "10:00:00",
				"startDelta": "00:01:00"
			}`,
			expected: Config{
				Laps:        2,
				LapLen:      4000,
				PenaltyLen:  150,
				FiringLines: 2,
				Start:       "10:00:00",
				StartDelta:  "00:01:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "config-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			tmpFile.Close()

			parseConfig(tmpFile.Name())

			if config != tt.expected {
				t.Errorf("parsed config = %+v; want %+v", config, tt.expected)
			}
		})
	}
}

func TestProcessEvent(t *testing.T) {
	tests := []struct {
		name         string
		eventLines   []string
		expectedFunc func(t *testing.T)
	}{
		{
			name:       "Register competitor",
			eventLines: []string{"[09:00:00.000] 1 42"},
			expectedFunc: func(t *testing.T) {
				c, ok := competitors[42]
				if !ok {
					t.Fatalf("competitor 42 not found after registration")
				}
				if !c.Registered {
					t.Errorf("competitor should be registered")
				}
			},
		},
		{
			name:       "Hit targets twice",
			eventLines: []string{"[09:10:00.000] 1 1", "[09:10:01.000] 6 1 1", "[09:10:02.000] 6 1 2"},
			expectedFunc: func(t *testing.T) {
				c := competitors[1]
				if c.Hits != 2 {
					t.Errorf("expected 2 hits, got %d", c.Hits)
				}
			},
		},
		{
			name:       "Penalty time is recorded",
			eventLines: []string{"[09:15:00.000] 1 1", "[09:15:00.000] 8 1", "[09:17:30.000] 9 1"},
			expectedFunc: func(t *testing.T) {
				c := competitors[1]
				expected := (time.Minute * 2) + (time.Second * 30)
				if c.PenaltyTime != expected {
					t.Errorf("penalty time mismatch, got %v, want %v", c.PenaltyTime, expected)
				}
			},
		},
		{
			name: "Mark as not finished",
			eventLines: []string{
				"[09:00:00.000] 1 77",
				"[09:30:00.000] 11 77 some_reason",
			},
			expectedFunc: func(t *testing.T) {
				c := competitors[77]
				if !c.NotFinished {
					t.Errorf("expected competitor to be marked as NotFinished")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCompetitors()
			for _, line := range tt.eventLines {
				processEvent(line)
			}
			tt.expectedFunc(t)
		})
	}
}

// helper to reset competitors map
func resetCompetitors() {
	competitors = make(map[int]*Competitor)
}
