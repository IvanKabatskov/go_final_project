package api

import (
	"errors"
	"fmt"
	"net/http"
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
	weekDays := make(map[time.Weekday]bool)
	weekDay := strings.Split(splitedRepeat[1], ",")
	for _, day := range weekDay {
		numOfDay, err := strconv.Atoi(day)
		if err != nil {
			err = fmt.Errorf("next date: failed to parse a string to a number %q: %w", day, err)
			return nil, err
		}
		if numOfDay < 1 || numOfDay > 7 {
			err = errors.New("invalid day interval: must be between 1 and 7")
			return nil, err
		}
		if numOfDay == 7 {
			numOfDay = 0
		}
		weekDays[time.Weekday(numOfDay)] = true
	}
	return weekDays, nil
}

func monthsInterval(splitedRepeat []string) (map[int]bool, map[int]bool, error) {
	if len(splitedRepeat) < 2 {
		err := errors.New("invalid 'm' rule format: missing days of month")
		return nil, nil, err
	}
	months := make(map[int]bool)
	dayOfMonth := make(map[int]bool)

	daysOfMoth := strings.Split(splitedRepeat[1], ",")
	for _, day := range daysOfMoth {
		numOfDay, err := strconv.Atoi(day)
		if err != nil {
			err = fmt.Errorf("next date: failed to parse a string to a number %q: %w", day, err)
			return nil, nil, err
		}
		if numOfDay < -2 || numOfDay > 31 || numOfDay == 0 {
			err = errors.New("invalid day interval: must be between -2 and 31, not include 0")
			return nil, nil, err
		}
		dayOfMonth[numOfDay] = true
	}
	if len(splitedRepeat) > 2 {
		spletedMonths := strings.Split(splitedRepeat[2], ",")
		for _, month := range spletedMonths {
			numOfMonth, err := strconv.Atoi(month)
			if err != nil {
				err = fmt.Errorf("next date: failed to parse a string to a number %q: %w", month, err)
				return nil, nil, err
			}
			if numOfMonth < 1 || numOfMonth > 12 {
				err = errors.New("invalid month interval: must be between 1 and 12")
				return nil, nil, err
			}
			months[numOfMonth] = true
		}
	}
	return dayOfMonth, months, nil
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
			nextDate = nextDate.AddDate(0, 0, days)
			if afterNow(nextDate, now) {
				break
			}
		}
	case "w":
		weekdays, err := weeksInterval(splitedRepeat)
		if err != nil {
			return "", err
		}
		for {
			nextDate = nextDate.AddDate(0, 0, 1)
			if weekdays[nextDate.Weekday()] && afterNow(nextDate, now) {
				break
			}
		}
	case "m":
		dayOfMonth, months, err := monthsInterval(splitedRepeat)
		if err != nil {
			return "", err
		}
		for {
			nextDate = nextDate.AddDate(0, 0, 1)
			endOfMonth := (nextDate.AddDate(0, 1, -nextDate.Day())).Day()
			if len(months) == 0 {
				dayMatches := dayOfMonth[nextDate.Day()] ||
					(dayOfMonth[-1] && nextDate.Day() == endOfMonth) ||
					(dayOfMonth[-2] && nextDate.Day() == endOfMonth-1)
				if dayMatches && afterNow(nextDate, now) {
					break
				}
			} else {
				dayMatches := months[int(nextDate.Month())] && (dayOfMonth[nextDate.Day()] ||
					(dayOfMonth[-1] && nextDate.Day() == endOfMonth) ||
					(dayOfMonth[-2] && nextDate.Day() == endOfMonth-1))
				if dayMatches && afterNow(nextDate, now) {
					break
				}
			}
		}

	default:
		err := fmt.Errorf("unknown repeat rule: %s", rule)
		return "", err
	}
	return nextDate.Format(DateFormat), nil
}

func NextDayHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	nowValue := req.FormValue("now")
	dstart := req.FormValue("date")
	repeat := req.FormValue("repeat")

	var now time.Time
	var err error
	if nowValue == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(DateFormat, nowValue)
		if err != nil {
			http.Error(res, "invalid 'now' date format", http.StatusBadRequest)
			return
		}
	}
	nextDate, err := NextDate(now, dstart, repeat)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.Write([]byte(nextDate))
}
