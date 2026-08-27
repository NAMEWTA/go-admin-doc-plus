package scheduler

import (
	"encoding/json"
	"sort"
	"time"
)

type Schedule struct {
	Minutes     []int `json:"minutes"`
	Hours       []int `json:"hours"`
	DaysOfMonth []int `json:"daysOfMonth"`
	Months      []int `json:"months"`
	Weekdays    []int `json:"weekdays"`
}

func normalizeSchedule(value Schedule) (Schedule, bool) {
	var ok bool
	if value.Minutes, ok = normalizedSet(value.Minutes, 0, 59, true); !ok {
		return Schedule{}, false
	}
	if value.Hours, ok = normalizedSet(value.Hours, 0, 23, true); !ok {
		return Schedule{}, false
	}
	if value.DaysOfMonth, ok = normalizedSet(value.DaysOfMonth, 1, 31, false); !ok {
		return Schedule{}, false
	}
	if value.Months, ok = normalizedSet(value.Months, 1, 12, true); !ok {
		return Schedule{}, false
	}
	if value.Weekdays, ok = normalizedSet(value.Weekdays, 0, 6, false); !ok {
		return Schedule{}, false
	}
	if len(value.DaysOfMonth) != 0 && len(value.Weekdays) != 0 {
		return Schedule{}, false
	}
	return value, true
}

func normalizedSet(values []int, minimum, maximum int, required bool) ([]int, bool) {
	if required && len(values) == 0 || len(values) > maximum-minimum+1 {
		return nil, false
	}
	result := append([]int{}, values...)
	sort.Ints(result)
	for index, value := range result {
		if value < minimum || value > maximum || index > 0 && result[index-1] == value {
			return nil, false
		}
	}
	return result, true
}

func nextOccurrence(value Schedule, after time.Time) (time.Time, bool) {
	normalized, ok := normalizeSchedule(value)
	if !ok || after.IsZero() || after.Location() != time.UTC {
		return time.Time{}, false
	}
	limit := after.AddDate(5, 0, 0)
	for candidate := after.Truncate(time.Minute).Add(time.Minute); !candidate.After(limit); candidate = candidate.Add(time.Minute) {
		if containsInteger(normalized.Months, int(candidate.Month())) && containsInteger(normalized.Hours, candidate.Hour()) && containsInteger(normalized.Minutes, candidate.Minute()) && matchesDay(normalized, candidate) {
			return candidate, true
		}
	}
	return time.Time{}, false
}

func matchesDay(value Schedule, candidate time.Time) bool {
	if len(value.DaysOfMonth) > 0 {
		return containsInteger(value.DaysOfMonth, candidate.Day())
	}
	if len(value.Weekdays) > 0 {
		return containsInteger(value.Weekdays, int(candidate.Weekday()))
	}
	return true
}

func containsInteger(values []int, target int) bool {
	index := sort.SearchInts(values, target)
	return index < len(values) && values[index] == target
}

func marshalSchedule(value Schedule) ([]byte, error) { return json.Marshal(value) }
func unmarshalSchedule(value []byte) (Schedule, error) {
	var result Schedule
	if err := json.Unmarshal(value, &result); err != nil {
		return Schedule{}, err
	}
	normalized, ok := normalizeSchedule(result)
	if !ok {
		return Schedule{}, ErrInternal
	}
	return normalized, nil
}
