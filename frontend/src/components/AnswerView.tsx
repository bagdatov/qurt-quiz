import { useState, useEffect, useRef } from 'react'
import { useGameStore } from '../store/gameStore'

export default function AnswerView() {
  const {
    your_session_id,
    buzz_winner_id,
    active_question,
    answer_deadline_ms,
    answer_duration_ms,
    submitAnswer,
    players,
  } = useGameStore()

  const isAnswering = your_session_id === buzz_winner_id
  const [text, setText] = useState('')
  const [timeLeft, setTimeLeft] = useState(30)
  const inputRef = useRef<HTMLInputElement>(null)

  const winner = players.find(p => p.id === buzz_winner_id)

  useEffect(() => {
    if (!answer_deadline_ms) return

    function tick() {
      const remaining = Math.max(0, answer_deadline_ms - Date.now())
      setTimeLeft(Math.ceil(remaining / 1000))
    }

    tick()
    const interval = setInterval(tick, 200)
    return () => clearInterval(interval)
  }, [answer_deadline_ms])

  useEffect(() => {
    if (isAnswering) inputRef.current?.focus()
  }, [isAnswering])

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!text.trim()) return
    submitAnswer(text.trim())
  }

  const totalSeconds = answer_duration_ms / 1000
  const pct = (timeLeft / totalSeconds) * 100
  const timerColor = timeLeft > 10 ? 'bg-quiz-green' : timeLeft > 5 ? 'bg-quiz-gold' : 'bg-quiz-red'

  return (
    <div className="flex flex-col items-center gap-6 py-8 max-w-xl mx-auto w-full px-4">
      {active_question && (
        <div className="card w-full text-center">
          <p className="text-quiz-muted text-xs mb-1">{active_question.value} pts</p>
          <p className="text-quiz-text text-lg">{active_question.text}</p>
        </div>
      )}

      {/* Timer bar */}
      <div className="w-full">
        <div className="flex justify-between text-xs text-quiz-muted mb-1">
          <span>{winner?.name ?? 'Player'} is answering…</span>
          <span>{timeLeft}s</span>
        </div>
        <div className="w-full h-2 bg-quiz-border rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-200 ${timerColor}`}
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>

      {isAnswering ? (
        <form onSubmit={handleSubmit} className="flex flex-col gap-3 w-full">
          <input
            ref={inputRef}
            className="input text-center text-xl"
            placeholder="Type your answer…"
            value={text}
            onChange={e => setText(e.target.value)}
            autoComplete="off"
          />
          <button type="submit" className="btn-primary" disabled={!text.trim()}>
            Submit Answer
          </button>
        </form>
      ) : (
        <p className="text-quiz-muted animate-pulse">
          {winner?.name ?? 'A player'} is typing their answer…
        </p>
      )}
    </div>
  )
}
