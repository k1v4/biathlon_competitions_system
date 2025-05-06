package main

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
	if len(planningStart) == 0 {
		log.Fatal("2. start time: wrong format of extra field")
	}

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

func handleFiringRange(c *Competitor, timeStr, firing string) {
	if len(firing) == 0 {
		log.Fatal("5. firing range: wrong format of extra field")
	}

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) is on the firing range(%s)", c.ID, firing))
}

func handleShooting(c *Competitor, timeStr, aim string) {
	if len(aim) == 0 {
		log.Fatal("6. the target has been hit: wrong format of extra field")
	}

	c.Hits++

	logEvent(timeStr, fmt.Sprintf("The target(%s) has been hit by competitor(%d)", aim, c.ID))
}

func handleLeftFiringRange(c *Competitor, timeStr string) {
	c.Shots += 5

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) left the firing range", c.ID))
}

func handleEnteringPenalty(c *Competitor, timeStr string, eventTime time.Time) {
	c.PenaltyStart = eventTime

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) entered the penalty laps", c.ID))
}

func handleLeftPenalty(c *Competitor, timeStr string, eventTime time.Time) {
	penaltyDuration := eventTime.Sub(c.PenaltyStart)
	c.PenaltyTime += penaltyDuration

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) left the penalty laps", c.ID))
}

func handleEndMainLap(c *Competitor, timeStr string, eventTime time.Time) {
	var lapDuration time.Duration

	if len(c.LapTimes) == 0 {
		lapDuration = eventTime.Sub(c.StartTimePlanned)
	} else {
		lapDuration = eventTime.Sub(c.CurrentLapStart)
	}

	c.LapTimes = append(c.LapTimes, lapDuration)
	c.CurrentLapStart = eventTime

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) ended the main lap", c.ID))

	if len(c.LapTimes) == config.Laps {
		c.Finished = true

		parseEnd, err := time.Parse(timeLayout, timeStr)
		if err != nil {
			log.Fatal(fmt.Errorf("error parsing time: %s", err))
		}

		c.FinishTime = parseEnd

		logEvent(eventTime.Format(timeLayout), fmt.Sprintf("The competitor(%d) has finished", c.ID))
	}
}

func handleCantContinue(c *Competitor, timeStr, comment string) {
	if len(comment) == 0 {
		log.Fatal("11. the competitor cant continue: wrong format of extra field")
	}

	c.NotFinished = true

	logEvent(timeStr, fmt.Sprintf("The competitor(%d) can`t continue: %s", c.ID, comment))
}
