package hub

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/qurt-quiz/qurt-quiz/internal/game"
	"github.com/qurt-quiz/qurt-quiz/internal/store"
)

const buzzWindow = 50 * time.Millisecond

// internalMsg is sent by timers or the hub to the room's incoming channel.
type internalMsg struct {
	typ      string
	timerGen int // used to detect stale timer callbacks
}

// roomMsg bundles a client with the envelope it sent (or an internal event).
type roomMsg struct {
	client   *Client
	env      Envelope
	internal *internalMsg
}

// Room is the actor that owns all mutable game state for one session.
// All mutations are serialized through the incoming channel — no locks on game state.
type Room struct {
	game    *game.Room
	clients map[string]*Client
	store   store.RoomStore
	hub     *Hub

	incoming chan roomMsg
	quit     chan struct{}
	closed   bool // guards against double close(r.quit); only written inside run()

	buzzCandidates  []string
	buzzWindowOpen  bool // true once the first BUZZ_IN starts the 50ms collection window
	buzzTimer       *time.Timer

	answerTimer    *time.Timer
	answerTimerGen int // incremented each time a new answer timer starts
}

func newRoom(gr *game.Room, s store.RoomStore, h *Hub) *Room {
	return &Room{
		game:     gr,
		clients:  make(map[string]*Client),
		store:    s,
		hub:      h,
		incoming: make(chan roomMsg, 128),
		quit:     make(chan struct{}),
	}
}

func (r *Room) run() {
	defer func() {
		r.hub.removeRoom(r.game.ID)
		if err := r.store.Delete(r.game.ID); err != nil {
			log.Printf("room %s store delete: %v", r.game.ID, err)
		}
	}()

	for {
		select {
		case msg := <-r.incoming:
			if msg.internal != nil {
				r.handleInternal(msg)
			} else {
				r.handleMsg(msg.client, msg.env)
			}
		case <-r.quit:
			return
		}
	}
}

func (r *Room) send(msg roomMsg) {
	select {
	case r.incoming <- msg:
	case <-r.quit:
	}
}

// handleInternal processes hub-injected and timer-fired events.
func (r *Room) handleInternal(msg roomMsg) {
	switch msg.internal.typ {
	case MsgInternalInit:
		r.broadcastState()
	case MsgInternalClientLeft:
		r.clientLeft(msg.client)
	case MsgEnterRoom:
		r.handleEnterRoom(msg.client, msg.env.Payload)
	case MsgInternalBuzzResolve:
		r.resolveBuzz()
	case MsgInternalAnswerExpire:
		// Ignore stale timer callbacks that fired for a previous question.
		if msg.internal.timerGen == r.answerTimerGen {
			r.expireAnswer()
		}
	}
}

// handleMsg processes a WebSocket message from a client inside the room.
func (r *Room) handleMsg(c *Client, env Envelope) {
	switch env.Type {
	case MsgEnterRoom:
		// Clients with an existing roomID reaching us via routeToRoom.
		r.handleEnterRoom(c, env.Payload)
	case MsgStartGame:
		r.handleStartGame(c)
	case MsgSelectQuestion:
		r.handleSelectQuestion(c, env.Payload)
	case MsgOpenBuzzer:
		r.handleOpenBuzzer(c)
	case MsgSkipQuestion:
		r.handleSkipQuestion(c)
	case MsgBuzzIn:
		r.handleBuzzIn(c)
	case MsgSubmitAnswer:
		r.handleSubmitAnswer(c, env.Payload)
	case MsgJudgeAnswer:
		r.handleJudgeAnswer(c, env.Payload)
	default:
		c.sendError("unknown message type: " + env.Type)
	}
}

// --- Join / reconnect ---

func (r *Room) handleEnterRoom(c *Client, raw json.RawMessage) {
	var p EnterRoomPayload
	if err := json.Unmarshal(raw, &p); err != nil || p.SessionID == "" {
		c.sendError("ENTER_ROOM requires session_id and room_id")
		return
	}

	existing, isReconnect := r.game.Players[p.SessionID]
	if isReconnect {
		existing.Connected = true
		c.setIdentity(p.SessionID, r.game.ID)
		r.clients[p.SessionID] = c
		r.broadcastState()
		return
	}

	// New player joining.
	if r.game.State.Phase != game.PhaseLobby {
		c.sendError("game already in progress")
		return
	}
	if r.game.PlayerCount() >= game.MaxPlayers {
		c.sendError("room is full (max 5 players)")
		return
	}
	if p.Name == "" {
		c.sendError("name is required to join")
		return
	}

	r.game.Players[p.SessionID] = &game.Player{
		ID:        p.SessionID,
		Name:      p.Name,
		Connected: true,
	}
	c.setIdentity(p.SessionID, r.game.ID)
	r.clients[p.SessionID] = c
	r.broadcastState()
}

