package main

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"
)

func setupCompetitor(id int) *Competitor {
	return &Competitor{
		ID: id,
	}
}

func captureLogs(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(log.Writer())
	f()
	return buf.String()
}

func TestHandleRegistration(t *testing.T) {
	c := setupCompetitor(1)
	logs := captureLogs(func() {
		handleRegistration(c, "09:00:00.000")
	})

	if !c.Registered {
		t.Errorf("expected competitor to be registered")
	}

	if !strings.Contains(logs, "The competitor(1) registered") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandlePlanningStartTime(t *testing.T) {
	c := setupCompetitor(2)
	planningTime := "10:00:00.000"

	logs := captureLogs(func() {
		handlePlanningStartTime(c, "09:00:00.000", planningTime)
	})

	if c.StartTimePlanned.Format(timeLayout) != planningTime {
		t.Errorf("expected StartTimePlanned to be %s, got %s", planningTime, c.StartTimePlanned.Format(timeLayout))
	}
	if !strings.Contains(logs, "was set by a draw to") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandleStartTime(t *testing.T) {
	config.StartDelta = "00:01:00"

	tests := []struct {
		name           string
		planned        string
		actual         string
		expectDQ       bool
		expectedLogMsg string
	}{
		{
			name:           "Starts in time",
			planned:        "09:00:00.000",
			actual:         "09:00:59.000",
			expectDQ:       false,
			expectedLogMsg: "has started",
		},
		{
			name:           "Starts too late",
			planned:        "09:00:00.000",
			actual:         "09:02:00.000",
			expectDQ:       true,
			expectedLogMsg: "is disqualified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupCompetitor(3)
			planned, _ := time.Parse(timeLayout, tt.planned)
			actual, _ := time.Parse(timeLayout, tt.actual)

			c.StartTimePlanned = planned

			logs := captureLogs(func() {
				handleStartTime(c, tt.actual, actual)
			})

			if tt.expectDQ != c.Disqualified {
				t.Errorf("expected disqualified=%v, got %v", tt.expectDQ, c.Disqualified)
			}
			if !strings.Contains(logs, tt.expectedLogMsg) {
				t.Errorf("unexpected log output: %s", logs)
			}
		})
	}
}

func TestHandleFiringRange(t *testing.T) {
	c := setupCompetitor(4)
	logs := captureLogs(func() {
		handleFiringRange(c, "09:10:00.000", "A")
	})

	if !strings.Contains(logs, "is on the firing range(A)") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandleShooting(t *testing.T) {
	c := setupCompetitor(5)
	logs := captureLogs(func() {
		handleShooting(c, "09:11:00.000", "1")
	})

	if c.Hits != 1 {
		t.Errorf("expected Hits=1, got %d", c.Hits)
	}
	if !strings.Contains(logs, "has been hit by competitor(5)") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandleLeftFiringRange(t *testing.T) {
	c := setupCompetitor(6)
	logs := captureLogs(func() {
		handleLeftFiringRange(c, "09:12:00.000")
	})

	if c.Shots != 5 {
		t.Errorf("expected Shots=5, got %d", c.Shots)
	}
	if !strings.Contains(logs, "left the firing range") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandleEnteringPenalty(t *testing.T) {
	c := setupCompetitor(7)
	eventTime := time.Now()

	logs := captureLogs(func() {
		handleEnteringPenalty(c, "09:15:00.000", eventTime)
	})

	if c.PenaltyStart != eventTime {
		t.Errorf("expected PenaltyStart to be set")
	}
	if !strings.Contains(logs, "entered the penalty laps") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandleLeftPenalty(t *testing.T) {
	c := setupCompetitor(8)
	start := time.Now()
	end := start.Add(30 * time.Second)
	c.PenaltyStart = start

	logs := captureLogs(func() {
		handleLeftPenalty(c, "09:16:00.000", end)
	})

	if c.PenaltyTime < 30*time.Second {
		t.Errorf("expected PenaltyTime to be >=30s, got %v", c.PenaltyTime)
	}
	if !strings.Contains(logs, "left the penalty laps") {
		t.Errorf("unexpected log output: %s", logs)
	}
}

func TestHandleEndMainLap(t *testing.T) {
	config.Laps = 2

	c := setupCompetitor(9)
	start := time.Now()
	lap1End := start.Add(1 * time.Minute)
	lap2End := lap1End.Add(1 * time.Minute)

	logs := captureLogs(func() {
		c.StartTimePlanned = start
		handleEndMainLap(c, lap1End.Format(timeLayout), lap1End)
	})
	if len(c.LapTimes) != 1 {
		t.Errorf("expected 1 lap time, got %d", len(c.LapTimes))
	}
	if !strings.Contains(logs, "ended the main lap") {
		t.Errorf("unexpected log output (lap 1): %s", logs)
	}

	logs = captureLogs(func() {
		handleEndMainLap(c, lap2End.Format(timeLayout), lap2End)
	})

	if !c.Finished {
		t.Errorf("expected competitor to be finished")
	}
	if !strings.Contains(logs, "has finished") {
		t.Errorf("unexpected log output (lap 2): %s", logs)
	}
}

func TestHandleCantContinue(t *testing.T) {
	c := setupCompetitor(10)
	logs := captureLogs(func() {
		handleCantContinue(c, "09:20:00.000", "injury")
	})

	if !c.NotFinished {
		t.Errorf("expected NotFinished=true")
	}
	if !strings.Contains(logs, "can`t continue: injury") {
		t.Errorf("unexpected log output: %s", logs)
	}
}
