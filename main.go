package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Question struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"` // "choice", "true_false"
	Text      string   `json:"text"`
	Options   []string `json:"options,omitempty"`
	Answer    string   `json:"-"`
	Timestamp int64    `json:"timestamp"`
}

type Answer struct {
	QuestionID string `json:"question_id"`
	StudentID  string `json:"student_id"`
	Answer     string `json:"answer"`
}

type Statistics struct {
	QuestionID string         `json:"question_id"`
	Total      int            `json:"total"`
	Answers    map[string]int `json:"answers"`
}

var (
	questions   = make(map[string]Question)
	answers     = make(map[string][]Answer)
	statistics  = make(map[string]Statistics)
	questionsMu sync.RWMutex
	answersMu   sync.RWMutex
)

func createQuestion(w http.ResponseWriter, r *http.Request) {
	var q Question
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	q.ID = fmt.Sprintf("q%d", time.Now().UnixNano())
	q.Timestamp = time.Now().Unix()

	questionsMu.Lock()
	questions[q.ID] = q
	questionsMu.Unlock()

	// 初始化统计数据
	stats := Statistics{
		QuestionID: q.ID,
		Total:      0,
		Answers:    make(map[string]int),
	}

	// 如果是选择题，初始化选项计数
	if q.Type == "choice" {
		for _, option := range q.Options {
			stats.Answers[option] = 0
		}
	} else if q.Type == "true_false" {
		stats.Answers["true"] = 0
		stats.Answers["false"] = 0
	}

	answersMu.Lock()
	statistics[q.ID] = stats
	answersMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func submitAnswer(w http.ResponseWriter, r *http.Request) {
	var a Answer
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	answersMu.Lock()
	defer answersMu.Unlock()

	// 添加到答案列表
	answers[a.QuestionID] = append(answers[a.QuestionID], a)

	// 更新统计
	if stats, ok := statistics[a.QuestionID]; ok {
		stats.Total++
		stats.Answers[a.Answer]++
		statistics[a.QuestionID] = stats
	}

	w.WriteHeader(http.StatusCreated)
}

func getStatistics(w http.ResponseWriter, r *http.Request) {
	questionID := r.URL.Query().Get("question_id")

	answersMu.RLock()
	stats, ok := statistics[questionID]
	answersMu.RUnlock()

	if !ok {
		http.Error(w, "Question not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	http.HandleFunc("/question", createQuestion)
	http.HandleFunc("/answer", submitAnswer)
	http.HandleFunc("/statistics", getStatistics)

	fmt.Println("Classroom service started on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