// --- Game action handlers ---

func (r *Room) handleStartGame(c *Client) {
	if !r.isHost(c) {
		c.sendError("only the host can start the game")
		return
	}
	if r.game.State.Phase != game.PhaseLobby {
		c.sendError("game already started")
		return
	}
	r.game.State.Phase = game.PhaseBoard
	r.game.State.ActiveChooserID = r.randomPlayerID()
	r.broadcastState()
}

func (r *Room) handleSelectQuestion(c *Client, raw json.RawMessage) {
	if r.game.State.Phase != game.PhaseBoard {
		c.sendError("not in board phase")
		return
	}
	if !r.canChoose(c) {
		c.sendError("it's not your turn to pick a question")
		return
	}

	var p SelectQuestionPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		c.sendError("invalid payload")
		return
	}

	pk := r.game.State.Pack
	if p.CategoryIndex < 0 || p.CategoryIndex >= len(pk.Categories) {
		c.sendError("invalid category index")
		return
	}
	cat := pk.Categories[p.CategoryIndex]
	if p.QuestionIndex < 0 || p.QuestionIndex >= len(cat.Questions) {
		c.sendError("invalid question index")
		return
	}

	key := game.CellKey(p.CategoryIndex, p.QuestionIndex)
	if r.game.State.UsedCells[key] {
		c.sendError("question already used")
		return
	}

	q := cat.Questions[p.QuestionIndex]
	r.game.State.ActiveQuestion = &game.ActiveQuestion{
		CategoryIndex: p.CategoryIndex,
		QuestionIndex: p.QuestionIndex,
		Text:          q.Text,
		Value:         q.Value,
	}
	r.game.State.Phase = game.PhaseQuestion
	r.broadcastState()
}

func (r *Room) handleOpenBuzzer(c *Client) {
	if !r.isHost(c) {
		c.sendError("only the host can open the buzzer")
		return
	}
	if r.game.State.Phase != game.PhaseQuestion {
		c.sendError("not in question phase")
		return
	}

	r.game.State.Phase = game.PhaseBuzzer
	r.buzzCandidates = nil
	r.buzzWindowOpen = false
	// No timer yet — the collection window starts only when the first player buzzes in.
	r.broadcastState()
}

func (r *Room) handleBuzzIn(c *Client) {
	if r.game.State.Phase != game.PhaseBuzzer {
		return
	}
	p := r.game.Players[c.getSessionID()]
	if p == nil || p.IsHost {
		return
	}
	r.buzzCandidates = append(r.buzzCandidates, c.getSessionID())

	// Start the 50ms collection window on the first buzz so that near-simultaneous
	// presses are all gathered before picking a random winner.
	if !r.buzzWindowOpen {
		r.buzzWindowOpen = true
		r.buzzTimer = time.AfterFunc(buzzWindow, func() {
			r.send(roomMsg{internal: &internalMsg{typ: MsgInternalBuzzResolve}})
		})
	}
}

func (r *Room) handleSkipQuestion(c *Client) {
	if !r.isHost(c) {
		c.sendError("only the host can skip a question")
		return
	}
	if r.game.State.Phase != game.PhaseBuzzer && r.game.State.Phase != game.PhaseQuestion {
		c.sendError("can only skip during question or buzzer phase")
		return
	}
	r.stopBuzzTimer()
	r.buzzWindowOpen = false
	// Mark used with no score change so the cell is greyed out on the board.
	r.concludeQuestion(0, "")
}

func (r *Room) resolveBuzz() {
	if r.game.State.Phase != game.PhaseBuzzer || len(r.buzzCandidates) == 0 {
		return
	}

	winnerID := r.buzzCandidates[rand.Intn(len(r.buzzCandidates))]
	r.game.State.BuzzWinnerID = winnerID
	r.game.State.Phase = game.PhaseAnswering
	r.game.State.AnswerDeadline = time.Now().Add(r.game.State.AnswerDuration)

	r.answerTimerGen++
	gen := r.answerTimerGen
	r.answerTimer = time.AfterFunc(r.game.State.AnswerDuration, func() {
		r.send(roomMsg{internal: &internalMsg{typ: MsgInternalAnswerExpire, timerGen: gen}})
	})

	winner := r.game.Players[winnerID]
	r.broadcast(MsgPlayerBuzzed, PlayerBuzzedPayload{
		PlayerID:   winnerID,
		PlayerName: winner.Name,
	})
	r.broadcastState()
}

