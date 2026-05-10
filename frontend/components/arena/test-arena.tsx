"use client"

import { useEffect, useState, useRef, useCallback } from "react"
import {
  ChevronRight,
  Clock,
  Code2,
  FileQuestion,
  Loader2,
  Play,
  Radio,
  Send,
  Shield,
  Swords,
  Target,
  Trophy,
  Zap,
  CheckCircle2,
  XCircle,
  AlertTriangle,
} from "lucide-react"
 update-cors-amplify
import { API_BASE, WS_BASE } from "@/lib/api-config"
import { useAntiCheat } from "@/hooks/useAntiCheat"

import { API_URL, WS_BASE } from "@/lib/api-config"
 main

// Removed hardcoded API constant

// ─── Types ────────────────────────────────────

interface Test {
  id: string
  title: string
  startTime: string
  durationSeconds: number
  isPublished: boolean
  isActive: boolean
}

interface MCQOption {
  id: string
  questionId: string
  optionText: string
}

interface CodingDetail {
  id: string
  constraints: string
  starterCode: string
  timeLimitMs: number
}

interface TestCase {
  id: string
  input: string
  expectedOutput: string
}

interface Question {
  id: string
  testId: string
  type: "mcq" | "coding"
  position: number
  title: string
  description: string
  points: number
  mcqOptions: MCQOption[] | null
  codingDetail: CodingDetail | null
  testCases: TestCase[] | null
}

interface Submission {
  id: string
  attemptId: string
  questionId: string
  type: string
  selectedOptionId: string
  code: string
  language: string
  verdict: string
  passedCount: number
  totalCount: number
}

interface RunCaseResult {
  input: string
  expected: string
  actual: string
  pass: boolean
  hidden?: boolean
  errorType?: string
  compileOut?: string
  durationMs?: number
}

// ─── Helper: format seconds to MM:SS ──────────

