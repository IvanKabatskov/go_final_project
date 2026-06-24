package api

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

func daysInterval(splitedRepeat []string) (int, error) {
	if len(splitedRepeat) < 2 {
		err := errors.New("invalid 'd' rule format: missing days interval")
		return 0, err
	}
	days, err := strconv.Atoi(splitedRepeat[1])
	if err != nil {
		err = fmt.Errorf("next date: failed to parse a string to a number %q: %w", splitedRepeat[1], err)
		return 0, err
	}
	if days < 1 || days > 400 {
		err = errors.New("invalid day interval: must be between 1 and 400")
		return 0, err
	}
	return days, nil
}

func weeksInterval(splitedRepeat []string) (map[time.Weekday]bool, error) {
	if len(splitedRepeat) < 2 {
		err := errors.New("invalid 'w' rule format: missing weekdays")
		return nil, err
	}
	weekdays := make(map[time.Weekday]bool)
	weekday := strings.Split(splitedRepeat[1], ",")
	for _, day := range weekday {
		numOfDay, err := strconv.Atoi(day)
		if err != nil {
			err = fmt.Errorf("next date: failed to parse a string to a number %q: %w", day, err)
			return nil, err
		}
		if numOfDay < 1 || numOfDay > 7 {
			err = errors.New("invalid day interval: must be between 1 and 7")
			return nil, err
		}
		weekdays[time.Weekday(numOfDay)] = true
	}
	return weekdays, nil
}

func afterNow(nextDate time.Time, now time.Time) bool {
	return nextDate.After(now)
}
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		err = fmt.Errorf("next date: failed to parse date %q: %w", dstart, err)
		return "", err
	}

	if repeat == "" {
		return "", nil
	}
	splitedRepeat := strings.Split(repeat, " ")
	rule := splitedRepeat[0]
	nextDate := date
	switch rule {
	case "y":
		for {
			nextDate = nextDate.AddDate(1, 0, 0)
			if afterNow(nextDate, now) {
				break
			}
		}
	case "d":
		days, err := daysInterval(splitedRepeat)
		if err != nil {
			return "", err
		}
		for {
			nextDate = date.AddDate(0, 0, days)
			if afterNow(nextDate, now) {
				break
			}
		}
	case "w":
		if len(splitedRepeat) < 2 {
			err := errors.New("invalid 'w' rule format: missing week interval")
			return "", err
		}
		weekdays, err := weeksInterval(splitedRepeat)
		if err != nil {
			return "", err
		}
		for {
			nextDate = date.AddDate(0, 0, days)
			if afterNow(nextDate, now) {
				break
			}
		}
	case "m":
	default:
	}
}
