import { useGameStore } from '../store/gameStore'

export default function Lobby() {
  const { players, your_role, startGame, room_id } = useGameStore()
  const shareUrl = `${location.origin}/room/${room_id}`

  const playerList = players.filter(p => !p.is_host)
  const host = players.find(p => p.is_host)
  const canStart = playerList.length >= 1

  function copyLink() {
    navigator.clipboard.writeText(shareUrl).catch(() => {})
  }

  return (
    <div className="flex flex-col items-center gap-8 py-8 max-w-md mx-auto">
      <div className="text-center">
        <h2 className="font-display text-4xl text-quiz-gold mb-1">Waiting Room</h2>
        <p className="text-quiz-muted text-sm">Share the code or link below</p>
      </div>

      {/* Room code */}
      <div className="card text-center w-full">
        <p className="text-xs text-quiz-muted uppercase tracking-wider mb-1">Room code</p>
        <p className="font-display text-5xl text-quiz-text tracking-widest">{room_id}</p>
        <button onClick={copyLink} className="btn-ghost mt-3 text-sm w-full">
          Copy invite link
        </button>
      </div>

      {/* Player list */}
      <div className="card w-full">
        <p className="text-xs text-quiz-muted uppercase tracking-wider mb-3">
          Players ({playerList.length}/5)
        </p>
        <ul className="flex flex-col gap-2">
          {host && (
            <li className="flex items-center gap-2">
              <span className="w-2 h-2 rounded-full bg-quiz-gold" />
              <span>{host.name}</span>
              <span className="text-xs text-quiz-gold ml-auto">Host</span>
            </li>
          )}
          {playerList.map(p => (
            <li key={p.id} className="flex items-center gap-2">
              <span className={`w-2 h-2 rounded-full ${p.connected ? 'bg-quiz-green' : 'bg-quiz-muted'}`} />
              <span>{p.name}</span>
              {!p.connected && <span className="text-xs text-quiz-muted ml-auto">away</span>}
            </li>
          ))}
          {playerList.length === 0 && (
            <li className="text-quiz-muted text-sm">No players yet…</li>
          )}
        </ul>
      </div>

      {your_role === 'host' && (
        <button
          className="btn-primary w-full text-lg"
          onClick={startGame}
          disabled={!canStart}
          title={!canStart ? 'Need at least 1 player to start' : ''}
        >
          Start Game
        </button>
      )}

      {your_role === 'player' && (
        <p className="text-quiz-muted text-sm animate-pulse">Waiting for host to start…</p>
      )}
    </div>
  )
}
