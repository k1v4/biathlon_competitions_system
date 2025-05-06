package main

import "C"
import (
	"fmt"
	"log"
	"time"
)

func handleRegistration(c *Competitor, timeStr string) {
	c.Registered = true

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) registered", c.ID))
}

func handlePlanningStartTime(c *Competitor, timeStr, planningStart string) {
	var err error
	c.StartTimePlanned, err = time.Parse(timeLayout, planningStart)
	if err != nil {
		log.Fatalf("could not parse start time: %s", planningStart)

	}

	logEvent(timeStr, fmt.Sprintf("The start time for the competitor(%d) was set by a draw to %s", c.ID, planningStart))
}

func handleStartTime(c *Competitor, timeStr string, eventTime time.Time) {
	t, err := time.Parse("15:04:05", config.StartDelta)
	if err != nil {
		log.Printf("error parsing start delta time: %s\n", err)

		return
	}

	parseStartingDelta := time.Duration(t.Hour())*time.Hour +
		time.Duration(t.Minute())*time.Minute +
		time.Duration(t.Second())*time.Second

	if eventTime.Compare(c.StartTimePlanned.Add(parseStartingDelta)) == 1 {
		c.Disqualified = true

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) is disqualified", c.ID))

		return
	}

	c.ActualStartTime = eventTime
	c.CurrentLapStart = eventTime

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) has started", c.ID))
}
