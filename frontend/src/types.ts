export type Phase =
  | 'LOBBY'
  | 'BOARD'
  | 'QUESTION'
  | 'BUZZER'
  | 'ANSWERING'
  | 'JUDGING'
  | 'GAME_OVER'

export type Role = 'host' | 'player'

export interface Player {
  id: string
  name: string
  score: number
  is_host: boolean
  connected: boolean
}

export interface QuestionView {
  value: number
}

export interface CategoryView {
  name: string
  questions: QuestionView[]
}

export interface ActiveQuestion {
  category_index: number
  question_index: number
  text: string
  value: number
}

export interface RoomState {
  room_id: string
  your_role: Role
  your_session_id: string
  phase: Phase
  players: Player[]
  categories: CategoryView[]
  used_cells: Record<string, boolean>
  active_question: ActiveQuestion | null
  active_chooser_id: string
  buzz_winner_id: string
  answer_deadline_ms: number
  answer_duration_ms: number
  submitted_answer: string
  correct_answer: string
  answer_pending: boolean
}

export interface AnswerResult {
  correct: boolean
  player_id: string
  score_delta: number
}

export interface ServerEnvelope {
  type: string
  payload: unknown
}
