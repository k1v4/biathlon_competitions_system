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
		log.Fatalf("error parsing config file: %s", err)
	}
}

func logEvent(timeStr string, text string) {
	log.Printf("[%s] %s\n", timeStr, text)
}

func processEvent(event string) {
	re := regexp.MustCompile(`\[(.*?)\] (\d+) (\d+)(?: (.*))?`)
	matches := re.FindStringSubmatch(event)

	if len(matches) < 4 {
		log.Fatal("bad format for event")

		return
	}

	timeStr, eventID, competitorID, eventTime, extra, err := parseParams(matches)
	if err != nil {
		log.Fatalf("failed to parse event: %s", err)
	}

	c, ok := competitors[competitorID]
	if !ok {
		c = &Competitor{ID: competitorID}
		competitors[competitorID] = c
	}

	if c.Disqualified {
		logEvent(timeStr, fmt.Sprintf("The competitor(%d) is disqualified", competitorID))

		return
	}

	switch eventID {
	case 1:
		handleRegistration(c, timeStr)
	case 2:
		handlePlanningStartTime(c, timeStr, extra)
	case 3:
		logEvent(timeStr, fmt.Sprintf("The competitor(%d) is on the start line", competitorID))
	case 4:
		handleStartTime(c, timeStr, eventTime)
	case 5:
		handleFiringRange(c, timeStr, extra)
	case 6:
		handleShooting(c, timeStr, extra)
	case 7:
		handleLeftFiringRange(c, timeStr)
	case 8:
		handleEnteringPenalty(c, timeStr, eventTime)
	case 9:
		handleLeftPenalty(c, timeStr, eventTime)
	case 10:
		handleEndMainLap(c, timeStr, eventTime)
	case 11:
		handleCantContinue(c, timeStr, extra)
	default:
		log.Printf("Unhandled event: %d", eventID)
	}
}

func parseParams(matches []string) (string, int, int, time.Time, string, error) {
	timeStr, eventIDStr, competitorIDStr := matches[1], matches[2], matches[3]

	extra := ""
	if len(matches) > 4 {
		extra = matches[4]
	}

	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		return "", 0, 0, time.Time{}, "", fmt.Errorf("cant parse event id: %s", eventIDStr)
	}

	competitorID, err := strconv.Atoi(competitorIDStr)
	if err != nil {
		return "", 0, 0, time.Time{}, "", fmt.Errorf("cant parse competitor id: %s", competitorIDStr)
	}

	eventTime, err := time.Parse(timeLayout, timeStr)
	if err != nil {
		return "", 0, 0, time.Time{}, "", fmt.Errorf("cant parse event time: %s", timeStr)
	}

	return timeStr, eventID, competitorID, eventTime, extra, nil
}

func truncateFloat(f float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))

	return math.Floor(f*shift) / shift
}

func generateReport() {
	fmt.Println()
	log.Println("Final report")

	var results []Result
	for _, c := range competitors {
		status := ""
		var total time.Duration

		if !c.Registered {
			continue
		}

		if c.ActualStartTime.IsZero() || c.Disqualified {
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

		log.Print(reportString)
	}
}

func main() {
	log.SetFlags(0)

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
