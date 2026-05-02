import { useGameStore } from '../store/gameStore'

export default function JudgeView() {
  const { active_question, submitted_answer, correct_answer, buzz_winner_id, players, judgeAnswer } = useGameStore()
  const winner = players.find(p => p.id === buzz_winner_id)

  return (
    <div className="flex flex-col items-center gap-5 py-8 max-w-xl mx-auto px-4 text-center">
      <h2 className="font-display text-3xl text-quiz-gold">Judge the Answer</h2>

      {active_question && (
        <div className="card w-full">
          <p className="text-xs text-quiz-muted mb-1">Question ({active_question.value} pts)</p>
          <p className="text-quiz-text">{active_question.text}</p>
        </div>
      )}

      {/* Player's submitted answer */}
      <div className="card w-full bg-quiz-accent/10 border-quiz-accent">
        <p className="text-xs text-quiz-muted mb-1">
          {winner?.name ?? 'Player'}'s answer
        </p>
        <p className="text-quiz-text text-2xl font-semibold">
          {submitted_answer || '(no answer submitted)'}
        </p>
      </div>

      {/* Correct answer from the pack */}
      {correct_answer && (
        <div className="card w-full bg-quiz-green/10 border-quiz-green">
          <p className="text-xs text-quiz-muted mb-1">Correct answer</p>
          <p className="text-quiz-green text-lg font-semibold">{correct_answer}</p>
        </div>
      )}

      <div className="flex gap-4 w-full">
        <button className="btn-success flex-1 text-lg" onClick={() => judgeAnswer(true)}>
          ✓ Correct
        </button>
        <button className="btn-danger flex-1 text-lg" onClick={() => judgeAnswer(false)}>
          ✗ Wrong
        </button>
      </div>

      <p className="text-xs text-quiz-muted">
        Correct: +{active_question?.value ?? 0} pts &nbsp;·&nbsp; Wrong: −{active_question?.value ?? 0} pts
      </p>
    </div>
  )
}
