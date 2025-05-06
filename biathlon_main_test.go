package tests

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "00:00:00.000"},
		{"milliseconds", 1500 * time.Millisecond, "00:00:01.500"},
		{"minutes", 2*time.Minute + 5*time.Second + 123*time.Millisecond, "00:02:05.123"},
		{"hours", time.Hour + 2*time.Minute + 3*time.Second + 4*time.Millisecond, "01:02:03.004"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestTruncateFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		decimals int
		want     float64
	}{
		{"round down", 1.23456, 2, 1.23},
		{"no decimals", 5.987, 0, 5},
		{"exact", 2.5, 1, 2.5},
		{"negative", -1.678, 1, -1.7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateFloat(tt.input, tt.decimals)
			if got != tt.want {
				t.Errorf("got %.3f, want %.3f", got, tt.want)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	data := `{
        "LapLen": 500,
        "Laps": 2,
        "PenaltyLen": 150,
        "StartDelta": "00:00:30"
    }`
	tmpFile.WriteString(data)
	tmpFile.Close()

	parseConfig(tmpFile.Name())

	if config.LapLen != 500 || config.Laps != 2 || config.PenaltyLen != 150 {
		t.Errorf("config not parsed correctly: %+v", config)
	}
}

func TestLogEvent(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	logEvent("12:00:00.000", "Test event")
	if !strings.Contains(buf.String(), "[12:00:00.000] Test event") {
		t.Errorf("log output missing expected text, got: %s", buf.String())
	}
}

func TestProcessEvent(t *testing.T) {
	competitors = make(map[int]*Competitor)
	config.Laps = 2
	config.LapLen = 500
	config.PenaltyLen = 150
	config.StartDelta = "00:00:10"

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	event := "[12:00:00.000] 1 42"
	processEvent(event)

	c, exists := competitors[42]
	if !exists {
		t.Fatal("competitor was not created")
	}
	if !c.Registered {
		t.Error("competitor not registered after event 1")
	}
	if !strings.Contains(buf.String(), "registered") {
		t.Errorf("log output missing expected text: %s", buf.String())
	}
}

func TestParseParams(t *testing.T) {
	tests := []struct {
		name       string
		matches    []string
		wantTime   string
		wantEvent  int
		wantCompID int
		wantTimeT  time.Time
		wantExtra  string
		wantErr    bool
	}{
		{
			name:       "valid input with extra",
			matches:    []string{"full", "12:00:00.000", "1", "42", "extra data"},
			wantTime:   "12:00:00.000",
			wantEvent:  1,
			wantCompID: 42,
			wantTimeT:  mustParseTime("12:00:00.000"),
			wantExtra:  "extra data",
			wantErr:    false,
		},
		{
			name:       "valid input without extra",
			matches:    []string{"full", "12:00:00.000", "2", "33"},
			wantTime:   "12:00:00.000",
			wantEvent:  2,
			wantCompID: 33,
			wantTimeT:  mustParseTime("12:00:00.000"),
			wantExtra:  "",
			wantErr:    false,
		},
		{
			name:    "invalid eventID",
			matches: []string{"full", "12:00:00.000", "notNumber", "33"},
			wantErr: true,
		},
		{
			name:    "invalid competitorID",
			matches: []string{"full", "12:00:00.000", "1", "notNumber"},
			wantErr: true,
		},
		{
			name:    "invalid time format",
			matches: []string{"full", "badtime", "1", "33"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTime, gotEvent, gotCompID, gotTimeT, gotExtra, err := parseParams(tt.matches)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParams() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTime != tt.wantTime {
					t.Errorf("timeStr = %v, want %v", gotTime, tt.wantTime)
				}
				if gotEvent != tt.wantEvent {
					t.Errorf("eventID = %v, want %v", gotEvent, tt.wantEvent)
				}
				if gotCompID != tt.wantCompID {
					t.Errorf("competitorID = %v, want %v", gotCompID, tt.wantCompID)
				}
				if !gotTimeT.Equal(tt.wantTimeT) {
					t.Errorf("eventTime = %v, want %v", gotTimeT, tt.wantTimeT)
				}
				if gotExtra != tt.wantExtra {
					t.Errorf("extra = %v, want %v", gotExtra, tt.wantExtra)
				}
			}
		})
	}
}

// helper function to panic on parse error in test setup
func mustParseTime(s string) time.Time {
	t, err := time.Parse(timeLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}
