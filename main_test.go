package main

import (
	"reflect"
	"testing"
	"time"
)

func TestFindCommonSlots_ExcludesLunchBreak(t *testing.T) {
	// Setup: 1 participant, no events
	participantIds := []string{"user1"}
	periodStart := parseTime("2025-08-02T09:00:00Z")
	periodEnd := parseTime("2025-08-02T17:00:00Z")
	duration := 60 * time.Minute

	// Clear events for this test
	events = []Event{}

	slots := findCommonSlots(participantIds, periodStart, periodEnd, duration)
	for _, slot := range slots {
		// Lunch break should be excluded
		if slot.Start.Hour() == 13 || slot.End.Hour() == 14 ||
			(slot.Start.Before(time.Date(slot.Start.Year(), slot.Start.Month(), slot.Start.Day(), 14, 0, 0, 0, slot.Start.Location())) &&
				slot.End.After(time.Date(slot.Start.Year(), slot.Start.Month(), slot.Start.Day(), 13, 0, 0, 0, slot.Start.Location()))) {
			t.Errorf("Slot overlaps lunch break: %v - %v", slot.Start, slot.End)
		}
	}
}

func TestScoreAndSelectSlot_PrefersEarlierSlot(t *testing.T) {
	participantIds := []string{"user1"}
	// Two slots: one at 10am, one at 16pm
	slot1 := TimeSlot{parseTime("2025-08-02T10:00:00Z"), parseTime("2025-08-02T11:00:00Z")}
	slot2 := TimeSlot{parseTime("2025-08-02T16:00:00Z"), parseTime("2025-08-02T17:00:00Z")}
	slots := []TimeSlot{slot2, slot1}

	events = []Event{} // No events

	best := scoreAndSelectSlot(slots, participantIds)
	if !reflect.DeepEqual(best, slot1) {
		t.Errorf("Expected earlier slot to be selected, got %v", best)
	}
}

func TestFindCommonSlots_RespectsBusyIntervals(t *testing.T) {
	participantIds := []string{"user1"}
	periodStart := parseTime("2025-08-02T09:00:00Z")
	periodEnd := parseTime("2025-08-02T17:00:00Z")
	duration := 60 * time.Minute

	// Add a busy event from 10:00 to 11:00
	events = []Event{
		{"e1", "Busy", "user1", parseTime("2025-08-02T10:00:00Z"), parseTime("2025-08-02T11:00:00Z")},
	}

	slots := findCommonSlots(participantIds, periodStart, periodEnd, duration)
	for _, slot := range slots {
		if slot.Start.Before(parseTime("2025-08-02T11:00:00Z")) && slot.End.After(parseTime("2025-08-02T10:00:00Z")) {
			t.Errorf("Slot overlaps busy event: %v - %v", slot.Start, slot.End)
		}
	}
}
