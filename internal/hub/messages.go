package hub

import "encoding/json"

// Envelope is the wire format for all WebSocket messages.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func encode(msgType string, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Payload: p})
}

// Client → Server message types
const (
	MsgCreateRoom     = "CREATE_ROOM"
	MsgEnterRoom      = "ENTER_ROOM"
	MsgSelectQuestion = "SELECT_QUESTION"
	MsgOpenBuzzer     = "OPEN_BUZZER"
	MsgSkipQuestion   = "SKIP_QUESTION"
	MsgBuzzIn         = "BUZZ_IN"
	MsgSubmitAnswer   = "SUBMIT_ANSWER"
	MsgJudgeAnswer    = "JUDGE_ANSWER"
	MsgStartGame      = "START_GAME"
)

// Server → Client message types
const (
	MsgRoomStateUpdate = "ROOM_STATE_UPDATE"
	MsgPlayerBuzzed    = "PLAYER_BUZZED"
	MsgAnswerResult    = "ANSWER_RESULT"
	MsgAnswerTimeout   = "ANSWER_TIMEOUT"
	MsgError           = "ERROR"
)

// Internal message types (timer callbacks → room goroutine, never sent over WS)
const (
	MsgInternalInit         = "_INIT"
	MsgInternalClientLeft   = "_CLIENT_LEFT"
	MsgInternalBuzzResolve  = "_BUZZ_RESOLVE"
	MsgInternalAnswerExpire = "_ANSWER_EXPIRE"
)

// Payload structs — Client → Server

type CreateRoomPayload struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	PackID    string `json:"pack_id"`
	Lang      string `json:"lang"`
}

type EnterRoomPayload struct {
	SessionID string `json:"session_id"`
	RoomID    string `json:"room_id"`
	Name      string `json:"name"` // required for new players, optional on reconnect
}

type SelectQuestionPayload struct {
	CategoryIndex int `json:"category_index"`
	QuestionIndex int `json:"question_index"`
}

type JudgeAnswerPayload struct {
	Correct bool `json:"correct"`
}

type SubmitAnswerPayload struct {
	Text string `json:"text"`
}

// Payload structs — Server → Client

type ErrorPayload struct {
	Message string `json:"message"`
}

type PlayerBuzzedPayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

type AnswerResultPayload struct {
	Correct    bool   `json:"correct"`
	PlayerID   string `json:"player_id"`
	ScoreDelta int    `json:"score_delta"`
}
