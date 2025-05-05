package main

import "time"

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

type Competitor struct {
	ID               int
	Registered       bool
	StartTimePlanned time.Time
	ActualStartTime  time.Time
	Finished         bool
	Disqualified     bool
	NotFinished      bool
	LapTimes         []time.Duration
	PenaltyTime      time.Duration
	Hits             int
	Shots            int
	CurrentLapStart  time.Time
	PenaltyStart     time.Time
	PenaltyLaps      int
	ReportRows       []string
}

type Result struct {
	CompetitorID int
	TotalTime    time.Duration
	Competitor   *Competitor
	Status       string
}
