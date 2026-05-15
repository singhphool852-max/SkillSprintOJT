"use client"

import { useEffect, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import { TrainingSolver } from "@/components/train/training-solver"
import { Loader2, ShieldAlert, ArrowLeft, RefreshCcw, ShieldCheck } from "lucide-react"
import { API_URL } from "@/lib/api-config"
import { ErrorBoundary } from "@/components/train/ErrorBoundary"

// This component calls useSearchParams() and must be wrapped in <Suspense>
// by its parent (page.tsx). All logic is unchanged from the original page.

export default function TrainPlayContent({ id }: { id: string }) {
  const router = useRouter()
  const searchParams = useSearchParams()

  const topic = searchParams.get("topic") || "General"
  const mode = searchParams.get("mode") || "Standard"
  const difficulty = searchParams.get("difficulty") || "Medium"
  const count = parseInt(searchParams.get("count") || "10")

  const [questions, setQuestions] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [offlineStatus, setOfflineStatus] = useState<string | null>(null)
  const [summary, setSummary] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (typeof window === "undefined") return
    const isNotesMode = mode.toUpperCase() === "NOTES_SYNC_MODE"
    const savedSummary = sessionStorage.getItem("skillsprint_notes_summary")

    if (isNotesMode && savedSummary) {
      setSummary(savedSummary)
    } else {
      setSummary(undefined)
      if (!isNotesMode) sessionStorage.removeItem("skillsprint_notes_summary")
    }
  }, [mode])

  const fetchQuestions = async () => {
    if (!id) return
    setLoading(true)
    setOfflineStatus(null)

    try {
      const isUUID = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/.test(id)
      const isRecovery =
        mode.toUpperCase() === "MISTAKES" ||
        mode.toUpperCase() === "RECOVERY MODE" ||
        mode.toUpperCase() === "RECOVERY"

      let apiData: any

      if (isUUID && isRecovery) {
        // PATH C: Special Recovery Path from Attempt ID
        const res = await fetch(`${API_URL}/api/training/adaptive/start`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ mode: "recovery", attemptId: id }),
          credentials: "include",
        })

        if (!res.ok) {
          const errData = await res.json().catch(() => ({}))
          setOfflineStatus(errData.error || `Recovery failed (${res.status})`)
          setQuestions([])
          setLoading(false)
          return
        }
        apiData = await res.json()
      } else if (isUUID) {
        // PATH A: Retrieve existing session by UUID
        const res = await fetch(`${API_URL}/api/train/session/${id}`, {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        })

        if (res.status === 404) {
          setOfflineStatus("Session not found. Please regenerate.")
          setQuestions([])
          setLoading(false)
          return
        }

        if (!res.ok) {
          const errData = await res.json().catch(() => ({}))
          setOfflineStatus(errData.error || `Failed to fetch session (${res.status})`)
          setQuestions([])
          setLoading(false)
          return
        }

        apiData = await res.json()
      } else {
        // PATH B: Create new session from topic name
        const res = await fetch(`${API_URL}/api/train/session`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            topic: id.toLowerCase(),
            count,
            difficulty: difficulty.toLowerCase(),
          }),
          credentials: "include",
        })

        if (!res.ok) {
          const errData = await res.json().catch(() => ({}))
          setOfflineStatus(errData.error || `Failed to create session (${res.status})`)
          setQuestions([])
          setLoading(false)
          return
        }

        apiData = await res.json()
      }

      console.log("[TrainPlay] API Success:", apiData)

      const rawQuestions = apiData.questions || apiData.Questions || []
      const sessionId = apiData.session_id || apiData.sessionId || apiData.SessionID || id

      if (!Array.isArray(rawQuestions) || rawQuestions.length === 0) {
        console.warn("[TrainPlay] No questions found in payload:", apiData)
        setOfflineStatus("Neural vault returned no compatible datasets for this query.")
        setQuestions([])
        setLoading(false)
        return
      }

      const mapped = rawQuestions
        .map((q: any, index: number) => {
          if (!q) {
            console.warn(`[TrainPlay] Skipping null question at index ${index}`)
            return null
          }

          let normalizedOptions = q.options || q.Options
          if (typeof normalizedOptions === "string") {
            try {
              normalizedOptions = JSON.parse(normalizedOptions)
            } catch {
              normalizedOptions = []
            }
          }

          if ((q.type === "mcq" || q.Type === "mcq") && Array.isArray(normalizedOptions)) {
            normalizedOptions = normalizedOptions.map((optText: any, idx: number) => {
              if (typeof optText === "object" && optText !== null) return optText
              return { id: `OPT_${q.id || index}_${idx}`, text: String(optText || "") }
            })
          }

          let normalizedTestCases = q.testCases || q.TestCases
          if (typeof normalizedTestCases === "string") {
            try {
              normalizedTestCases = JSON.parse(normalizedTestCases)
            } catch {
              normalizedTestCases = []
            }
          }

          return {
            id: String(q.id || q.ID || index),
            prompt: q.prompt || q.Prompt || "No prompt provided",
            type: (q.type || q.Type || "mcq").toLowerCase(),
            options: Array.isArray(normalizedOptions) ? normalizedOptions : [],
            explanation: q.explanation || q.Explanation || "",
            starterCode: q.starterCode || q.StarterCode || "",
            constraints: q.constraints || q.Constraints || "",
            testCases: Array.isArray(normalizedTestCases) ? normalizedTestCases : [],
            maxScore: 10,
            _answer: q.answer || q.Answer || "",
            _source: q.source || q.Source || "unknown",
            _sessionId: sessionId,
          }
        })
        .filter(Boolean)

      console.log(`[TrainPlay] Successfully mapped ${mapped.length} questions`)
      setQuestions(mapped)
      setLoading(false)
    } catch (err: any) {
      console.error("[TrainPlay] Error:", err?.message || err)
      setOfflineStatus(err?.message || "Failed to load session")
      setQuestions([])
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchQuestions()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-4 bg-deep-bg">
        <div className="relative">
          <Loader2 className="h-12 w-12 text-neon-cyan animate-spin" />
          <div className="absolute inset-0 h-12 w-12 border-b-2 border-neon-cyan/30 rounded-full animate-pulse" />
        </div>
        <div className="flex flex-col items-center gap-1">
          <span className="font-mono text-[10px] tracking-[0.4em] text-neon-cyan uppercase">
            Initializing Neural Session
          </span>
          <span className="font-mono text-[8px] text-muted-foreground uppercase opacity-40 italic">
            Decrypting logical parameters...
          </span>
        </div>
      </div>
    )
  }

  if (questions.length === 0 && !loading) {
    const isMistakesMode =
      mode.toUpperCase() === "MISTAKES" ||
      mode.toUpperCase() === "RECOVERY MODE" ||
      mode.toUpperCase() === "RECOVERY"

    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-6 bg-deep-bg p-4 relative overflow-hidden">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,_var(--tw-gradient-stops))] from-neon-cyan/5 via-transparent to-transparent animate-pulse" />

        {isMistakesMode ? (
          <div className="flex flex-col items-center gap-6 animate-in fade-in zoom-in duration-500">
            <div className="h-20 w-20 rounded-full border-2 border-neon-cyan/20 flex items-center justify-center relative">
              <ShieldCheck className="h-10 w-10 text-neon-cyan" />
              <div className="absolute inset-0 rounded-full border-2 border-neon-cyan animate-ping opacity-20" />
            </div>
            <div className="text-center max-w-md">
              <h2 className="text-2xl font-bold font-mono text-foreground uppercase tracking-widest mb-2">
                NEURAL VAULT SECURED
              </h2>
              <p className="text-sm text-muted-foreground font-mono uppercase leading-relaxed">
                Great job! You have no pending recovery questions. Your mastery of this topic is currently optimal.
              </p>
            </div>
            <div className="flex flex-col gap-3 w-full max-w-xs">
              <button
                onClick={() => router.push("/train")}
                className="w-full px-6 py-3 border border-panel-border bg-panel-bg/50 text-xs font-mono font-bold tracking-widest text-foreground uppercase hover:bg-white/5 transition-all flex items-center justify-center gap-2 group"
              >
                <ArrowLeft className="h-3 w-3 group-hover:-translate-x-1 transition-transform" /> Return to Training Hub
              </button>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-6">
            <ShieldAlert className="h-12 w-12 text-neon-pink" />
            <div className="text-center">
              <h2 className="text-xl font-bold font-mono text-foreground uppercase tracking-widest">
                SESSION INITIALIZATION FAILED
              </h2>
              <p className="text-xs text-muted-foreground font-mono mt-2 uppercase">
                Neural Vault response was empty or datasets are currently unavailable.
              </p>
            </div>
            <div className="flex flex-col gap-3 w-full max-w-xs">
              <button
                onClick={() => fetchQuestions()}
                className="w-full px-6 py-3 border border-neon-cyan/30 bg-neon-cyan/5 text-xs font-mono font-bold tracking-widest text-neon-cyan uppercase hover:bg-neon-cyan/10 transition-all flex items-center justify-center gap-2"
              >
                <RefreshCcw className="h-3 w-3" /> Retry Initialization
              </button>
              <button
                onClick={() => router.push("/train")}
                className="w-full px-6 py-3 border border-panel-border bg-panel-bg/50 text-xs font-mono font-bold tracking-widest text-muted-foreground uppercase hover:bg-white/5 transition-all flex items-center justify-center gap-2"
              >
                <ArrowLeft className="h-3 w-3" /> Return to Hub
              </button>
            </div>
          </div>
        )}
      </div>
    )
  }

  return (
    <ErrorBoundary>
      <ProtectedRoute>
        <div className="relative">
          {offlineStatus && (
            <div className="fixed top-20 right-8 z-50 animate-pulse">
              <div className="flex items-center gap-2 px-3 py-2 bg-neon-yellow/10 border border-neon-yellow/40 backdrop-blur">
                <ShieldCheck className="h-4 w-4 text-neon-yellow" />
                <span className="font-mono text-[9px] text-neon-yellow uppercase tracking-widest font-black max-w-[200px] truncate">
                  {offlineStatus}
                </span>
              </div>
            </div>
          )}
          <TrainingSolver
            initialQuestions={questions}
            topic={topic + (offlineStatus ? " (Offline)" : "")}
            mode={mode}
            difficulty={difficulty}
            count={offlineStatus ? questions.length : count}
            arenaId={id}
            summary={summary}
          />
        </div>
      </ProtectedRoute>
    </ErrorBoundary>
  )
}
