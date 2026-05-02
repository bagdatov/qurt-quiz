import { useNavigate } from 'react-router-dom'
import { useGameStore } from '../store/gameStore'

export default function GameOver() {
  const navigate = useNavigate()
  const { players } = useGameStore()

  const sorted = [...players].sort((a, b) => b.score - a.score)
  const winner = sorted[0]

  const medals = ['🥇', '🥈', '🥉']

  function handlePlayAgain() {
    localStorage.removeItem('qurt_room_id')
    navigate('/')
  }

  return (
    <div className="flex flex-col items-center gap-8 py-12 max-w-md mx-auto text-center px-4">
      <div>
        <h2 className="font-display text-5xl text-quiz-gold mb-2">Game Over!</h2>
        {winner && (
          <p className="text-quiz-text text-xl">
            🏆 {winner.name} wins with {winner.score} points!
          </p>
        )}
      </div>

      <div className="card w-full">
        <h3 className="text-xs text-quiz-muted uppercase tracking-wider mb-4">Final Scores</h3>
        <ul className="flex flex-col gap-3">
          {sorted.map((p, i) => (
            <li key={p.id} className="flex items-center gap-3">
              <span className="text-2xl w-8">{medals[i] ?? ''}</span>
              <span className="flex-1 text-left">{p.name}</span>
              <span className="font-bold text-quiz-gold tabular-nums">{p.score}</span>
            </li>
          ))}
        </ul>
      </div>

      <button className="btn-primary w-full" onClick={handlePlayAgain}>
        Play Again
      </button>
    </div>
  )
}
