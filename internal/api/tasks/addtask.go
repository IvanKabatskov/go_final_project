package tasks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"final_project/internal/api"
	"final_project/internal/db"
)

func writeJSON(res http.ResponseWriter, statusCode int, data any) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.WriteHeader(statusCode)
	json.NewEncoder(res).Encode(data)
}

func AddTaskHandler(res http.ResponseWriter, req *http.Request) {

	var task db.Task

	err := json.NewDecoder(req.Body).Decode(&task)
	if err != nil {
		writeJSON(res, http.StatusBadRequest, map[string]string{"error": "Ошибка десериализации JSON"})
		return
	}
	if task.Title == "" {
		writeJSON(res, http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	now := time.Now()

	if task.Date == "" {
		task.Date = now.Format(api.DateFormat)
	}

	parsedDate, err := time.Parse(api.DateFormat, task.Date)
	if err != nil {
		writeJSON(res, http.StatusBadRequest, map[string]string{"error": "Некорректный формат даты"})
		return
	}
	if task.Repeat != "" {
		next, err := api.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJSON(res, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if api.AfterNow(now, parsedDate) {
			task.Date = next
		}
	} else {
		if api.AfterNow(now, parsedDate) {
			task.Date = now.Format(api.DateFormat)
		}
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSON(res, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(res, http.StatusOK, map[string]string{"id": strconv.FormatInt(id, 10)})
}
