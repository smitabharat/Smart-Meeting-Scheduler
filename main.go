package main

import (
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// --------- Models --------

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Event struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	UserID    string    `json:"userId"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type MeetingRequest struct {
	ParticipantIds  []string `json:"participantIds"`
	DurationMinutes int      `json:"durationMinutes"`
	TimeRange       struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"timeRange"`
}

type MeetingResponse struct {
	MeetingId      string   `json:"meetingId"`
	Title          string   `json:"title"`
	ParticipantIds []string `json:"participantIds"`
	StartTime      string   `json:"startTime"`
	EndTime        string   `json:"endTime"`
}

// --------- In-memory DB ---------

var users = []User{
	{"u1", "Revti"},
	{"u2", "Sejal"},
	{"u3", "Tanvi"},
}
var events = []Event{
	{"e1", "Existing Meeting", "u1", parseTime("2025-08-02T14:00:00Z"), parseTime("2025-08-02T15:00:00Z")},
	{"e2", "Existing Meeting", "u2", parseTime("2025-08-03T09:00:00Z"), parseTime("2025-08-03T10:00:00Z")},
	{"e3", "Existing Meeting", "u3", parseTime("2025-08-04T13:00:00Z"), parseTime("2025-08-04T14:00:00Z")},
}

// --------- Utils ---------

func parseTime(str string) time.Time {
	t, _ := time.Parse(time.RFC3339, str)
	return t
}
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// --------- Main ---------

func main() {
	app := fiber.New()

	app.Post("/schedule", scheduleMeeting)
	app.Get("/users/:userId/calendar", getUserCalendar)

	app.Listen(":8080")
}

// --------- Handlers ---------

func getUserCalendar(c *fiber.Ctx) error {
	userId := c.Params("userId") // ✅ Get userId from URL path
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId is required",
		})
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "start and end query params are required",
		})
	}

	layout := time.RFC3339 // same as "2006-01-02T15:04:05Z07:00"

	start, err := time.Parse(layout, strings.TrimSpace(startStr))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start date format",
		})
	}

	end, err := time.Parse(layout, strings.TrimSpace(endStr))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end date format",
		})
	}

	// Example response: return parsed values
	return c.JSON(fiber.Map{
		"userId":  userId,
		"start":   start,
		"end":     end,
		"message": "Calendar fetched successfully",
	})
}

func scheduleMeeting(c *fiber.Ctx) error {
	var req MeetingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	start := parseTime(req.TimeRange.Start)
	end := parseTime(req.TimeRange.End)
	duration := time.Duration(req.DurationMinutes) * time.Minute

	slots := findCommonSlots(req.ParticipantIds, start, end, duration)
	if len(slots) == 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "No available time slot found for all participants."})
	}

	bestSlot := scoreAndSelectSlot(slots, req.ParticipantIds)
	meetingId := uuid.New().String()

	for _, userId := range req.ParticipantIds {
		events = append(events, Event{
			ID:        uuid.New().String(),
			Title:     "New Meeting",
			UserID:    userId,
			StartTime: bestSlot.Start,
			EndTime:   bestSlot.End,
		})
	}

	resp := MeetingResponse{
		MeetingId:      meetingId,
		Title:          "New Meeting",
		ParticipantIds: req.ParticipantIds,
		StartTime:      bestSlot.Start.Format(time.RFC3339),
		EndTime:        bestSlot.End.Format(time.RFC3339),
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// --------- Slot Discovery and Scoring ---------

type TimeSlot struct {
	Start time.Time
	End   time.Time
}

func findCommonSlots(participantIds []string, periodStart, periodEnd time.Time, duration time.Duration) []TimeSlot {
	busyIntervals := map[string][]TimeSlot{}
	for _, id := range participantIds {
		bi := []TimeSlot{}
		for _, ev := range events {
			if ev.UserID == id {
				bi = append(bi, TimeSlot{ev.StartTime, ev.EndTime})
			}
		}
		busyIntervals[id] = bi
	}

	slots := []TimeSlot{}
	date := periodStart
	for date.Add(duration).Before(periodEnd) || date.Add(duration).Equal(periodEnd) {
		possible := true
		slot := TimeSlot{date, date.Add(duration)}
		for _, id := range participantIds {
			for _, bi := range busyIntervals[id] {
				if slot.Start.Before(bi.End) && slot.End.After(bi.Start) {
					possible = false
					break
				}
			}
			if !possible {
				break
			}
		}
		if possible {
			slots = append(slots, slot)
		}
		date = date.Add(15 * time.Minute)
	}
	return slots
}

func scoreAndSelectSlot(slots []TimeSlot, participantIds []string) TimeSlot {
	type ScoredSlot struct {
		Slot  TimeSlot
		Score int
	}
	var scored []ScoredSlot
	for _, slot := range slots {
		score := 0
		hour := slot.Start.Hour()

		// Prefer earlier slots
		score += hour * 2

		// Working hours boost
		if hour < 9 || hour > 17 {
			score += 20
		}

		// Buffer + gap penalties
		for _, id := range participantIds {
			for _, e := range events {
				if e.UserID == id {
					// Buffer penalty (<15 min)
					if abs(int(slot.Start.Sub(e.EndTime).Minutes())) < 15 ||
						abs(int(e.StartTime.Sub(slot.End).Minutes())) < 15 {
						score += 10
					}
					// Awkward gap penalty (0 < gap < 30)
					gap := int(slot.Start.Sub(e.EndTime).Minutes())
					if gap > 0 && gap < 30 {
						score += 5
					}
				}
			}
		}

		scored = append(scored, ScoredSlot{slot, score})
	}

	sort.Slice(scored, func(i, j int) bool { return scored[i].Score < scored[j].Score })
	return scored[0].Slot
}
