package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var (
	config      Config
	competitors = make(map[int]*Competitor)
	timeLayout  = "15:04:05.000"
)

func formatDuration(d time.Duration) string {
	ms := d.Milliseconds() % 1000
	sec := int(d.Seconds()) % 60
	mins := int(d.Minutes()) % 60
	hour := int(d.Hours())

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hour, mins, sec, ms)
}

func parseConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		log.Fatal(err)
	}
}

func logEvent(timeStr string, text string) {
	fmt.Printf("[%s] %s\n", timeStr, text)
}

func processEvent(event string) {
	re := regexp.MustCompile(`\[(.*?)\] (\d+) (\d+)(?: (.*))?`)
	matches := re.FindStringSubmatch(event)

	if len(matches) < 4 {
		return
	}

	timeStr, eventIDStr, competitorIDStr := matches[1], matches[2], matches[3]

	extra := ""
	if len(matches) > 4 {
		extra = matches[4]
	}

	eventID, _ := strconv.Atoi(eventIDStr)
	competitorID, _ := strconv.Atoi(competitorIDStr)
	eventTime, _ := time.Parse(timeLayout, timeStr)

	c, ok := competitors[competitorID]
	if !ok {
		c = &Competitor{ID: competitorID}
		competitors[competitorID] = c
	}

	switch eventID {
	case 1:
		c.Registered = true

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) registered", competitorID))
	case 2:
		if len(extra) == 0 {
			log.Fatal("2. start time: wrong format of extra field")
		}

		c.StartTimePlanned, _ = time.Parse(timeLayout, extra)

		logEvent(timeStr, fmt.Sprintf("The start time for the competitor(%d) was set by a draw to %s", competitorID, extra))
	case 3:
		logEvent(timeStr, fmt.Sprintf("The competitor(%d) is on the start line", competitorID))
	case 4:
		c.ActualStartTime = eventTime
		c.CurrentLapStart = eventTime

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) has started", competitorID))
	case 5:
		if len(extra) == 0 {
			log.Fatal("5. firing range: wrong format of extra field")
		}

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) is on the firing range(%s)", competitorID, extra))
	case 6:
		if len(extra) == 0 {
			log.Fatal("6. the target has been hit: wrong format of extra field")
		}

		c.Hits++
		target := extra

		logEvent(timeStr, fmt.Sprintf("The target(%s) has been hit by competitor(%d)", target, competitorID))
	case 7:
		c.Shots += 5

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) left the firing range", competitorID))
	case 8:
		c.PenaltyStart = eventTime

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) entered the penalty laps", competitorID))
	case 9:
		penaltyDuration := eventTime.Sub(c.PenaltyStart)
		c.PenaltyTime += penaltyDuration

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) left the penalty laps", competitorID))
	case 10:
		var lapDuration time.Duration

		if len(c.LapTimes) == 0 {
			lapDuration = eventTime.Sub(c.StartTimePlanned)
		} else {
			lapDuration = eventTime.Sub(c.CurrentLapStart)
		}

		c.LapTimes = append(c.LapTimes, lapDuration)
		c.CurrentLapStart = eventTime

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) ended the main lap", competitorID))

		if len(c.LapTimes) == config.Laps {
			c.Finished = true

			parseEnd, err := time.Parse(timeLayout, timeStr)
			if err != nil {
				log.Fatal(fmt.Errorf("error parsing time: %s", err))
			}

			c.FinishTime = parseEnd

			logEvent(eventTime.Format(timeLayout), fmt.Sprintf("The competitor(%d) has finished", competitorID))
		}
	case 11:
		c.NotFinished = true

		logEvent(timeStr, fmt.Sprintf("The competitor(%d) can`t continue: %s", competitorID, extra))
	default:
		log.Printf("Unhandled event: %d", eventID)
	}
}

func truncateFloat(f float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))

	return math.Floor(f*shift) / shift
}

func generateReport() {
	fmt.Println()

	var results []Result
	for _, c := range competitors {
		status := ""
		var total time.Duration

		if !c.Registered {
			continue
		}

		if c.ActualStartTime.IsZero() {
			status = "[NotStarted]"
		} else if c.NotFinished {
			status = "[NotFinished]"
		} else {
			total = c.FinishTime.Sub(c.StartTimePlanned)
		}

		results = append(results, Result{CompetitorID: c.ID, TotalTime: total, Competitor: c, Status: status})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CompetitorID < results[j].CompetitorID
	})

	for _, r := range results {
		reportString := ""

		if r.Status == "" {
			reportString += fmt.Sprintf("%s %d [", formatDuration(r.TotalTime), r.CompetitorID)
		} else {
			reportString += fmt.Sprintf("%s %d [", r.Status, r.CompetitorID)
		}

		for _, lap := range r.Competitor.LapTimes {
			speed := float64(config.LapLen) / lap.Seconds()
			speedTrunc := truncateFloat(speed, 3)

			reportString += fmt.Sprintf("{%s, %.3f}, ", formatDuration(lap), speedTrunc)
		}

		for i := 0; i < config.Laps-len(r.Competitor.LapTimes); i++ {
			reportString += "{,}, "
		}

		reportString = reportString[:len(reportString)-2]
		reportString += "] "

		speedPenalty := 0.0
		if r.Competitor.PenaltyTime > 0 {
			speedPenalty = float64(config.PenaltyLen) * float64(r.Competitor.Shots-r.Competitor.Hits) / r.Competitor.PenaltyTime.Seconds()
		}

		reportString += fmt.Sprintf("{%s, %.3f} %d/%d\n",
			formatDuration(r.Competitor.PenaltyTime),
			speedPenalty,
			r.Competitor.Hits,
			r.Competitor.Shots,
		)

		fmt.Print(reportString)
	}
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("run app by: go run main.go <config.json> <events.txt>")
	}

	parseConfig(os.Args[1])

	file, err := os.Open(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		processEvent(scanner.Text())
	}

	generateReport()
}
