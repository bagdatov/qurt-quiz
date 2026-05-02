import { useGameStore } from '../store/gameStore'

interface Props {
  compact?: boolean
}

export default function Scoreboard({ compact = false }: Props) {
  const { players, your_session_id } = useGameStore()

  const sorted = [...players].sort((a, b) => b.score - a.score)

  if (compact) {
    return (
      <div className="flex gap-3 overflow-x-auto py-1">
        {sorted.map(p => (
          <div
            key={p.id}
            className={`flex flex-col items-center flex-shrink-0 ${
              p.id === your_session_id ? 'text-quiz-gold' : 'text-quiz-text'
            }`}
          >
            <span className="text-xs truncate max-w-[60px]">{p.name}</span>
            <span className="text-sm font-bold">{p.score}</span>
          </div>
        ))}
      </div>
    )
  }

  return (
    <div>
      <h3 className="text-xs text-quiz-muted uppercase tracking-wider mb-3">Scores</h3>
      <ul className="flex flex-col gap-2">
        {sorted.map((p, rank) => (
          <li
            key={p.id}
            className={`flex items-center gap-2 text-sm ${
              p.id === your_session_id ? 'text-quiz-gold' : 'text-quiz-text'
            } ${!p.connected ? 'opacity-40' : ''}`}
          >
            <span className="text-quiz-muted w-4 text-right">{rank + 1}.</span>
            <span className="flex-1 truncate">{p.name}</span>
            <span className="font-bold tabular-nums">{p.score}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
