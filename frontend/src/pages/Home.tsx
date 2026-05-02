import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useGameStore } from '../store/gameStore'

interface PackInfo {
  id: string
  langs: string[]
}

export default function Home() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const joinParam = searchParams.get('join') ?? ''

  const { createRoom, enterRoom, room_id } = useGameStore()

  const [packs, setPacks] = useState<PackInfo[]>([])
  const [name, setName] = useState('')
  const [packId, setPackId] = useState('')
  const [lang, setLang] = useState('')
  const [joinCode, setJoinCode] = useState(joinParam.toUpperCase())
  const [tab, setTab] = useState<'create' | 'join'>(joinParam ? 'join' : 'create')

  // Navigate to room once the server responds with a room_id.
  useEffect(() => {
    if (room_id) navigate(`/room/${room_id}`)
  }, [room_id, navigate])

  useEffect(() => {
    fetch('/api/packs')
      .then(r => r.json())
      .then(d => {
        const list: PackInfo[] = d.packs ?? []
        setPacks(list)
        if (list.length > 0) {
          setPackId(list[0].id)
          setLang(list[0].langs[0] ?? 'en')
        }
      })
      .catch(console.error)
  }, [])

  const availableLangs = packs.find(p => p.id === packId)?.langs ?? []

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim() || !packId || !lang) return
    createRoom(name.trim(), packId, lang)
  }

  function handleJoin(e: React.FormEvent) {
    e.preventDefault()
    const code = joinCode.trim().toUpperCase()
    if (!code || !name.trim()) return
    enterRoom(code, name.trim())
  }

  return (
    <div className="min-h-screen flex flex-col items-center justify-center px-4 py-10 gap-8">
      <div className="text-center">
        <h1 className="font-display text-6xl sm:text-8xl text-quiz-gold tracking-widest">
          QURT QUIZ
        </h1>
        <p className="text-quiz-muted mt-2 text-sm">Real-time Jeopardy-style multiplayer</p>
      </div>

      <div className="card w-full max-w-md">
        {/* Tab switcher */}
        <div className="flex mb-6 rounded-xl overflow-hidden border border-quiz-border">
          {(['create', 'join'] as const).map(t => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`flex-1 py-2 text-sm font-semibold transition-colors ${
                tab === t
                  ? 'bg-quiz-accent text-white'
                  : 'text-quiz-muted hover:text-quiz-text'
              }`}
            >
              {t === 'create' ? 'Create Room' : 'Join Room'}
            </button>
          ))}
        </div>

        {tab === 'create' ? (
          <form onSubmit={handleCreate} className="flex flex-col gap-4">
            <Field label="Your name">
              <input
                className="input"
                placeholder="Enter your name"
                value={name}
                onChange={e => setName(e.target.value)}
                maxLength={24}
                autoFocus
                required
              />
            </Field>

            <Field label="Quiz pack">
              <select
                className="input"
                value={packId}
                onChange={e => {
                  setPackId(e.target.value)
                  const p = packs.find(p => p.id === e.target.value)
                  setLang(p?.langs[0] ?? 'en')
                }}
              >
                {packs.map(p => (
                  <option key={p.id} value={p.id}>{p.id}</option>
                ))}
              </select>
            </Field>

            <Field label="Language">
              <select className="input" value={lang} onChange={e => setLang(e.target.value)}>
                {availableLangs.map(l => (
                  <option key={l} value={l}>{l.toUpperCase()}</option>
                ))}
              </select>
            </Field>

            <button type="submit" className="btn-primary mt-2" disabled={!name.trim() || !packId}>
              Create Room
            </button>
          </form>
        ) : (
          <form onSubmit={handleJoin} className="flex flex-col gap-4">
            <Field label="Your name">
              <input
                className="input"
                placeholder="Enter your name"
                value={name}
                onChange={e => setName(e.target.value)}
                maxLength={24}
                autoFocus
                required
              />
            </Field>

            <Field label="Room code">
              <input
                className="input uppercase tracking-widest text-center text-lg"
                placeholder="ABC123"
                value={joinCode}
                onChange={e => setJoinCode(e.target.value.toUpperCase())}
                maxLength={6}
                required
              />
            </Field>

            <button
              type="submit"
              className="btn-primary mt-2"
              disabled={!name.trim() || joinCode.length < 6}
            >
              Join Room
            </button>
          </form>
        )}
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-xs text-quiz-muted uppercase tracking-wider">{label}</span>
      {children}
    </label>
  )
}
