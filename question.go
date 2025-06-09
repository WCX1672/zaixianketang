package interact

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"redisutil"
)

type QuestionHandler struct {
	redis *redisutil.RedisClient
}

func NewQuestionHandler(redis *redisutil.RedisClient) *QuestionHandler {
	return &QuestionHandler{redis: redis}
}

func (h *QuestionHandler) PushQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var question Question
	if err := json.NewDecoder(r.Body).Decode(&question); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	question.ID = strconv.FormatInt(time.Now().UnixNano(), 10)
	questionKey := "question:" + question.RoomID + ":" + question.ID

	// 保存题目到Redis
	if err := h.redis.SetJSON(questionKey, question, 24*time.Hour); err != nil {
		http.Error(w, "Failed to save question", http.StatusInternalServerError)
		return
	}

	// 初始化统计
	statsKey := "stats:" + question.RoomID + ":" + question.ID
	for _, option := range question.Options {
		h.redis.HSet(statsKey, option.ID, "0")
	}
	h.redis.Expire(statsKey, 24*time.Hour)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": question.ID})
}

func (h *QuestionHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var answer Answer
	if err := json.NewDecoder(r.Body).Decode(&answer); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statsKey := "stats:" + answer.RoomID + ":" + answer.QuestionID

	// 原子增加答案计数
	if _, err := h.redis.HIncrBy(statsKey, answer.OptionID, 1); err != nil {
		http.Error(w, "Failed to record answer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *QuestionHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	questionID := r.URL.Query().Get("question")
	statsKey := "stats:" + roomID + ":" + questionID

	stats, err := h.redis.HGetAll(statsKey)
	if err != nil {
		http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

type Question struct {
	ID       string   `json:"id"`
	RoomID   string   `json:"room_id"`
	Text     string   `json:"text"`
	Type     string   `json:"type"` // "choice", "truefalse"
	Options  []Option `json:"options"`
	Duration int      `json:"duration"` // 答题时间(秒)
}

type Option struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

type Answer struct {
	QuestionID string `json:"question_id"`
	RoomID     string `json:"room_id"`
	UserID     string `json:"user_id"`
	OptionID   string `json:"option_id"`
}