function formatTime(s: number): string {
  if (s < 0) s = 0
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${String(m).padStart(2, "0")}:${String(sec).padStart(2, "0")}`
}

// ─── Helper: test status ──────────────────────

function getTestStatus(t: Test): "upcoming" | "live" | "ended" {
  const now = Date.now()
  const start = new Date(t.startTime).getTime()
  const end = start + t.durationSeconds * 1000

  // If the time window has passed, it's ended regardless of isActive
  if (now > end) return "ended"
  // If within the time window, it's live
  if (now >= start) return "live"
  // If start is in the future but admin has activated it, treat as live
  // (the backend resets startTime on activation, so this is a safety net)
  if (t.isActive) return "live"
  return "upcoming"
}

const statusStyle: Record<string, { color: string; bg: string; label: string }> = {
  upcoming: { color: "text-neon-yellow", bg: "bg-neon-yellow", label: "UPCOMING" },
  live:     { color: "text-neon-cyan",   bg: "bg-neon-cyan",   label: "LIVE" },
  ended:    { color: "text-neon-pink",   bg: "bg-neon-pink",   label: "ENDED" },
}

// ═══════════════════════════════════════════════
// MAIN COMPONENT
// ═══════════════════════════════════════════════

export function TestArena() {
  // ── Top-level state: which view to show ──
  const [attemptId, setAttemptId] = useState<string | null>(null)
  const [testId, setTestId] = useState<string | null>(null)

  // On mount: check localStorage for a saved attempt
  useEffect(() => {
    const saved = localStorage.getItem("testArena_attemptId")
    const savedTest = localStorage.getItem("testArena_testId")
    if (saved && savedTest) {
      setAttemptId(saved)
      setTestId(savedTest)
    }
  }, [])

  function handleJoined(aId: string, tId: string) {
    setAttemptId(aId)
    setTestId(tId)
    localStorage.setItem("testArena_attemptId", aId)
    localStorage.setItem("testArena_testId", tId)
  }

  function handleExit() {
    setAttemptId(null)
    setTestId(null)
    localStorage.removeItem("testArena_attemptId")
    localStorage.removeItem("testArena_testId")
  }

  if (attemptId && testId) {
    return <ActiveTest attemptId={attemptId} testId={testId} onExit={handleExit} />
  }

  return <TestList onJoined={handleJoined} />
}

// ═══════════════════════════════════════════════
// VIEW 1 — Test List
// ═══════════════════════════════════════════════

function TestList({ onJoined }: { onJoined: (attemptId: string, testId: string) => void }) {
  const [tests, setTests] = useState<Test[]>([])
  const [loading, setLoading] = useState(true)
  const [joiningId, setJoiningId] = useState<string | null>(null)

  useEffect(() => {
    async function fetchTests() {
      try {
        const res = await fetch(`${API_URL}/api/arena/tests`, { credentials: "include" })
        if (res.ok) setTests(await res.json())
      } catch (e) {
        console.error("Failed to fetch tests:", e)
      } finally {
        setLoading(false)
      }
    }
    fetchTests()
  }, [])

  async function handleJoin(testId: string) {
    setJoiningId(testId)
    try {
      const res = await fetch(`${API_URL}/api/arena/tests/${testId}/join`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        const data = await res.json()
        alert(data.error || "Cannot join test")
        return
      }
      const data = await res.json()
      onJoined(data.attempt.id || data.attempt.ID, testId)
    } catch (e) {
      console.error("Join failed:", e)
    } finally {
      setJoiningId(null)
    }
  }

  return (
    <div className="relative min-h-[60vh]">
      <div className="absolute inset-0 grid-bg opacity-40" />
      <div className="absolute top-0 left-0 w-[400px] h-[400px] rounded-full bg-neon-cyan/5 blur-[120px]" />

      <div className="relative z-10 mx-auto max-w-5xl px-4 py-8 lg:px-8">
        {/* Header */}
        <div className="flex items-center gap-3 mb-2">
          <Target className="h-4 w-4 text-neon-cyan animate-pulse-glow" />
          <span className="font-mono text-[10px] tracking-[0.3em] text-neon-cyan">
            CODING CHALLENGES
          </span>
        </div>
        <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl mb-2">
          AVAILABLE <span className="text-neon-cyan text-glow-cyan">TESTS</span>
        </h2>
        <p className="text-sm text-muted-foreground mb-8">
          Join a live test to solve MCQ and coding challenges under time pressure.
        </p>

        {/* Loading */}
        {loading && (
          <div className="flex flex-col gap-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-20 w-full animate-pulse border border-panel-border bg-panel-bg/20" />
            ))}
          </div>
        )}

        {/* Empty */}
        {!loading && tests.length === 0 && (
          <div className="flex flex-col items-center justify-center py-20 gap-4">
            <FileQuestion className="h-10 w-10 text-muted-foreground" />
            <span className="font-mono text-xs tracking-widest text-muted-foreground">NO TESTS AVAILABLE</span>
          </div>
        )}

        {/* Test cards */}
        <div className="flex flex-col gap-3">
          {tests.map((test) => {
            const status = getTestStatus(test)
            const st = statusStyle[status]
            const isLive = status === "live"

            return (
              <div
                key={test.id}
                className={`group relative border border-panel-border bg-panel-bg/60 transition-all ${
                  isLive ? "hover:border-neon-cyan/40 hover:bg-panel-bg/80" : "opacity-70"
                }`}
              >
                <div className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
                  {/* Left info */}
                  <div className="flex items-center gap-4 flex-1 min-w-0">
                    <div className="flex h-10 w-10 items-center justify-center border border-panel-border text-neon-cyan">
                      <Code2 className="h-5 w-5" strokeWidth={1.5} />
                    </div>
                    <div className="min-w-0">
                      <span className="font-mono text-sm font-bold tracking-wider text-foreground">
                        {test.title.toUpperCase()}
                      </span>
                      <div className="flex items-center gap-4 mt-1 flex-wrap">
                        <span className="font-mono text-[10px] text-muted-foreground">
                          {new Date(test.startTime).toLocaleString()}
                        </span>
                        <span className="flex items-center gap-1.5">
                          <Clock className="h-3 w-3 text-muted-foreground" />
                          <span className="font-mono text-[10px] text-muted-foreground">
                            {Math.floor(test.durationSeconds / 60)}MIN
                          </span>
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Right: status + join */}
                  <div className="flex items-center gap-6 flex-wrap">
                    <div className="flex items-center gap-2">
                      <div className={`h-1.5 w-1.5 rounded-full ${st.bg} ${isLive ? "animate-pulse-glow" : ""}`} />
                      <span className={`font-mono text-[10px] tracking-wider ${st.color} uppercase`}>
                        {st.label}
                      </span>
                    </div>

                    {isLive ? (
                      <button
                        onClick={() => handleJoin(test.id)}
                        disabled={joiningId === test.id}
                        className="flex items-center gap-2 border border-neon-cyan/50 bg-neon-cyan/10 px-5 py-2 font-mono text-[11px] tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_15px_rgba(0,240,255,0.15)] disabled:opacity-50"
                      >
                        {joiningId === test.id ? (
                          <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        ) : (
                          <Swords className="h-3.5 w-3.5" />
                        )}
                        JOIN
                        <ChevronRight className="h-3 w-3" />
                      </button>
                    ) : (
                      <span className="font-mono text-[10px] tracking-wider text-muted-foreground px-4 py-2 border border-panel-border">
                        {status === "upcoming" ? "NOT STARTED" : "ENDED"}
                      </span>
                    )}
                  </div>
                </div>

                {/* Bottom accent line */}
                <div className="h-0.5 bg-panel-border">
                  <div className={`h-full ${isLive ? "bg-neon-cyan/40" : "bg-neon-pink/20"} w-full`} />
                </div>
              </div>
            )
          })}
        </div>

        {/* Bottom info */}
        {tests.length > 0 && (
          <div className="mt-8 flex items-center gap-4">
            <div className="h-px flex-1 bg-panel-border" />
            <div className="flex items-center gap-2">
              <Zap className="h-3 w-3 text-muted-foreground" />
              <span className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">
                {tests.filter(t => getTestStatus(t) === "live").length} LIVE TESTS
              </span>
            </div>
            <div className="h-px flex-1 bg-panel-border" />
          </div>
        )}
      </div>
    </div>
  )
}

// ═══════════════════════════════════════════════
// VIEW 2 — Active Test
// ═══════════════════════════════════════════════

function ActiveTest({ attemptId, testId, onExit }: { attemptId: string; testId: string; onExit: () => void }) {
  const [loading, setLoading] = useState(true)
  const [questions, setQuestions] = useState<Question[]>([])
  const [submissions, setSubmissions] = useState<Submission[]>([])
  const [remainingSeconds, setRemainingSeconds] = useState(0)
  const [currentQ, setCurrentQ] = useState(0)
  const [submitted, setSubmitted] = useState(false)
  const [finalScore, setFinalScore] = useState<number | null>(null)
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  // Code editing state
  const [code, setCode] = useState("")
  const [language, setLanguage] = useState("python")
  const [langTemplates, setLangTemplates] = useState<Record<string, string>>({})
  const [runResults, setRunResults] = useState<RunCaseResult[] | null>(null)
  const [submitResult, setSubmitResult] = useState<{ verdict: string; passedCount: number; totalCount: number } | null>(null)
  const [runLoading, setRunLoading] = useState(false)
  const [submitLoading, setSubmitLoading] = useState(false)
  const [compileError, setCompileError] = useState<string | null>(null)
  const [outputTab, setOutputTab] = useState<"results" | "output">("results")
  const draftSaveRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // ── Anti-cheat hook ──
  const { violationCount, warningMessage, showWarning } = useAntiCheat({
    attemptId,
    testId,
    remainingSeconds,
    onAutoSubmit: () => handleSubmitAttempt(),
    enabled: !loading && !submitted,
  })

  // Fetch attempt data
  const fetchAttempt = useCallback(async () => {
    try {
      const res = await fetch(`${API_URL}/api/arena/attempts/${attemptId}`, { credentials: "include" })
      if (!res.ok) {
        onExit()
        return
      }
      const data = await res.json()
      setQuestions(data.questions || [])
      setSubmissions(data.submissions || [])
      setRemainingSeconds(data.remainingSeconds || 0)

      // If already submitted
      if (data.attempt?.submittedAt && data.attempt.submittedAt !== "0001-01-01T00:00:00Z") {
        setSubmitted(true)
        setFinalScore(data.attempt.score)
      }
    } catch (e) {
      console.error("Failed to fetch attempt:", e)
      onExit()
    } finally {
      setLoading(false)
    }
  }, [attemptId, onExit])

  useEffect(() => {
    fetchAttempt()
    // Fetch language templates for the editor
    async function fetchLangTemplates() {
      try {
        const res = await fetch(`${API_BASE}/api/arena/languages`, { credentials: "include" })
        if (res.ok) {
          const data = await res.json()
          const templates: Record<string, string> = {}
          for (const lang of (data.languages || [])) {
            templates[lang.id] = lang.template || ""
          }
          setLangTemplates(templates)
        }
      } catch (e) {
        console.error("Failed to fetch language templates:", e)
      }
    }
    fetchLangTemplates()
  }, [fetchAttempt])

  // WebSocket connection for backend-driven timer sync
  const wsRef = useRef<WebSocket | null>(null)
  const wsReconnectRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    if (loading || submitted) return

    function connectWS() {
      const ws = new WebSocket(`${WS_BASE}/ws/arena/${attemptId}`)
      wsRef.current = ws

      ws.onopen = () => {
        console.log("[arena ws] connected")
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          switch (msg.type) {
            case "timer_sync":
              // Correct client timer with server value
              setRemainingSeconds(msg.data.remainingSeconds)
              if (msg.data.status === "ended") {
                handleSubmitAttempt()
              }
              break
            case "auto_submit":
              setSubmitted(true)
              setFinalScore(msg.data.score)
              break
            case "session_ended":
              setSubmitted(true)
              break
            case "pong":
              break
          }
        } catch (e) {
          console.error("[arena ws] parse error:", e)
        }
      }

      ws.onclose = () => {
        console.log("[arena ws] disconnected, reconnecting in 3s...")
        if (!submitted) {
          wsReconnectRef.current = setTimeout(connectWS, 3000)
        }
      }

      ws.onerror = (err) => {
        console.error("[arena ws] error:", err)
        ws.close()
      }
    }

    connectWS()

    // Keep-alive ping every 30s
    const pingInterval = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send("ping")
      }
    }, 30000)

    return () => {
      clearInterval(pingInterval)
      if (wsReconnectRef.current) clearTimeout(wsReconnectRef.current)
      wsRef.current?.close()
    }
  }, [loading, submitted, attemptId]) // eslint-disable-line react-hooks/exhaustive-deps

  // Timer countdown (client-side, corrected by WS timer_sync)
  useEffect(() => {
    if (loading || submitted || remainingSeconds <= 0) return

    timerRef.current = setInterval(() => {
      setRemainingSeconds((prev) => {
        if (prev <= 1) {
          // Auto-submit
          clearInterval(timerRef.current!)
          handleSubmitAttempt()
          return 0
        }
        return prev - 1
      })
    }, 1000)

    return () => {
      if (timerRef.current) clearInterval(timerRef.current)
    }
  }, [loading, submitted]) // eslint-disable-line react-hooks/exhaustive-deps

  // Reset code editor when switching questions
  useEffect(() => {
    const q = questions[currentQ]
    if (!q) return
    setRunResults(null)
    setSubmitResult(null)
    setCompileError(null)

    if (q.type === "coding") {
      // Load previously submitted code, or language-specific template
      const existingSub = submissions.find((s) => s.questionId === q.id && s.type === "coding")
      if (existingSub?.code) {
        setCode(existingSub.code)
        setLanguage(existingSub.language || "python")
      } else {
        setLanguage("python")
        setCode(langTemplates["python"] || "")
      }
    }
  }, [currentQ, questions]) // eslint-disable-line react-hooks/exhaustive-deps

  // When language changes, load the appropriate template (unless user has existing submission for this question)
  useEffect(() => {
    const q = questions[currentQ]
    if (!q || q.type !== "coding") return
    const existingSub = submissions.find((s) => s.questionId === q.id && s.type === "coding" && s.language === language)
    if (existingSub?.code) {
      setCode(existingSub.code)
    } else {
      setCode(langTemplates[language] || "")
    }
  }, [language]) // eslint-disable-line react-hooks/exhaustive-deps

  // Auto-save draft on code change (debounced 3s)
  function handleCodeChange(newCode: string) {
    setCode(newCode)
    const q = questions[currentQ]
    if (!q || q.type !== "coding") return

    if (draftSaveRef.current) clearTimeout(draftSaveRef.current)
    draftSaveRef.current = setTimeout(async () => {
      try {
        await fetch(`${API_URL}/api/arena/submissions/draft`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ attemptId, questionId: q.id, code: newCode, language }),
        })
      } catch (e) {
        // Silent fail for draft save
      }
    }, 3000)
  }

  // ── MCQ handler ──
  async function handleMCQSelect(questionId: string, optionId: string) {
    // Optimistic update in local state
    setSubmissions((prev) => {
      const existing = prev.findIndex((s) => s.questionId === questionId)
      if (existing >= 0) {
        const updated = [...prev]
        updated[existing] = { ...updated[existing], selectedOptionId: optionId }
        return updated
      }
      return [...prev, { id: "", attemptId, questionId, type: "mcq", selectedOptionId: optionId, code: "", language: "", verdict: "", passedCount: 0, totalCount: 0 }]
    })

    try {
      await fetch(`${API_URL}/api/arena/submissions/mcq`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ attemptId, questionId, selectedOptionId: optionId }),
      })
    } catch (e) {
      console.error("MCQ save failed:", e)
    }
  }

  // ── Run code handler ──
  async function handleRunCode() {
    const q = questions[currentQ]
    if (!q) return
    setRunLoading(true)
    setRunResults(null)
    setSubmitResult(null)
    setCompileError(null)
    try {
      const res = await fetch(`${API_URL}/api/arena/submissions/run`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ attemptId, questionId: q.id, code, language }),
      })
      if (res.ok) {
        const data = await res.json()
        setRunResults(data.results || [])
        if (data.compileOutput) {
          setCompileError(data.compileOutput)
          setOutputTab("output")
        } else {
          setOutputTab("results")
        }
      }
    } catch (e) {
      console.error("Run failed:", e)
    } finally {
      setRunLoading(false)
    }
  }

  // ── Submit code handler ──
  async function handleSubmitCode() {
    const q = questions[currentQ]
    if (!q) return
    setSubmitLoading(true)
    setSubmitResult(null)
    setRunResults(null)
    setCompileError(null)
    try {
      const res = await fetch(`${API_URL}/api/arena/submissions/code`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ attemptId, questionId: q.id, code, language }),
      })
      if (res.ok) {
        const data = await res.json()
        setSubmitResult({ verdict: data.verdict, passedCount: data.passedCount, totalCount: data.totalCount })
        setRunResults(data.results || [])
        if (data.compileOutput) {
          setCompileError(data.compileOutput)
          setOutputTab("output")
        } else {
          setOutputTab("results")
        }
        // Update submission in local state
        setSubmissions((prev) => {
          const existing = prev.findIndex((s) => s.questionId === q.id)
          const newSub: Submission = { id: "", attemptId, questionId: q.id, type: "coding", selectedOptionId: "", code, language, verdict: data.verdict, passedCount: data.passedCount, totalCount: data.totalCount }
          if (existing >= 0) {
            const updated = [...prev]
            updated[existing] = newSub
            return updated
          }
          return [...prev, newSub]
        })
      }
    } catch (e) {
      console.error("Submit failed:", e)
    } finally {
      setSubmitLoading(false)
    }
  }

  // ── Submit entire attempt ──
  async function handleSubmitAttempt() {
    try {
      const res = await fetch(`${API_URL}/api/arena/attempts/${attemptId}/submit`, {
        method: "POST",
        credentials: "include",
      })
      if (res.ok) {
        const data = await res.json()
        setSubmitted(true)
        setFinalScore(data.score)
      }
    } catch (e) {
      console.error("Submit attempt failed:", e)
    }
  }

  // ── Loading state ──
  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 text-neon-cyan animate-spin" />
        <span className="font-mono text-xs tracking-widest text-neon-cyan">LOADING TEST...</span>
      </div>
    )
  }

  // ── Submitted state ──
  if (submitted) {
    return (
      <div className="relative min-h-[60vh]">
        <div className="absolute inset-0 grid-bg opacity-30" />
        <div className="relative z-10 flex flex-col items-center justify-center min-h-[60vh] gap-6">
          <Trophy className="h-16 w-16 text-neon-yellow animate-pulse-glow" />
          <h2 className="text-3xl font-bold tracking-tight text-foreground">
            TEST <span className="text-neon-cyan text-glow-cyan">COMPLETE</span>
          </h2>
          <div className="flex items-center gap-2 border border-neon-cyan/40 bg-neon-cyan/10 px-8 py-4">
            <span className="font-mono text-[10px] tracking-widest text-muted-foreground mr-4">FINAL SCORE</span>
            <span className="font-mono text-4xl font-bold text-neon-cyan text-glow-cyan">
              {finalScore ?? 0}
            </span>
          </div>
          <button
            onClick={onExit}
            className="mt-4 flex items-center gap-2 border border-panel-border bg-panel-bg/60 px-6 py-2 font-mono text-[11px] tracking-widest text-foreground transition-all hover:border-neon-cyan/40 hover:bg-panel-bg/80"
          >
            BACK TO TESTS
            <ChevronRight className="h-3 w-3" />
          </button>
        </div>
      </div>
    )
  }

  const currentQuestion = questions[currentQ]
  const isWarning = remainingSeconds <= 300 // 5 minutes
  const isDanger = remainingSeconds <= 60

  return (
    <div className="relative min-h-[60vh] flex flex-col">
      <div className="absolute inset-0 grid-bg opacity-30" />

      {/* ── Anti-cheat warning overlay ── */}
      {showWarning && warningMessage && (
        <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/70 backdrop-blur-sm">
          <div className={`relative border-2 px-8 py-6 max-w-lg text-center ${
            violationCount >= 3
              ? "border-neon-pink bg-neon-pink/10 shadow-[0_0_40px_rgba(255,50,100,0.3)]"
              : violationCount === 2
                ? "border-neon-pink/80 bg-neon-pink/5 shadow-[0_0_30px_rgba(255,50,100,0.2)]"
                : "border-neon-yellow/80 bg-neon-yellow/5 shadow-[0_0_30px_rgba(255,220,50,0.2)]"
          }`}>
            <div className="flex items-center justify-center gap-2 mb-3">
              <Shield className={`h-6 w-6 ${
                violationCount >= 2 ? "text-neon-pink animate-pulse" : "text-neon-yellow animate-pulse"
              }`} />
              <span className={`font-mono text-[10px] tracking-[0.3em] uppercase ${
                violationCount >= 2 ? "text-neon-pink" : "text-neon-yellow"
              }`}>
                ANTI-CHEAT SYSTEM
              </span>
            </div>
            <p className={`font-mono text-sm leading-relaxed ${
              violationCount >= 2 ? "text-neon-pink" : "text-neon-yellow"
            }`}>
              {warningMessage}
            </p>
            <div className="mt-4 flex items-center justify-center gap-2">
              <div className={`h-2 w-2 rounded-full ${violationCount >= 1 ? "bg-neon-pink" : "bg-panel-border"}`} />
              <div className={`h-2 w-2 rounded-full ${violationCount >= 2 ? "bg-neon-pink" : "bg-panel-border"}`} />
              <div className={`h-2 w-2 rounded-full ${violationCount >= 3 ? "bg-neon-pink" : "bg-panel-border"}`} />
            </div>
          </div>
        </div>
      )}

      {/* ── Timer bar ── */}
      <div className="relative z-10 border-b border-panel-border bg-panel-bg/80 backdrop-blur-sm">
        <div className="mx-auto max-w-6xl flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3">
            <Swords className="h-4 w-4 text-neon-pink" />
            <span className="font-mono text-[10px] tracking-[0.2em] text-neon-pink uppercase">
              ACTIVE TEST
            </span>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Target className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="font-mono text-xs text-foreground">
                Q{currentQ + 1}/{questions.length}
              </span>
            </div>

            <div className={`flex items-center gap-2 border px-3 py-1 ${
              isDanger ? "border-neon-pink/60 bg-neon-pink/10" :
              isWarning ? "border-neon-yellow/60 bg-neon-yellow/10" :
              "border-neon-cyan/40 bg-neon-cyan/5"
            }`}>
              <Clock className={`h-3.5 w-3.5 ${isDanger ? "text-neon-pink animate-pulse" : isWarning ? "text-neon-yellow" : "text-neon-cyan"}`} />
              <span className={`font-mono text-sm font-bold tracking-wider ${
                isDanger ? "text-neon-pink text-glow-pink" :
                isWarning ? "text-neon-yellow text-glow-yellow" :
                "text-neon-cyan"
              }`}>
                {formatTime(remainingSeconds)}
              </span>
            </div>

            <button
              onClick={handleSubmitAttempt}
              className="flex items-center gap-2 bg-neon-pink/90 hover:bg-neon-pink px-4 py-1.5 font-mono text-[10px] font-bold tracking-widest text-white transition-all"
            >
              FINISH TEST
              <Send className="h-3 w-3" />
            </button>
          </div>
        </div>

        {/* Timer progress bar */}
        <div className="h-1 bg-panel-border">
          <div
            className={`h-full transition-all duration-1000 ease-linear ${
              isDanger ? "bg-neon-pink" : isWarning ? "bg-neon-yellow" : "bg-neon-cyan"
            }`}
            style={{ width: `${Math.max((remainingSeconds / (remainingSeconds + 1)) * 100, 0)}%` }}
          />
        </div>
      </div>

      {/* ── Main content ── */}
      <div className="relative z-10 flex-1 mx-auto max-w-6xl w-full px-4 py-6">
        <div className="flex gap-6">

          {/* ── Question navigation sidebar ── */}
          <div className="flex flex-col gap-2 shrink-0">
            <span className="font-mono text-[9px] tracking-widest text-muted-foreground mb-1 uppercase">
              Questions
            </span>
            {questions.map((q, i) => {
              const isActive = i === currentQ
              const sub = submissions.find((s) => s.questionId === q.id)
              const isAnswered = !!(sub && (sub.selectedOptionId || sub.code))

              return (
                <button
                  key={q.id}
                  onClick={() => setCurrentQ(i)}
                  className={`relative flex items-center justify-center h-10 w-10 font-mono text-xs font-bold border transition-all ${
                    isActive
                      ? "border-neon-cyan bg-neon-cyan/15 text-neon-cyan shadow-[0_0_10px_rgba(0,240,255,0.2)]"
                      : isAnswered
                        ? "border-neon-cyan/30 bg-neon-cyan/5 text-neon-cyan/70"
                        : "border-panel-border bg-panel-bg/40 text-muted-foreground hover:border-panel-border/80"
                  }`}
                >
                  Q{i + 1}
                  {isAnswered && !isActive && (
                    <div className="absolute -top-1 -right-1 h-2 w-2 rounded-full bg-neon-cyan" />
                  )}
                </button>
              )
            })}
          </div>

          {/* ── Question area ── */}
          <div className="flex-1 min-w-0">
            {currentQuestion && (
              <>
                {/* Question header */}
                <div className="flex items-center gap-3 mb-4">
                  <div className="flex items-center gap-2">
                    {currentQuestion.type === "mcq" ? (
                      <Radio className="h-4 w-4 text-neon-cyan" />
                    ) : (
                      <Code2 className="h-4 w-4 text-neon-pink" />
                    )}
                    <span className="font-mono text-[10px] tracking-[0.2em] text-muted-foreground uppercase">
                      {currentQuestion.type === "mcq" ? "MULTIPLE CHOICE" : "CODING CHALLENGE"}
                    </span>
                  </div>
                  <span className="font-mono text-[10px] tracking-wider text-neon-yellow">
                    {currentQuestion.points} PTS
                  </span>
                </div>

                <h3 className="text-lg font-bold tracking-tight text-foreground mb-2">
                  {currentQuestion.title}
                </h3>
                {currentQuestion.description && (
                  <p className="text-sm text-muted-foreground mb-6 whitespace-pre-wrap">
                    {currentQuestion.description}
                  </p>
                )}

                {/* ── MCQ ── */}
                {currentQuestion.type === "mcq" && currentQuestion.mcqOptions && (
                  <div className="grid gap-3 sm:grid-cols-2">
                    {currentQuestion.mcqOptions.map((opt) => {
                      const selectedId = submissions.find((s) => s.questionId === currentQuestion.id)?.selectedOptionId
                      const isSelected = selectedId === opt.id
                      return (
                        <button
                          key={opt.id}
                          onClick={() => handleMCQSelect(currentQuestion.id, opt.id)}
                          className={`p-5 text-left border transition-all ${
                            isSelected
                              ? "border-neon-cyan bg-neon-cyan/10 shadow-[0_0_15px_rgba(0,240,255,0.2)]"
                              : "border-panel-border bg-panel-bg/60 hover:border-neon-cyan/40"
                          }`}
                        >
                          <span className="font-mono text-sm text-foreground">{opt.optionText}</span>
                        </button>
                      )
                    })}
                  </div>
                )}

                {/* ── Coding ── */}
                {currentQuestion.type === "coding" && (
                  <div className="flex flex-col gap-0 h-full">
                    {/* ── Problem Description Panel ── */}
                    <div className="border border-panel-border bg-panel-bg/40 p-4 mb-3">
                      {/* Constraints */}
                      {currentQuestion.codingDetail?.constraints && (
                        <div className="mb-3">
                          <span className="font-mono text-[9px] tracking-widest text-muted-foreground block mb-1 uppercase">
                            CONSTRAINTS
                          </span>
                          <p className="font-mono text-xs text-foreground whitespace-pre-wrap bg-deep-bg/60 p-2 border border-panel-border/50">
                            {currentQuestion.codingDetail.constraints}
                          </p>
                        </div>
                      )}

                      {/* Sample testcases */}
                      {currentQuestion.testCases && currentQuestion.testCases.length > 0 && (
                        <div>
                          <span className="font-mono text-[9px] tracking-widest text-muted-foreground block mb-1 uppercase">
                            SAMPLE CASES
                          </span>
                          <div className="flex flex-col gap-2">
                            {currentQuestion.testCases.map((tc, i) => (
                              <div key={tc.id} className="grid grid-cols-2 gap-3">
                                <div>
                                  <span className="font-mono text-[8px] tracking-wider text-muted-foreground">INPUT {i + 1}</span>
                                  <pre className="font-mono text-xs text-foreground bg-deep-bg/60 p-2 mt-1 border border-panel-border/50 overflow-x-auto">
                                    {tc.input}
                                  </pre>
                                </div>
                                <div>
                                  <span className="font-mono text-[8px] tracking-wider text-muted-foreground">EXPECTED {i + 1}</span>
                                  <pre className="font-mono text-xs text-foreground bg-deep-bg/60 p-2 mt-1 border border-panel-border/50 overflow-x-auto">
                                    {tc.expectedOutput}
                                  </pre>
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>

                    {/* ── Code Editor + Terminal Split ── */}
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-0 flex-1 min-h-0">
                      {/* ── Left: Editor ── */}
                      <div className="flex flex-col border border-panel-border bg-deep-bg/60">
                        {/* Editor toolbar */}
                        <div className="flex items-center justify-between px-3 py-2 border-b border-panel-border bg-panel-bg/60">
                          <div className="flex items-center gap-3">
                            <Code2 className="h-3.5 w-3.5 text-neon-cyan" />
                            <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">EDITOR</span>
                          </div>
                          <div className="flex items-center gap-2">
                            <span className="font-mono text-[8px] tracking-wider text-muted-foreground uppercase">LANG</span>
                            <select
                              value={language}
                              onChange={(e) => setLanguage(e.target.value)}
                              className="bg-deep-bg border border-panel-border text-foreground font-mono text-[11px] px-2 py-1 focus:border-neon-cyan/50 focus:outline-none cursor-pointer"
                            >
                              <option value="python">Python 3</option>
                              <option value="javascript">JavaScript</option>
                              <option value="cpp">C++17</option>
                              <option value="java">Java</option>
                              <option value="go">Go</option>
                            </select>
                          </div>
                        </div>
                        {/* Code textarea */}
                        <textarea
                          value={code}
                          onChange={(e) => handleCodeChange(e.target.value)}
                          spellCheck={false}
                          className="flex-1 min-h-[280px] w-full bg-transparent p-4 font-mono text-sm text-foreground focus:outline-none resize-y leading-6"
                          style={{ fontFamily: "'Geist Mono', 'Fira Code', 'Consolas', monospace", tabSize: 4 }}
                          placeholder="// Write your solution here..."
                          onKeyDown={(e) => {
                            const target = e.target as HTMLTextAreaElement
                            const start = target.selectionStart
                            const end = target.selectionEnd

                            // Tab key inserts spaces instead of changing focus
                            if (e.key === "Tab") {
                              e.preventDefault()
                              const newCode = code.substring(0, start) + "    " + code.substring(end)
                              handleCodeChange(newCode)
                              setTimeout(() => {
                                target.selectionStart = target.selectionEnd = start + 4
                              }, 0)
                              return
                            }

                            // Enter key: auto-indent
                            if (e.key === "Enter") {
                              e.preventDefault()
                              // Find the current line
                              const before = code.substring(0, start)
                              const after = code.substring(end)
                              const currentLine = before.split("\n").pop() || ""
                              // Get leading whitespace of current line
                              const indentMatch = currentLine.match(/^(\s*)/)
                              let indent = indentMatch ? indentMatch[1] : ""
                              // Add extra indent if line ends with { : ( or def/if/for/while/class etc.
                              const trimmedLine = currentLine.trimEnd()
                              if (trimmedLine.endsWith("{") || trimmedLine.endsWith(":") || trimmedLine.endsWith("(") || trimmedLine.endsWith(",")) {
                                indent += "    "
                              }
                              const newCode = before + "\n" + indent + after
                              handleCodeChange(newCode)
                              const newPos = start + 1 + indent.length
                              setTimeout(() => {
                                target.selectionStart = target.selectionEnd = newPos
                              }, 0)
                              return
                            }
                          }}
                        />
                        {/* Action buttons bar */}
                        <div className="flex items-center gap-2 px-3 py-2 border-t border-panel-border bg-panel-bg/60">
                          <button
                            onClick={handleRunCode}
                            disabled={runLoading || !code.trim()}
                            className="flex items-center gap-1.5 border border-neon-cyan/50 bg-neon-cyan/10 px-4 py-1.5 font-mono text-[10px] tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 disabled:opacity-40"
                          >
                            {runLoading ? <Loader2 className="h-3 w-3 animate-spin" /> : <Play className="h-3 w-3" />}
                            RUN
                          </button>
                          <button
                            onClick={handleSubmitCode}
                            disabled={submitLoading || !code.trim()}
                            className="flex items-center gap-1.5 bg-neon-cyan/90 hover:bg-neon-cyan px-4 py-1.5 font-mono text-[10px] font-bold tracking-widest text-deep-bg transition-all disabled:opacity-40"
                          >
                            {submitLoading ? <Loader2 className="h-3 w-3 animate-spin" /> : <Send className="h-3 w-3" />}
                            SUBMIT
                          </button>
                          <div className="ml-auto font-mono text-[9px] text-muted-foreground/60 tracking-wider">
                            {code.split("\n").length} lines
                          </div>
                        </div>
                      </div>

                      {/* ── Right: Terminal / Output Panel ── */}
                      <div className="flex flex-col border border-panel-border border-l-0 lg:border-l-0 bg-deep-bg/60">
                        {/* Terminal tabs */}
                        <div className="flex items-center border-b border-panel-border bg-panel-bg/60">
                          <button
                            onClick={() => setOutputTab("results")}
                            className={`px-4 py-2 font-mono text-[9px] tracking-widest uppercase border-b-2 transition-all ${
                              outputTab === "results"
                                ? "border-neon-cyan text-neon-cyan"
                                : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                          >
                            TEST RESULTS
                          </button>
                          <button
                            onClick={() => setOutputTab("output")}
                            className={`px-4 py-2 font-mono text-[9px] tracking-widest uppercase border-b-2 transition-all ${
                              outputTab === "output"
                                ? "border-neon-cyan text-neon-cyan"
                                : "border-transparent text-muted-foreground hover:text-foreground"
                            }`}
                          >
                            OUTPUT {compileError && <span className="text-neon-pink ml-1">●</span>}
                          </button>
                          {/* Verdict badge inline */}
                          {submitResult && (
                            <div className={`ml-auto mr-3 flex items-center gap-1.5 px-2 py-0.5 border text-[9px] font-mono tracking-wider ${
                              submitResult.verdict === "accepted"
                                ? "border-green-500/50 bg-green-500/10 text-green-400"
                                : "border-neon-pink/50 bg-neon-pink/10 text-neon-pink"
                            }`}>
                              {submitResult.verdict === "accepted" ? (
                                <CheckCircle2 className="h-3 w-3" />
                              ) : (
                                <XCircle className="h-3 w-3" />
                              )}
                              {submitResult.verdict.replace(/_/g, " ").toUpperCase()}
                              <span className="text-muted-foreground ml-1">
                                {submitResult.passedCount}/{submitResult.totalCount}
                              </span>
                            </div>
                          )}
                        </div>

                        {/* Terminal content */}
                        <div className="flex-1 min-h-[280px] overflow-auto p-3">
                          {outputTab === "results" && (
                            <>
                              {runLoading || submitLoading ? (
                                <div className="flex items-center justify-center h-full">
                                  <div className="flex items-center gap-2 text-neon-cyan">
                                    <Loader2 className="h-5 w-5 animate-spin" />
                                    <span className="font-mono text-xs tracking-wider">Executing...</span>
                                  </div>
                                </div>
                              ) : runResults && runResults.length > 0 ? (
                                <div className="flex flex-col gap-2">
                                  {runResults.map((r, i) => (
                                    <div
                                      key={i}
                                      className={`border p-3 ${
                                        r.pass
                                          ? "border-green-500/30 bg-green-500/5"
                                          : "border-neon-pink/30 bg-neon-pink/5"
                                      }`}
                                    >
                                      <div className="flex items-center justify-between mb-2">
                                        <span className="font-mono text-[10px] tracking-wider text-muted-foreground">
                                          {r.hidden ? `HIDDEN CASE ${i + 1}` : `CASE ${i + 1}`}
                                        </span>
                                        <div className="flex items-center gap-2">
                                          {r.durationMs != null && r.durationMs > 0 && (
                                            <span className="font-mono text-[9px] text-muted-foreground/70">{r.durationMs}ms</span>
                                          )}
                                          {r.pass ? (
                                            <span className="flex items-center gap-1 font-mono text-[9px] tracking-wider text-green-400">
                                              <CheckCircle2 className="h-3 w-3" /> PASS
                                            </span>
                                          ) : (
                                            <span className="flex items-center gap-1 font-mono text-[9px] tracking-wider text-neon-pink">
                                              <XCircle className="h-3 w-3" /> {r.errorType ? r.errorType.replace(/_/g, " ").toUpperCase() : "FAIL"}
                                            </span>
                                          )}
                                        </div>
                                      </div>
                                      {!r.hidden && (
                                        <div className="grid grid-cols-3 gap-2">
                                          <div>
                                            <span className="font-mono text-[8px] tracking-wider text-muted-foreground/70 block">INPUT</span>
                                            <pre className="font-mono text-[11px] text-foreground bg-deep-bg/80 p-1.5 mt-0.5 border border-panel-border/30 overflow-x-auto whitespace-pre-wrap max-h-20 overflow-y-auto">{r.input}</pre>
                                          </div>
                                          <div>
                                            <span className="font-mono text-[8px] tracking-wider text-muted-foreground/70 block">EXPECTED</span>
                                            <pre className="font-mono text-[11px] text-foreground bg-deep-bg/80 p-1.5 mt-0.5 border border-panel-border/30 overflow-x-auto whitespace-pre-wrap max-h-20 overflow-y-auto">{r.expected}</pre>
                                          </div>
                                          <div>
                                            <span className="font-mono text-[8px] tracking-wider text-muted-foreground/70 block">ACTUAL</span>
                                            <pre className={`font-mono text-[11px] bg-deep-bg/80 p-1.5 mt-0.5 border border-panel-border/30 overflow-x-auto whitespace-pre-wrap max-h-20 overflow-y-auto ${r.pass ? "text-green-400" : "text-neon-pink"}`}>{r.actual}</pre>
                                          </div>
                                        </div>
                                      )}
                                    </div>
                                  ))}
                                </div>
                              ) : (
                                <div className="flex flex-col items-center justify-center h-full text-muted-foreground/50">
                                  <Play className="h-8 w-8 mb-2" />
                                  <span className="font-mono text-[10px] tracking-wider">Click RUN to test your code</span>
                                </div>
                              )}
                            </>
                          )}

                          {outputTab === "output" && (
                            <div className="font-mono text-xs">
                              {compileError ? (
                                <div>
                                  <div className="flex items-center gap-2 mb-2">
                                    <AlertTriangle className="h-3.5 w-3.5 text-neon-pink" />
                                    <span className="text-[10px] tracking-wider text-neon-pink font-bold uppercase">Compilation Error</span>
                                  </div>
                                  <pre className="text-neon-pink/90 bg-neon-pink/5 border border-neon-pink/20 p-3 whitespace-pre-wrap overflow-auto max-h-60">
                                    {compileError}
                                  </pre>
                                </div>
                              ) : (
                                <div className="flex flex-col items-center justify-center h-full text-muted-foreground/50">
                                  <Code2 className="h-8 w-8 mb-2" />
                                  <span className="text-[10px] tracking-wider">Compilation output will appear here</span>
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

