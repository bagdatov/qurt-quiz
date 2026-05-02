import { create } from 'zustand'
import { connect, send, addListener } from '../ws/socket'
import type { RoomState, AnswerResult, Phase } from '../types'

const SESSION_KEY = 'qurt_session_id'
const ROOM_KEY = 'qurt_room_id'

function getOrCreateSessionId(): string {
  let id = localStorage.getItem(SESSION_KEY)
  if (!id) {
    id = crypto.randomUUID()
    localStorage.setItem(SESSION_KEY, id)
  }
  return id
}

interface Toast {
  id: number
  message: string
  kind: 'info' | 'success' | 'error'
}

interface GameStore extends RoomState {
  sessionId: string
  connected: boolean
  lastResult: AnswerResult | null
  toasts: Toast[]
  _toastCounter: number

  // Actions
  init: () => void
  createRoom: (name: string, packId: string, lang: string) => void
  enterRoom: (roomId: string, name?: string) => void
  startGame: () => void
  selectQuestion: (catIdx: number, qIdx: number) => void
  openBuzzer: () => void
  buzzIn: () => void
  submitAnswer: (text: string) => void
  judgeAnswer: (correct: boolean) => void
  addToast: (message: string, kind?: Toast['kind']) => void
  removeToast: (id: number) => void
}

const defaultState: RoomState = {
  room_id: '',
  your_role: 'player',
  your_session_id: '',
  phase: 'LOBBY' as Phase,
  players: [],
  categories: [],
  used_cells: {},
  active_question: null,
  buzz_winner_id: '',
  answer_deadline_ms: 0,
  answer_duration_ms: 30000,
  submitted_answer: '',
  answer_pending: false,
}

// Guard against double-init from React StrictMode or multiple mount points.
let storeInitialized = false

export const useGameStore = create<GameStore>((set, get) => ({
  ...defaultState,
  sessionId: getOrCreateSessionId(),
  connected: false,
  lastResult: null,
  toasts: [],
  _toastCounter: 0,

  init() {
    if (storeInitialized) return
    storeInitialized = true

    const sessionId = get().sessionId

    connect(() => {
      set({ connected: true })
      // Auto-reconnect if we have a saved room.
      const savedRoom = localStorage.getItem(ROOM_KEY)
      if (savedRoom) {
        send('ENTER_ROOM', { session_id: sessionId, room_id: savedRoom })
      }
    })

    const remove = addListener((type, payload) => {
      switch (type) {
        case 'ROOM_STATE_UPDATE': {
          const state = payload as RoomState
          if (state.room_id) {
            localStorage.setItem(ROOM_KEY, state.room_id)
          }
          set({ ...state, lastResult: null })
          break
        }
        case 'ANSWER_RESULT':
        case 'ANSWER_TIMEOUT': {
          const result = payload as AnswerResult
          set({ lastResult: result })
          const { players } = get()
          const player = players.find(p => p.id === result.player_id)
          const name = player?.name ?? 'Player'
          if (type === 'ANSWER_TIMEOUT') {
            get().addToast(`⏱ ${name} ran out of time! −${result.score_delta * -1} pts`, 'error')
          } else {
            get().addToast(
              result.correct
                ? `✓ ${name} answered correctly! +${result.score_delta} pts`
                : `✗ ${name} answered incorrectly! ${result.score_delta} pts`,
              result.correct ? 'success' : 'error',
            )
          }
          break
        }
        case 'ERROR': {
          const { message } = payload as { message: string }
          get().addToast(message, 'error')
          break
        }
      }
    })

    // Cleanup on unmount (not called in practice for top-level store).
    return remove
  },

  createRoom(name, packId, lang) {
    send('CREATE_ROOM', {
      session_id: get().sessionId,
      name,
      pack_id: packId,
      lang,
    })
  },

  enterRoom(roomId, name) {
    send('ENTER_ROOM', {
      session_id: get().sessionId,
      room_id: roomId,
      name: name ?? '',
    })
  },

  startGame() { send('START_GAME', {}) },
  openBuzzer() { send('OPEN_BUZZER', {}) },
  buzzIn() { send('BUZZ_IN', {}) },

  selectQuestion(catIdx, qIdx) {
    send('SELECT_QUESTION', { category_index: catIdx, question_index: qIdx })
  },

  submitAnswer(text) {
    send('SUBMIT_ANSWER', { text })
  },

  judgeAnswer(correct) {
    send('JUDGE_ANSWER', { correct })
  },

  addToast(message, kind = 'info') {
    const id = get()._toastCounter + 1
    set(s => ({ _toastCounter: id, toasts: [...s.toasts, { id, message, kind }] }))
    setTimeout(() => get().removeToast(id), 4000)
  },

  removeToast(id) {
    set(s => ({ toasts: s.toasts.filter(t => t.id !== id) }))
  },
}))
