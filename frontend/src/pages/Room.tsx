import { useEffect } from 'react'
import type { Player } from '../types'
import { useParams, useNavigate } from 'react-router-dom'
import { useGameStore } from '../store/gameStore'
import Lobby from '../components/Lobby'
import Board from '../components/Board'
import QuestionView from '../components/QuestionView'
import BuzzerView from '../components/BuzzerView'
import AnswerView from '../components/AnswerView'
import JudgeView from '../components/JudgeView'
import GameOver from '../components/GameOver'
import Scoreboard from '../components/Scoreboard'

export default function Room() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { room_id, phase, your_role, players, enterRoom, sessionId } = useGameStore()

  // If we land on this page without a room (e.g. direct link), request to enter.
  useEffect(() => {
    if (!id) { navigate('/'); return }
    if (!room_id) {
      // The store will attempt reconnect via stored session; if that fails,
      // we need a name. Show the join modal by navigating home with the code pre-filled.
      const savedRoom = localStorage.getItem('qurt_room_id')
      if (savedRoom === id) {
        // Reconnect attempt is already fired by init() in gameStore.
        return
      }
      // New visitor — send them to home to enter their name.
      navigate(`/?join=${id}`)
    }
  }, [id, room_id, navigate, sessionId, enterRoom])

  if (!room_id) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-quiz-muted animate-pulse">Connecting…</p>
      </div>
    )
  }

  const showBoard = ['BOARD', 'QUESTION', 'BUZZER', 'ANSWERING', 'JUDGING'].includes(phase)

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="flex items-center justify-between px-4 py-3 border-b border-quiz-border">
        <span className="font-display text-2xl text-quiz-gold">QURT QUIZ</span>
        <div className="flex items-center gap-3">
          <span className="text-xs text-quiz-muted uppercase">
            Room <span className="font-mono text-quiz-text tracking-widest">{room_id}</span>
          </span>
          <span className={`text-xs px-2 py-0.5 rounded-full border ${
            your_role === 'host'
              ? 'border-quiz-gold text-quiz-gold'
              : 'border-quiz-muted text-quiz-muted'
          }`}>
            {your_role}
          </span>
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        {/* Main area */}
        <main className="flex-1 overflow-auto p-4">
          {phase === 'LOBBY'     && <Lobby />}
          {phase === 'BOARD'     && <Board />}
          {phase === 'QUESTION'  && <QuestionView />}
          {phase === 'BUZZER'    && <BuzzerView />}
          {phase === 'ANSWERING' && <AnswerView />}
          {phase === 'JUDGING'   && (your_role === 'host' ? <JudgeView /> : <WaitingForJudge />)}
          {phase === 'GAME_OVER' && <GameOver />}
        </main>

        {/* Scoreboard sidebar — visible on board and game phases */}
        {showBoard && (
          <aside className="hidden sm:block w-52 border-l border-quiz-border p-4 overflow-y-auto">
            <HostCard players={players} />
            <Scoreboard />
          </aside>
        )}
      </div>

      {/* Mobile scoreboard strip */}
      {showBoard && (
        <div className="sm:hidden border-t border-quiz-border px-4 py-2">
          <HostCard players={players} compact />
          <Scoreboard compact />
        </div>
      )}
    </div>
  )
}

function HostCard({ players, compact }: { players: Player[]; compact?: boolean }) {
  const host = players.find(p => p.is_host)
  if (!host) return null

  if (compact) {
    return (
      <div className="flex items-center gap-1.5 mb-2">
        <span className="text-xs font-bold text-quiz-gold uppercase tracking-wide">Host</span>
        <span className="text-xs text-quiz-text">{host.name}</span>
        {!host.connected && <span className="text-xs text-quiz-muted">(away)</span>}
      </div>
    )
  }

  return (
    <div className="mb-4 pb-4 border-b border-quiz-border/50">
      <p className="text-xs font-bold text-quiz-gold uppercase tracking-widest mb-1">Host</p>
      <div className="flex items-center gap-2">
        <div className={`w-2 h-2 rounded-full flex-shrink-0 ${host.connected ? 'bg-green-400' : 'bg-quiz-muted'}`} />
        <span className="text-sm text-quiz-text truncate">{host.name}</span>
      </div>
    </div>
  )
}

function WaitingForJudge() {
  const { answer_pending } = useGameStore()
  return (
    <div className="flex flex-col items-center justify-center h-full gap-4 py-20">
      <div className="text-5xl">⏳</div>
      <p className="text-quiz-muted text-center">
        {answer_pending ? 'Waiting for the host to judge the answer…' : 'Waiting…'}
      </p>
    </div>
  )
}
