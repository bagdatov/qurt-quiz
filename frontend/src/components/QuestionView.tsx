import { useGameStore } from '../store/gameStore'

export default function QuestionView() {
  const { active_question, your_role, openBuzzer } = useGameStore()

  if (!active_question) return null

  return (
    <div className="flex flex-col items-center justify-center gap-8 py-12 max-w-2xl mx-auto text-center">
      <p className="text-quiz-gold font-display text-4xl sm:text-5xl">
        {active_question.value} pts
      </p>

      <div className="card w-full">
        <p className="text-quiz-text text-xl sm:text-3xl leading-relaxed font-medium">
          {active_question.text}
        </p>
      </div>

      {your_role === 'host' ? (
        <button className="btn-primary text-lg px-12" onClick={openBuzzer}>
          Open Buzzer
        </button>
      ) : (
        <p className="text-quiz-muted animate-pulse">Host is reading the question…</p>
      )}
    </div>
  )
}
