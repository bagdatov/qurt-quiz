import { useGameStore } from '../store/gameStore'

export default function Toasts() {
  const { toasts, removeToast } = useGameStore()

  return (
    <div className="fixed bottom-4 right-4 flex flex-col gap-2 z-50 max-w-xs w-full">
      {toasts.map(t => (
        <div
          key={t.id}
          onClick={() => removeToast(t.id)}
          className={`card px-4 py-3 text-sm cursor-pointer border-l-4 transition-all ${
            t.kind === 'error'
              ? 'border-quiz-red'
              : t.kind === 'success'
                ? 'border-quiz-green'
                : 'border-quiz-accent'
          }`}
        >
          {t.message}
        </div>
      ))}
    </div>
  )
}