func (r *Room) handleSubmitAnswer(c *Client, raw json.RawMessage) {
	if r.game.State.Phase != game.PhaseAnswering {
		return
	}
	if c.getSessionID() != r.game.State.BuzzWinnerID {
		c.sendError("not your turn to answer")
		return
	}

	var p SubmitAnswerPayload
	if err := json.Unmarshal(raw, &p); err != nil || p.Text == "" {
		c.sendError("answer text is required")
		return
	}

	r.stopAnswerTimer()
	r.game.State.SubmittedAnswer = p.Text
	r.game.State.Phase = game.PhaseJudging
	r.broadcastState()
}

func (r *Room) handleJudgeAnswer(c *Client, raw json.RawMessage) {
	if !r.isHost(c) {
		c.sendError("only the host can judge answers")
		return
	}
	if r.game.State.Phase != game.PhaseJudging {
		c.sendError("not in judging phase")
		return
	}

	var p JudgeAnswerPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		c.sendError("invalid payload")
		return
	}

	aq := r.game.State.ActiveQuestion
	delta := aq.Value
	if !p.Correct {
		delta = -aq.Value
	}

	winnerID := r.game.State.BuzzWinnerID
	r.broadcast(MsgAnswerResult, AnswerResultPayload{
		Correct:    p.Correct,
		PlayerID:   winnerID,
		ScoreDelta: delta,
	})

	// Correct answer: winner earns the right to pick the next question.
	if p.Correct {
		r.game.State.ActiveChooserID = winnerID
	}
	r.concludeQuestion(delta, winnerID)
}

func (r *Room) expireAnswer() {
	if r.game.State.Phase != game.PhaseAnswering {
		return
	}

	winnerID := r.game.State.BuzzWinnerID
	aq := r.game.State.ActiveQuestion

	r.broadcast(MsgAnswerTimeout, AnswerResultPayload{
		Correct:    false,
		PlayerID:   winnerID,
		ScoreDelta: -aq.Value,
	})

	r.concludeQuestion(-aq.Value, winnerID)
}

// concludeQuestion applies the score delta, cleans up active question state, and advances the phase.
// winnerID may be empty (when nobody buzzed).
func (r *Room) concludeQuestion(delta int, winnerID string) {
	if winnerID != "" {
		if p := r.game.Players[winnerID]; p != nil {
			p.Score += delta
		}
	}
	if r.game.State.ActiveQuestion != nil {
		r.game.State.UsedCells[game.CellKey(
			r.game.State.ActiveQuestion.CategoryIndex,
			r.game.State.ActiveQuestion.QuestionIndex,
		)] = true
	}
	r.game.State.ActiveQuestion = nil
	r.game.State.BuzzWinnerID = ""
	r.game.State.SubmittedAnswer = ""
	r.game.State.AnswerDeadline = time.Time{}

	if r.game.AllCellsUsed() {
		r.game.State.Phase = game.PhaseGameOver
	} else {
		r.game.State.Phase = game.PhaseBoard
	}
	r.broadcastState()
}

// --- Lifecycle ---

func (r *Room) clientLeft(c *Client) {
	delete(r.clients, c.getSessionID())

	p := r.game.Players[c.getSessionID()]
	if p == nil {
		return
	}
	p.Connected = false

	if p.IsHost {
		p.IsHost = false
		r.game.PromoteNextHost()
	}

	if len(r.clients) == 0 && !r.closed {
		r.closed = true
		r.stopBuzzTimer()
		r.stopAnswerTimer()
		close(r.quit)
		return
	}

	if !r.closed {
		r.broadcastState()
	}
}

func (r *Room) stopBuzzTimer() {
	if r.buzzTimer != nil {
		r.buzzTimer.Stop()
		r.buzzTimer = nil
	}
	r.buzzWindowOpen = false
}

func (r *Room) stopAnswerTimer() {
	if r.answerTimer != nil {
		r.answerTimer.Stop()
		r.answerTimer = nil
	}
	// Invalidate generation so any in-flight callback is ignored.
	r.answerTimerGen++
}

// --- Helpers ---

func (r *Room) isHost(c *Client) bool {
	p := r.game.Players[c.getSessionID()]
	return p != nil && p.IsHost
}

