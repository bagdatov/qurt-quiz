import { useGameStore } from '../store/gameStore'

export default function Board() {
  const {
    categories,
    used_cells,
    your_role,
    your_session_id,
    active_chooser_id,
    players,
    selectQuestion,
  } = useGameStore()

  if (!categories || categories.length === 0) {
    return <div className="text-quiz-muted text-center py-20">Loading board…</div>
  }

  const chooser = players.find(p => p.id === active_chooser_id)
  const chooserConnected = chooser?.connected ?? false
  // The active chooser picks. Host steps in only if the chooser is disconnected.
  const canSelect =
    your_session_id === active_chooser_id ||
    (your_role === 'host' && !chooserConnected)

  const isMyTurn = your_session_id === active_chooser_id

  return (
    <div className="flex flex-col gap-3">
      {/* Chooser banner */}
      <div className={`text-center text-sm py-2 px-4 rounded-xl border ${
        isMyTurn
          ? 'border-quiz-gold text-quiz-gold bg-quiz-gold/10'
          : 'border-quiz-border text-quiz-muted'
      }`}>
        {isMyTurn
          ? '⭐ Your turn — pick a question!'
          : chooser
            ? `Waiting for ${chooser.name} to pick…`
            : 'Waiting for a player to pick…'}
      </div>

      <div className="w-full overflow-x-auto">
        <table className="w-full border-collapse min-w-[320px]">
          <thead>
            <tr>
              {categories.map((cat, i) => (
                <th
                  key={i}
                  className="bg-quiz-accent text-white font-display text-sm sm:text-base p-2 sm:p-3 text-center border border-quiz-border/50 uppercase tracking-wide"
                >
                  {cat.name}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {Array.from({ length: 5 }).map((_, qIdx) => (
              <tr key={qIdx}>
                {categories.map((cat, catIdx) => {
                  const key = `${catIdx}:${qIdx}`
                  const used = used_cells[key]
                  const value = cat.questions[qIdx]?.value ?? (qIdx + 1) * 100

                  return (
                    <td key={catIdx} className="p-1 sm:p-1.5 border border-quiz-border/30">
                      <button
                        disabled={used || !canSelect}
                        onClick={() => selectQuestion(catIdx, qIdx)}
                        className={`w-full h-14 sm:h-20 rounded-lg font-display text-xl sm:text-3xl transition-all duration-150
                          ${used
                            ? 'bg-quiz-surface/30 text-quiz-border cursor-default'
                            : canSelect
                              ? 'bg-quiz-surface hover:bg-quiz-accent/20 hover:scale-105 active:scale-95 text-quiz-gold cursor-pointer'
                              : 'bg-quiz-surface text-quiz-gold/50 cursor-default'
                          }`}
                      >
                        {used ? '' : value}
                      </button>
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
