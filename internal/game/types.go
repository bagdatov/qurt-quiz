package game

import (
	"fmt"
	"time"

	"github.com/qurt-quiz/qurt-quiz/internal/pack"
)

type Phase string

const (
	PhaseLobby     Phase = "LOBBY"
	PhaseBoard     Phase = "BOARD"
	PhaseQuestion  Phase = "QUESTION"
	PhaseBuzzer    Phase = "BUZZER"
	PhaseAnswering Phase = "ANSWERING"
	PhaseJudging   Phase = "JUDGING"
	PhaseGameOver  Phase = "GAME_OVER"
)

const MaxPlayers = 5

type Player struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Score     int    `json:"score"`
	IsHost    bool   `json:"is_host"`
	Connected bool   `json:"connected"`
}

type ActiveQuestion struct {
	CategoryIndex int    `json:"category_index"`
	QuestionIndex int    `json:"question_index"`
	Text          string `json:"text"`
	Value         int    `json:"value"`
}

type Room struct {
	ID      string             `json:"id"`
	PackID  string             `json:"pack_id"`
	Lang    string             `json:"lang"`
	Players map[string]*Player `json:"players"`
	State   GameState          `json:"state"`
}

type GameState struct {
	Phase           Phase           `json:"phase"`
	Pack            *pack.QuizPack  `json:"pack,omitempty"`
	UsedCells       map[string]bool `json:"used_cells"`
	ActiveQuestion  *ActiveQuestion `json:"active_question,omitempty"`
	ActiveChooserID string          `json:"active_chooser_id,omitempty"` // player who picks the next question
	BuzzWinnerID    string          `json:"buzz_winner_id,omitempty"`
	AnswerDeadline  time.Time       `json:"answer_deadline,omitempty"`
	AnswerDuration  time.Duration   `json:"answer_duration"`
	SubmittedAnswer string          `json:"submitted_answer,omitempty"`
}

func CellKey(catIdx, qIdx int) string {
	return fmt.Sprintf("%d:%d", catIdx, qIdx)
}

func (r *Room) HostID() string {
	for id, p := range r.Players {
		if p.IsHost {
			return id
		}
	}
	return ""
}

// PlayerCount returns the number of non-host players.
func (r *Room) PlayerCount() int {
	n := 0
	for _, p := range r.Players {
		if !p.IsHost {
			n++
		}
	}
	return n
}

func (r *Room) AllCellsUsed() bool {
	if r.State.Pack == nil {
		return false
	}
	for catIdx, cat := range r.State.Pack.Categories {
		for qIdx := range cat.Questions {
			if !r.State.UsedCells[CellKey(catIdx, qIdx)] {
				return false
			}
		}
	}
	return true
}

// PromoteNextHost picks the first connected non-host player and promotes them.
// Returns false if no eligible player exists.
func (r *Room) PromoteNextHost() bool {
	for _, p := range r.Players {
		if !p.IsHost && p.Connected {
			p.IsHost = true
			return true
		}
	}
	return false
}
