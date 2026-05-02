# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A real-time browser-based multiplayer quiz game (Jeopardy-style). Mobile-first, fully responsive. One Host controls the board; up to 5 Players buzz in and type answers; the Host judges each answer and scores update live.

## Tech Stack

- **Backend:** Go — WebSocket hub/actor pattern using `github.com/coder/websocket`
- **Frontend:** React (Vite) + TailwindCSS + Zustand
- **Storage:** In-memory (MVP); storage interface at `internal/store/store.go` ready for MongoDB swap
- **Infrastructure:** Docker Compose

## Commands

```bash
# Backend
go run ./cmd/server          # start dev server (default :8080)
go build ./...               # compile all packages
go vet ./...                 # vet all packages
go test ./...                # run all tests

# Frontend (from frontend/)
npm install
npm run dev                  # Vite dev server (default :5173, proxies to :8080)
npm run build                # production build → frontend/dist/
npm run lint

# Full stack
make dev                     # backend + frontend in one terminal
docker compose up --build    # containerised full stack
```

## Architecture

### Backend (Go)

Each room runs as an independent goroutine (actor). All game state mutations are serialised through `room.incoming` — no mutexes on game state itself. The Hub only holds the room registry under a `sync.RWMutex`.

```
cmd/server/main.go          — HTTP server, routes, SPA fallback
internal/hub/
  hub.go                    — Hub: registry of rooms, WS upgrade, message dispatch
  room.go                   — Room actor: game state machine, broadcasting
  client.go                 — Client: read/write pumps, identity (RWMutex-protected)
  messages.go               — Wire constants and payload structs
internal/game/types.go      — Room, Player, GameState, Phase, ActiveQuestion
internal/pack/
  types.go                  — QuizPack, Category, Question
  loader.go                 — Load + validate JSON; ListAvailable
internal/store/
  store.go                  — RoomStore interface
  memory/store.go           — In-memory implementation
packs/                      — Quiz pack JSON files ({id}_{lang}.json)
```

### Game Phase State Machine

```
LOBBY → BOARD → QUESTION → BUZZER → ANSWERING → JUDGING → BOARD (loop)
                                                          → GAME_OVER (when board empty)
```

- **LOBBY:** Host shares the 6-char room code; players join. Host starts game.
- **BOARD:** Host clicks a value cell; transitions to QUESTION.
- **QUESTION:** All see the question text. Host clicks "Open Buzzer".
- **BUZZER:** 50 ms collection window; random winner selected (jitter-fair). Phase → ANSWERING.
- **ANSWERING:** Winner types answer within `AnswerDuration` (default 30 s). On submit → JUDGING. On timeout → score penalty, back to BOARD.
- **JUDGING:** Host sees submitted text, clicks Correct/Wrong. Score applied, back to BOARD.

### WebSocket Protocol

All messages: `{ "type": "...", "payload": { ... } }`

Client → Server: `CREATE_ROOM`, `ENTER_ROOM`, `START_GAME`, `SELECT_QUESTION`, `OPEN_BUZZER`, `BUZZ_IN`, `SUBMIT_ANSWER`, `JUDGE_ANSWER`

Server → Client: `ROOM_STATE_UPDATE` (full state per client, host sees `submitted_answer`, players don't), `PLAYER_BUZZED`, `ANSWER_RESULT`, `ANSWER_TIMEOUT`, `ERROR`

### Quiz Pack JSON Format

Files named `{pack_id}_{lang}.json` (e.g., `sample_en.json`, `sample_ru.json`). Each pack: 5 categories × 5 questions at values 100–500.

```json
{
  "id": "sample", "lang": "en", "title": "General Knowledge",
  "categories": [
    { "name": "Science", "questions": [
      { "value": 100, "text": "...", "answer": "..." }
    ]}
  ]
}
```

### Frontend

Single WS connection managed as a module singleton (`src/ws/socket.ts`). Zustand store (`src/store/gameStore.ts`) holds all server-pushed state and exposes action functions that call `send()`. The `Room` page renders different components based on `phase` and `your_role` from the server.

Session identity: browser generates a UUID on first visit, stored in `localStorage`. Sent on every connect — the server uses it to reconnect players to their existing slot with score preserved.
