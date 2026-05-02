# Qurt Quiz

Real-time multiplayer quiz game. One host controls a Jeopardy-style board; up to 5 players race to buzz in and answer questions. The host judges each answer and scores update live across all devices.

## Requirements

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose (optional, for containerised setup)

## Quick Start

### Local Development

**1. Clone and install**

```bash
git clone https://github.com/qurt-quiz/qurt-quiz.git
cd qurt-quiz
```

**2. Start the backend**

```bash
go run ./cmd/server
# Server listens on http://localhost:8080
```

**3. Start the frontend** (in a separate terminal)

```bash
cd frontend
npm install
npm run dev
# UI available at http://localhost:5173
```

The Vite dev server proxies `/api` and `/ws` to the backend automatically.

### Using Make

```bash
make install   # install frontend dependencies
make dev       # start backend + frontend concurrently
make build     # compile backend binary + frontend bundle
make test      # run Go tests
make lint      # run Go vet + frontend ESLint
make clean     # remove build artefacts
```

### Docker Compose

```bash
docker compose up --build
# Full stack at http://localhost:5173 (frontend) and http://localhost:8080 (backend)
```

For a single-container production build:

```bash
docker build -t qurt-quiz .
docker run -p 8080:8080 qurt-quiz
# App served at http://localhost:8080
```

## Configuration

The backend reads configuration from environment variables or CLI flags (flags take precedence):

| Flag / Env var | Default | Description |
|---|---|---|
| `-addr` / `ADDR` | `:8080` | Listen address |
| `-packs` / `PACK_DIR` | `packs` | Directory of quiz pack JSON files |
| `-static` / `STATIC_DIR` | _(empty)_ | Serve built frontend from this directory |

Example production run:

```bash
./qurt-quiz -addr :80 -packs /data/packs -static /data/static
```

## Adding Quiz Packs

Create a JSON file in the `packs/` directory following the naming convention `{pack_id}_{lang}.json`:

```json
{
  "id": "movies",
  "lang": "en",
  "title": "Cinema",
  "categories": [
    {
      "name": "Directors",
      "questions": [
        { "value": 100, "text": "Who directed Inception?", "answer": "Christopher Nolan" },
        { "value": 200, "text": "...", "answer": "..." },
        { "value": 300, "text": "...", "answer": "..." },
        { "value": 400, "text": "...", "answer": "..." },
        { "value": 500, "text": "...", "answer": "..." }
      ]
    }
  ]
}
```

Rules:
- Each pack must have exactly 5 categories, each with exactly 5 questions at values 100, 200, 300, 400, 500.
- Multiple language files share the same `id` (e.g., `movies_en.json` and `movies_ru.json`). The host picks the language when creating a room.
- The server hard-fails on a malformed pack at startup — validate your JSON before deploying.

## Game Flow

1. **Host** visits the app, enters their name, selects a quiz pack and language, clicks **Create Room**.
2. A 6-character room code is displayed. **Players** visit the app, enter their name and the room code, click **Join Room**.
3. Host clicks **Start Game** (requires at least 1 player).
4. Host clicks a point value on the board to reveal a question, then clicks **Open Buzzer**.
5. Players tap the red **BUZZ** button — the first to buzz in (with a 50 ms fairness window) gets to answer.
6. The answering player types their answer and clicks **Submit** within 30 seconds.
7. Host sees the submitted answer and clicks **Correct** (+points) or **Wrong** (−points).
8. Repeat until the board is cleared. Final scores are shown on the Game Over screen.

**Reconnecting:** If a player or host loses their connection, they can rejoin the same room with the same code. Their score is preserved.

**Host leaving:** If the host disconnects, the next connected player is automatically promoted to host.

## Project Structure

```
cmd/server/         Go entry point
internal/
  hub/              WebSocket hub, room actor, client pumps
  game/             Game state types and helpers
  pack/             Quiz pack loading and validation
  store/            Storage interface + in-memory implementation
frontend/
  src/
    components/     Board, BuzzerView, AnswerView, JudgeView, Scoreboard, …
    pages/          Home (create/join), Room (game)
    store/          Zustand game store
    ws/             WebSocket singleton
packs/              Sample quiz packs (EN + RU)
Dockerfile          Multi-stage build
docker-compose.yml  Local dev and production compose
```

## Development Notes

- The Go backend uses an actor-per-room model: all game state mutations are serialised through a channel, so no mutexes are needed on the game state itself.
- Session identity is a UUID generated in the browser and stored in `localStorage`. The server uses it to restore player state on reconnect.
- The storage layer is behind a `RoomStore` interface (`internal/store/store.go`). Swapping in MongoDB requires implementing the three-method interface and updating the `hub.New(...)` call in `main.go`.
