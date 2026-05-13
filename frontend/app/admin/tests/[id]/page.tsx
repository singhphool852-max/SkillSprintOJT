"use client"

import { useEffect, useState, useCallback } from "react"
import { useParams, useRouter } from "next/navigation"
import Link from "next/link"
import {
  AlertTriangle,
  ArrowLeft,
  Check,
  ChevronRight,
  Code2,
  Eye,
  EyeOff,
  FileText,
  Loader2,
  Plus,
  Radio,
  Shield,
  Trash2,
  X,
} from "lucide-react"
import { API_URL } from "@/lib/api-config"

const API = API_URL

// ─── Types ──────────────────

interface MCQOption {
  id: string
  optionText: string
  isCorrect: boolean
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
  isHidden: boolean
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

interface Test {
  id: string
  title: string
  startTime: string
  durationSeconds: number
  isPublished: boolean
}

// ═══════════════════════════════════════════════
// MAIN PAGE
// ═══════════════════════════════════════════════

export default function AdminTestDetailPage() {
  const params = useParams()
  const router = useRouter()
  const testId = params?.id as string

  const [test, setTest] = useState<Test | null>(null)
  const [questions, setQuestions] = useState<Question[]>([])
  const [loading, setLoading] = useState(true)

  // Question form
  const [showQForm, setShowQForm] = useState(false)
  const [qType, setQType] = useState<"mcq" | "coding">("mcq")
  const [qTitle, setQTitle] = useState("")
  const [qDesc, setQDesc] = useState("")
  const [qPoints, setQPoints] = useState(10)
  const [qPosition, setQPosition] = useState(1)
  const [qCreating, setQCreating] = useState(false)

  // MCQ options
  const [mcqOptions, setMcqOptions] = useState<{ optionText: string; isCorrect: boolean }[]>([
    { optionText: "", isCorrect: true },
    { optionText: "", isCorrect: false },
    { optionText: "", isCorrect: false },
    { optionText: "", isCorrect: false },
  ])

  // Coding fields
  const [constraints, setConstraints] = useState("")
  const [starterCode, setStarterCode] = useState("")
  const [timeLimitMs, setTimeLimitMs] = useState(2000)

  // Testcase form per question
  const [tcQuestionId, setTcQuestionId] = useState<string | null>(null)
  const [tcInput, setTcInput] = useState("")
  const [tcExpected, setTcExpected] = useState("")
  const [tcHidden, setTcHidden] = useState(false)
  const [tcCreating, setTcCreating] = useState(false)

  // ── Delete handlers ──
  const handleDeleteTestcase = async (testcaseId: string) => {
    if (!confirm('Delete this testcase? This action cannot be undone.')) return
    
    try {
      const token = localStorage.getItem('token')
      const res = await fetch(`${API}/api/admin/testcases/${testcaseId}`, {
        method: 'DELETE',
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        credentials: 'include',
      })
      
      if (res.ok) {
        // Remove from local state immediately
        setQuestions(prev => prev.map(q => ({
          ...q,
          testCases: q.testCases?.filter(tc => tc.id !== testcaseId) || null
        })))
      } else {
        alert('Failed to delete testcase')
      }
    } catch (e) {
      console.error('Delete testcase failed:', e)
      alert('Failed to delete testcase')
    }
  }

  const handleDeleteQuestion = async (questionId: string) => {
    if (!confirm('Delete this question and all its testcases? This action cannot be undone.')) return
    
    try {
      const token = localStorage.getItem('token')
      const res = await fetch(`${API}/api/admin/questions/${questionId}`, {
        method: 'DELETE',
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        credentials: 'include',
      })
      
      if (res.ok) {
        // Remove from local state immediately
        setQuestions(prev => prev.filter(q => q.id !== questionId))
      } else {
        alert('Failed to delete question')
      }
    } catch (e) {
      console.error('Delete question failed:', e)
      alert('Failed to delete question')
    }
  }

  const fetchTest = useCallback(async () => {
    try {
      // 1. Fetch test metadata deeply (includes questions, options, details via backend Preload)
      const res = await fetch(`${API}/api/admin/tests/${testId}`, { credentials: "include" })
      if (res.ok) {
        const data: Test & { questions?: Question[] } = await res.json()
        setTest(data)
        if (data.questions) {
          setQuestions(data.questions)
        }
      }
    } catch (e) {
      console.error("Fetch failed:", e)
    } finally {
      setLoading(false)
    }
  }, [testId])

  useEffect(() => {
    if (testId) fetchTest()
  }, [testId, fetchTest])

  // ── Add question ──
  async function handleAddQuestion(e: React.FormEvent) {
    e.preventDefault()
    setQCreating(true)
    try {
      const body: any = {
        type: qType,
        title: qTitle.trim(),
        description: qDesc.trim(),
        points: qPoints,
        position: qPosition,
      }
      if (qType === "mcq") {
        body.options = mcqOptions.filter((o) => o.optionText.trim())
      } else {
        body.constraints = constraints
        body.starterCode = starterCode
        body.timeLimitMs = timeLimitMs
      }

      const res = await fetch(`${API}/api/admin/tests/${testId}/questions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(body),
      })
      if (res.ok) {
        resetQForm()
        setShowQForm(false)
        await fetchTest()
      }
    } catch (e) {
      console.error("Add question failed:", e)
    } finally {
      setQCreating(false)
    }
  }

  function resetQForm() {
    setQTitle("")
    setQDesc("")
    setQPoints(10)
    setQPosition(questions.length + 2)
    setQType("mcq")
    setConstraints("")
    setStarterCode("")
    setTimeLimitMs(2000)
    setMcqOptions([
      { optionText: "", isCorrect: true },
      { optionText: "", isCorrect: false },
      { optionText: "", isCorrect: false },
      { optionText: "", isCorrect: false },
    ])
  }

  // ── Add testcase ──
  async function handleAddTestcase(e: React.FormEvent) {
    e.preventDefault()
    if (!tcQuestionId) return
    setTcCreating(true)
    try {
      const res = await fetch(`${API}/api/admin/questions/${tcQuestionId}/testcases`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          input: tcInput,
          expectedOutput: tcExpected,
          isHidden: tcHidden,
        }),
      })
      if (res.ok) {
        setTcInput("")
        setTcExpected("")
        setTcHidden(false)
        setTcQuestionId(null)
        await fetchTest()
      }
    } catch (e) {
      console.error("Add testcase failed:", e)
    } finally {
      setTcCreating(false)
    }
  }

  // ── MCQ helpers ──
  function addOption() {
    setMcqOptions([...mcqOptions, { optionText: "", isCorrect: false }])
  }

  function removeOption(idx: number) {
    if (mcqOptions.length <= 2) return
    const updated = mcqOptions.filter((_, i) => i !== idx)
    // If the correct one was removed, make the first one correct
    if (!updated.some((o) => o.isCorrect) && updated.length > 0) {
      updated[0].isCorrect = true
    }
    setMcqOptions(updated)
  }

  function setCorrectOption(idx: number) {
    setMcqOptions(mcqOptions.map((o, i) => ({ ...o, isCorrect: i === idx })))
  }

  // ── Loading ──
  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 text-neon-pink animate-spin" />
        <span className="font-mono text-xs tracking-widest text-neon-pink">LOADING TEST...</span>
      </div>
    )
  }

  if (!test) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <AlertTriangle className="h-10 w-10 text-neon-pink" />
        <span className="font-mono text-xs tracking-widest text-neon-pink">TEST NOT FOUND</span>
        <Link href="/admin" className="font-mono text-[10px] tracking-widest text-muted-foreground hover:text-neon-pink transition-colors">
          ← BACK TO TESTS
        </Link>
      </div>
    )
  }

  return (
    <div className="relative min-h-screen">
      <div className="absolute inset-0 grid-bg opacity-20" />

      <div className="relative z-10 px-8 py-8">
        {/* Back + title */}
        <div className="flex items-center gap-3 mb-2">
          <Link href="/admin" className="text-muted-foreground hover:text-neon-pink transition-colors">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <Shield className="h-4 w-4 text-neon-pink" />
          <span className="font-mono text-[10px] tracking-[0.3em] text-neon-pink">
            TEST DETAIL
          </span>
        </div>

        <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl mb-1">
          {test.title.toUpperCase()}
        </h1>

        <div className="flex items-center gap-6 mb-8 flex-wrap">
          <span className="font-mono text-[10px] text-muted-foreground">
            {new Date(test.startTime).toLocaleString()}
          </span>
          <span className="font-mono text-[10px] text-muted-foreground">
            {Math.floor(test.durationSeconds / 60)}MIN
          </span>
          <div className="flex items-center gap-2">
            <div className={`h-1.5 w-1.5 rounded-full ${test.isPublished ? "bg-neon-cyan animate-pulse-glow" : "bg-muted-foreground"}`} />
            <span className={`font-mono text-[10px] tracking-wider ${test.isPublished ? "text-neon-cyan" : "text-muted-foreground"}`}>
              {test.isPublished ? "PUBLISHED" : "DRAFT"}
            </span>
          </div>
        </div>

        {/* Add question button */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <FileText className="h-4 w-4 text-neon-pink" />
            <span className="font-mono text-[10px] tracking-[0.2em] text-neon-pink uppercase">
              QUESTIONS ({questions.length})
            </span>
          </div>
          <button
            onClick={() => { setShowQForm(!showQForm); setQPosition(questions.length + 1) }}
            className="flex items-center gap-2 border border-neon-pink/50 bg-neon-pink/10 px-4 py-2 font-mono text-[10px] tracking-widest text-neon-pink transition-all hover:bg-neon-pink/20"
          >
            <Plus className="h-3 w-3" />
            {showQForm ? "CANCEL" : "ADD QUESTION"}
          </button>
        </div>

        {/* ── Add Question Form ── */}
        {showQForm && (
          <form onSubmit={handleAddQuestion} className="mb-8 border border-neon-pink/30 bg-panel-bg/60 backdrop-blur-sm p-6">
            <div className="flex items-center gap-2 mb-5">
              <Plus className="h-3.5 w-3.5 text-neon-pink" />
              <span className="font-mono text-[10px] tracking-[0.2em] text-neon-pink uppercase">
                NEW QUESTION
              </span>
            </div>

            {/* Type toggle */}
            <div className="flex items-center gap-2 mb-5">
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase mr-2">TYPE</span>
              <button
                type="button"
                onClick={() => setQType("mcq")}
                className={`flex items-center gap-2 px-4 py-2 font-mono text-[10px] tracking-wider border transition-all ${
                  qType === "mcq"
                    ? "border-neon-pink bg-neon-pink/15 text-neon-pink"
                    : "border-panel-border text-muted-foreground hover:border-neon-pink/30"
                }`}
              >
                <Radio className="h-3 w-3" />
                MCQ
              </button>
              <button
                type="button"
                onClick={() => setQType("coding")}
                className={`flex items-center gap-2 px-4 py-2 font-mono text-[10px] tracking-wider border transition-all ${
                  qType === "coding"
                    ? "border-neon-pink bg-neon-pink/15 text-neon-pink"
                    : "border-panel-border text-muted-foreground hover:border-neon-pink/30"
                }`}
              >
                <Code2 className="h-3 w-3" />
                CODING
              </button>
            </div>

            {/* Common fields */}
            <div className="grid gap-5 sm:grid-cols-2 mb-5">
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">TITLE</label>
                <input
                  type="text"
                  value={qTitle}
                  onChange={(e) => setQTitle(e.target.value)}
                  required
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <div className="flex flex-col gap-2">
                  <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">POINTS</label>
                  <input
                    type="number"
                    value={qPoints}
                    onChange={(e) => setQPoints(parseInt(e.target.value) || 0)}
                    min={1}
                    required
                    className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">POSITION</label>
                  <input
                    type="number"
                    value={qPosition}
                    onChange={(e) => setQPosition(parseInt(e.target.value) || 0)}
                    min={1}
                    required
                    className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                  />
                </div>
              </div>
            </div>
            <div className="flex flex-col gap-2 mb-5">
              <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">DESCRIPTION</label>
              <textarea
                value={qDesc}
                onChange={(e) => setQDesc(e.target.value)}
                rows={3}
                className="bg-deep-bg/80 border border-panel-border px-4 py-3 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none resize-y"
              />
            </div>

            {/* ── MCQ Options ── */}
            {qType === "mcq" && (
              <div className="mb-5">
                <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase block mb-3">
                  OPTIONS (click radio to mark correct)
                </span>
                <div className="flex flex-col gap-2">
                  {mcqOptions.map((opt, idx) => (
                    <div key={idx} className="flex items-center gap-3">
                      <button
                        type="button"
                        onClick={() => setCorrectOption(idx)}
                        className={`flex items-center justify-center h-6 w-6 border shrink-0 transition-all ${
                          opt.isCorrect
                            ? "border-green-500 bg-green-500/20 text-green-400"
                            : "border-panel-border text-muted-foreground hover:border-green-500/40"
                        }`}
                      >
                        {opt.isCorrect && <Check className="h-3 w-3" />}
                      </button>
                      <input
                        type="text"
                        value={opt.optionText}
                        onChange={(e) => {
                          const updated = [...mcqOptions]
                          updated[idx].optionText = e.target.value
                          setMcqOptions(updated)
                        }}
                        placeholder={`Option ${idx + 1}`}
                        className="flex-1 bg-deep-bg/80 border border-panel-border px-4 py-2 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                      />
                      <button
                        type="button"
                        onClick={() => removeOption(idx)}
                        className="text-muted-foreground hover:text-neon-pink transition-colors"
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  ))}
                </div>
                <button
                  type="button"
                  onClick={addOption}
                  className="mt-3 flex items-center gap-1.5 font-mono text-[9px] tracking-widest text-muted-foreground hover:text-neon-pink transition-colors"
                >
                  <Plus className="h-3 w-3" /> ADD OPTION
                </button>
              </div>
            )}

            {/* ── Coding Fields ── */}
            {qType === "coding" && (
              <div className="grid gap-5 mb-5">
                <div className="flex flex-col gap-2">
                  <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">CONSTRAINTS</label>
                  <textarea
                    value={constraints}
                    onChange={(e) => setConstraints(e.target.value)}
                    rows={3}
                    className="bg-deep-bg/80 border border-panel-border px-4 py-3 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none resize-y"
                    placeholder="1 <= N <= 10^5&#10;1 <= A[i] <= 10^9"
                  />
                </div>
                <div className="flex flex-col gap-2">
                  <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">STARTER CODE HINT <span className="text-muted-foreground/50">(optional — shown in problem description, per-language templates load automatically)</span></label>
                  <textarea
                    value={starterCode}
                    onChange={(e) => setStarterCode(e.target.value)}
                    rows={5}
                    className="bg-deep-bg/80 border border-panel-border px-4 py-3 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none resize-y"
                    style={{ fontFamily: "'Geist Mono', 'Fira Code', 'Consolas', monospace", tabSize: 4 }}
                    placeholder="Optional: function signature or approach hint for students"
                  />
                </div>
                <div className="flex flex-col gap-2 max-w-xs">
                  <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">TIME LIMIT (MS)</label>
                  <input
                    type="number"
                    value={timeLimitMs}
                    onChange={(e) => setTimeLimitMs(parseInt(e.target.value) || 0)}
                    min={100}
                    className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                  />
                </div>
              </div>
            )}

            <div className="flex justify-end">
              <button
                type="submit"
                disabled={qCreating}
                className="flex items-center gap-2 bg-neon-pink/90 hover:bg-neon-pink px-6 py-2 font-mono text-[11px] font-bold tracking-widest text-white transition-all disabled:opacity-50"
              >
                {qCreating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                ADD QUESTION
              </button>
            </div>
          </form>
        )}

        {/* ── Questions List ── */}
        {questions.length === 0 && !showQForm && (
          <div className="flex flex-col items-center justify-center py-16 gap-4 border border-panel-border bg-panel-bg/20">
            <FileText className="h-10 w-10 text-muted-foreground" />
            <span className="font-mono text-xs tracking-widest text-muted-foreground">
              NO QUESTIONS YET
            </span>
          </div>
        )}

        <div className="flex flex-col gap-4">
          {questions.map((q, qi) => (
            <div key={q.id} className="border border-panel-border bg-panel-bg/40">
              {/* Question header */}
              <div className="flex items-center justify-between px-5 py-3 border-b border-panel-border">
                <div className="flex items-center gap-3">
                  <span className="flex items-center justify-center h-7 w-7 border border-neon-pink/30 bg-neon-pink/10 font-mono text-xs font-bold text-neon-pink">
                    {q.position}
                  </span>
                  {q.type === "mcq" ? (
                    <Radio className="h-3.5 w-3.5 text-neon-cyan" />
                  ) : (
                    <Code2 className="h-3.5 w-3.5 text-neon-pink" />
                  )}
                  <span className="font-mono text-[10px] tracking-wider text-muted-foreground uppercase">
                    {q.type === "mcq" ? "MCQ" : "CODING"}
                  </span>
                  <span className="font-mono text-[10px] tracking-wider text-neon-yellow">
                    {q.points} PTS
                  </span>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleDeleteQuestion(q.id)}
                    className="flex items-center gap-1.5 px-3 py-1.5 font-mono text-[9px] tracking-widest border border-red-500/50 text-red-500 hover:bg-red-500/20 hover:border-red-500 transition-all"
                    title="Delete question"
                  >
                    <Trash2 className="h-3 w-3" />
                    DELETE QUESTION
                  </button>
                  <button
                    onClick={() => setTcQuestionId(tcQuestionId === q.id ? null : q.id)}
                    className="flex items-center gap-1.5 px-3 py-1.5 font-mono text-[9px] tracking-widest border border-panel-border text-muted-foreground hover:border-neon-cyan/40 hover:text-neon-cyan transition-all"
                  >
                    <Plus className="h-3 w-3" />
                    ADD TESTCASE
                  </button>
                </div>
              </div>

              {/* Question body */}
              <div className="px-5 py-4">
                <h3 className="text-sm font-bold tracking-tight text-foreground mb-1">{q.title}</h3>
                {q.description && (
                  <p className="text-xs text-muted-foreground whitespace-pre-wrap mb-3">{q.description}</p>
                )}

                {/* MCQ options */}
                {q.type === "mcq" && q.mcqOptions && (
                  <div className="flex flex-col gap-1.5 mb-3">
                    {q.mcqOptions.map((opt) => (
                      <div
                        key={opt.id}
                        className={`flex items-center gap-3 px-3 py-2 border text-xs font-mono ${
                          opt.isCorrect
                            ? "border-green-500/30 bg-green-500/5 text-green-400"
                            : "border-panel-border text-muted-foreground"
                        }`}
                      >
                        {opt.isCorrect && <Check className="h-3 w-3 text-green-400 shrink-0" />}
                        {opt.optionText}
                      </div>
                    ))}
                  </div>
                )}

                {/* Coding details */}
                {q.type === "coding" && q.codingDetail && (
                  <div className="flex flex-col gap-2 mb-3">
                    {q.codingDetail.constraints && (
                      <div className="border border-panel-border bg-deep-bg/40 p-3">
                        <span className="font-mono text-[8px] tracking-widest text-muted-foreground block mb-1">CONSTRAINTS</span>
                        <pre className="font-mono text-xs text-foreground whitespace-pre-wrap">{q.codingDetail.constraints}</pre>
                      </div>
                    )}
                    <span className="font-mono text-[9px] text-muted-foreground">
                      TIME LIMIT: {q.codingDetail.timeLimitMs}ms
                    </span>
                  </div>
                )}

                {/* Testcases */}
                {q.testCases && q.testCases.length > 0 && (
                  <div>
                    <span className="font-mono text-[9px] tracking-widest text-muted-foreground block mb-2 uppercase">
                      TESTCASES ({q.testCases.length})
                    </span>
                    <div className="flex flex-col gap-1.5">
                      {q.testCases.map((tc, tci) => (
                        <div key={tc.id} className="grid grid-cols-[2fr_2fr_auto_auto] gap-3 items-start border border-panel-border/50 bg-deep-bg/30 px-3 py-2">
                          <div>
                            <span className="font-mono text-[8px] tracking-wider text-muted-foreground">INPUT</span>
                            <pre className="font-mono text-[11px] text-foreground whitespace-pre-wrap mt-0.5">{tc.input}</pre>
                          </div>
                          <div>
                            <span className="font-mono text-[8px] tracking-wider text-muted-foreground">EXPECTED</span>
                            <pre className="font-mono text-[11px] text-foreground whitespace-pre-wrap mt-0.5">{tc.expectedOutput}</pre>
                          </div>
                          <div className="flex items-center gap-1.5 pt-2">
                            {tc.isHidden ? (
                              <span className="flex items-center gap-1 font-mono text-[8px] tracking-wider text-neon-yellow">
                                <EyeOff className="h-3 w-3" /> HIDDEN
                              </span>
                            ) : (
                              <span className="flex items-center gap-1 font-mono text-[8px] tracking-wider text-muted-foreground">
                                <Eye className="h-3 w-3" /> VISIBLE
                              </span>
                            )}
                          </div>
                          <div className="flex items-center pt-2">
                            <button
                              onClick={() => handleDeleteTestcase(tc.id)}
                              className="text-red-500 hover:text-red-700 transition-colors"
                              title="Delete testcase"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* ── Add Testcase Form (inline) ── */}
              {tcQuestionId === q.id && (
                <form onSubmit={handleAddTestcase} className="px-5 py-4 border-t border-neon-cyan/20 bg-neon-cyan/5">
                  <div className="flex items-center gap-2 mb-4">
                    <Plus className="h-3 w-3 text-neon-cyan" />
                    <span className="font-mono text-[9px] tracking-[0.2em] text-neon-cyan uppercase">
                      NEW TESTCASE
                    </span>
                  </div>

                  <div className="grid gap-4 sm:grid-cols-2 mb-4">
                    <div className="flex flex-col gap-2">
                      <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">INPUT</label>
                      <textarea
                        value={tcInput}
                        onChange={(e) => setTcInput(e.target.value)}
                        rows={3}
                        className="bg-deep-bg/80 border border-panel-border px-4 py-3 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none resize-y"
                      />
                    </div>
                    <div className="flex flex-col gap-2">
                      <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">EXPECTED OUTPUT</label>
                      <textarea
                        value={tcExpected}
                        onChange={(e) => setTcExpected(e.target.value)}
                        rows={3}
                        className="bg-deep-bg/80 border border-panel-border px-4 py-3 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none resize-y"
                      />
                    </div>
                  </div>

                  <div className="flex items-center gap-3 mb-4">
                    <button
                      type="button"
                      onClick={() => setTcHidden(!tcHidden)}
                      className={`flex items-center justify-center h-5 w-5 border transition-all ${
                        tcHidden
                          ? "border-neon-yellow bg-neon-yellow/20"
                          : "border-panel-border"
                      }`}
                    >
                      {tcHidden && <Check className="h-3 w-3 text-neon-yellow" />}
                    </button>
                    <div className="flex flex-col">
                      <span className="font-mono text-[10px] tracking-wider text-foreground">
                        Hidden
                      </span>
                      <span className="font-mono text-[8px] tracking-wider text-muted-foreground">
                        Used for judging only, never shown to users
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <button
                      type="submit"
                      disabled={tcCreating}
                      className="flex items-center gap-2 bg-neon-cyan/90 hover:bg-neon-cyan px-5 py-2 font-mono text-[11px] font-bold tracking-widest text-deep-bg transition-all disabled:opacity-50"
                    >
                      {tcCreating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                      ADD TESTCASE
                    </button>
                    <button
                      type="button"
                      onClick={() => setTcQuestionId(null)}
                      className="px-4 py-2 font-mono text-[10px] tracking-widest text-muted-foreground hover:text-foreground transition-colors"
                    >
                      CANCEL
                    </button>
                  </div>
                </form>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