// canChoose returns true if the client is allowed to select a question.
// The active chooser always can; the host may as a fallback when the chooser is disconnected.
func (r *Room) canChoose(c *Client) bool {
	sid := c.getSessionID()
	if sid == r.game.State.ActiveChooserID {
		return true
	}
	chooser := r.game.Players[r.game.State.ActiveChooserID]
	chooserGone := chooser == nil || !chooser.Connected
	return r.isHost(c) && chooserGone
}

// randomPlayerID picks a random connected non-host player session ID.
// Returns empty string if no eligible players exist.
func (r *Room) randomPlayerID() string {
	var ids []string
	for id, p := range r.game.Players {
		if !p.IsHost && p.Connected {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return ""
	}
	return ids[rand.Intn(len(ids))]
}

// --- Broadcasting ---

type roomStatePayload struct {
	RoomID           string               `json:"room_id"`
	YourRole         string               `json:"your_role"`
	YourSessionID    string               `json:"your_session_id"`
	Phase            game.Phase           `json:"phase"`
	Players          []*game.Player       `json:"players"`
	Categories       []categoryView       `json:"categories,omitempty"`
	UsedCells        map[string]bool      `json:"used_cells"`
	ActiveQuestion   *game.ActiveQuestion `json:"active_question,omitempty"`
	ActiveChooserID  string               `json:"active_chooser_id,omitempty"`
	BuzzWinnerID     string               `json:"buzz_winner_id,omitempty"`
	AnswerDeadlineMS int64                `json:"answer_deadline_ms,omitempty"`
	AnswerDurationMS int64                `json:"answer_duration_ms"`
	SubmittedAnswer  string               `json:"submitted_answer,omitempty"`
	CorrectAnswer    string               `json:"correct_answer,omitempty"` // host only
	AnswerPending    bool                 `json:"answer_pending"`
}

type categoryView struct {
	Name      string         `json:"name"`
	Questions []questionView `json:"questions"`
}

type questionView struct {
	Value int `json:"value"`
}

func (r *Room) broadcastState() {
	categories := r.buildCategoryView()

	players := make([]*game.Player, 0, len(r.game.Players))
	for _, p := range r.game.Players {
		players = append(players, p)
	}

	var deadlineMS int64
	if !r.game.State.AnswerDeadline.IsZero() {
		deadlineMS = r.game.State.AnswerDeadline.UnixMilli()
	}

	for sessionID, c := range r.clients {
		role := "player"
		if p := r.game.Players[sessionID]; p != nil && p.IsHost {
			role = "host"
		}

		// Host-only fields: submitted answer during judging, correct answer when a question is active.
		submittedAnswer := ""
		correctAnswer := ""
		answerPending := r.game.State.Phase == game.PhaseJudging
		if role == "host" {
			if answerPending {
				submittedAnswer = r.game.State.SubmittedAnswer
			}
			if aq := r.game.State.ActiveQuestion; aq != nil && r.game.State.Pack != nil {
				correctAnswer = r.game.State.Pack.Categories[aq.CategoryIndex].Questions[aq.QuestionIndex].Answer
			}
		}

		c.sendMsg(MsgRoomStateUpdate, roomStatePayload{
			RoomID:           r.game.ID,
			YourRole:         role,
			YourSessionID:    sessionID,
			Phase:            r.game.State.Phase,
			Players:          players,
			Categories:       categories,
			UsedCells:        r.game.State.UsedCells,
			ActiveQuestion:   r.game.State.ActiveQuestion,
			ActiveChooserID:  r.game.State.ActiveChooserID,
			BuzzWinnerID:     r.game.State.BuzzWinnerID,
			AnswerDeadlineMS: deadlineMS,
			AnswerDurationMS: r.game.State.AnswerDuration.Milliseconds(),
			SubmittedAnswer:  submittedAnswer,
			CorrectAnswer:    correctAnswer,
			AnswerPending:    answerPending,
		})
	}
}

func (r *Room) buildCategoryView() []categoryView {
	if r.game.State.Pack == nil {
		return nil
	}
	cats := make([]categoryView, len(r.game.State.Pack.Categories))
	for i, cat := range r.game.State.Pack.Categories {
		qs := make([]questionView, len(cat.Questions))
		for j, q := range cat.Questions {
			qs[j] = questionView{Value: q.Value}
		}
		cats[i] = categoryView{Name: cat.Name, Questions: qs}
	}
	return cats
}

func (r *Room) broadcast(msgType string, payload any) {
	data, err := encode(msgType, payload)
	if err != nil {
		log.Printf("broadcast encode %s: %v", msgType, err)
		return
	}
	for _, c := range r.clients {
		select {
		case c.send <- data:
		default:
		}
	}
}
