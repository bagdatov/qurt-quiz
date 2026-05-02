import { useGameStore } from '../store/gameStore'

export default function BuzzerView() {
  const { your_role, buzzIn, active_question } = useGameStore()

  return (
    <div className="flex flex-col items-center justify-center gap-8 py-8 text-center">
      {active_question && (
        <p className="text-quiz-muted text-sm max-w-xl px-4">{active_question.text}</p>
      )}

      {your_role === 'player' ? (
        <>
          <p className="text-quiz-gold font-display text-3xl">BUZZ IN!</p>
          <button
            onClick={buzzIn}
            className="w-52 h-52 rounded-full bg-quiz-red hover:bg-red-400 active:scale-90 transition-all duration-100 shadow-[0_0_60px_rgba(239,68,68,0.5)] font-display text-4xl text-white select-none"
          >
            BUZZ
          </button>
          <p className="text-quiz-muted text-sm">First to buzz gets to answer</p>
        </>
      ) : (
        <div className="flex flex-col items-center gap-4">
          <div className="w-24 h-24 rounded-full bg-quiz-red/20 border-2 border-quiz-red animate-pulse" />
          <p className="text-quiz-muted">Waiting for a player to buzz in…</p>
        </div>
      )}
    </div>
  )
}
