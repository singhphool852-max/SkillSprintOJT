"use client"

import { useEffect, useState, useCallback, useMemo } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  ChevronRight,
  Clock,
  Shield,
  Swords,
  Target,
  Trophy,
  Zap,
  Loader2,
} from "lucide-react"
import { API_URL } from "@/lib/api-config"

interface Question {
  id: string;
  prompt: string;
  type: "mcq" | "subjective";
  options?: Array<{ id: string; text: string }>;
  maxScore: number;
}

interface Quiz {
  id: string;
  title: string;
}

export function LiveArena({ arenaId }: { arenaId: string }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const mode = searchParams.get("mode") || "STANDARD"
  
  // Define timer based on mode
  const defaultTime = useMemo(() => {
    const m = mode.toUpperCase()
    if (m.includes("SPEED")) return 10
    if (m.includes("PRACTICE")) return 60 // Relaxed
    return 30 // Standard
  }, [mode])

  const [loading, setLoading] = useState(true)
  const [quiz, setQuiz] = useState<Quiz | null>(null)
  const [questions, setQuestions] = useState<Question[]>([])
  const [currentQ, setCurrentQ] = useState(0)
  const [selectedOptionId, setSelectedOptionId] = useState<string | null>(null)
  const [writtenAnswer, setWrittenAnswer] = useState("")
  const [evaluating, setEvaluating] = useState(false)
  const [timeLeft, setTimeLeft] = useState(defaultTime)
  const [totalScore, setTotalScore] = useState(0)
  const [answers, setAnswers] = useState<any[]>([])
  const [finished, setFinished] = useState(false)
  const [startedAt] = useState(new Date())

  const fetchQuizData = useCallback(async () => {
    try {
      // 1. Get Quizzes
      const quizRes = await fetch(`${API_URL}/api/arenas/${arenaId}/quizzes`)
      if (!quizRes.ok) throw new Error("Failed to fetch quiz")
      const quizzes = await quizRes.json()
      if (quizzes.length === 0) throw new Error("No active quizzes in this arena")
      
      const activeQuiz = quizzes[0]
      setQuiz(activeQuiz)

      // 2. Get Questions
      const qRes = await fetch(`${API_URL}/api/quizzes/${activeQuiz.id}/questions`)
      if (!qRes.ok) throw new Error("Failed to fetch questions")
      const qs = await qRes.json()
      setQuestions(qs)
    } catch (err) {
      console.error(err)
      router.push("/arena")
    } finally {
      setLoading(false)
    }
  }, [arenaId, router])

  useEffect(() => {
    fetchQuizData()
  }, [fetchQuizData])

  const currentQuestion = questions[currentQ]

  const handleNext = useCallback(async () => {
    // Collect current answer
    const currentAnswer = {
      questionId: currentQuestion.id,
      selectedOptionId: selectedOptionId || "",
      writtenAnswer: writtenAnswer || "",
    }
    
    setAnswers(prev => [...prev, currentAnswer])

    if (currentQ < questions.length - 1) {
      setCurrentQ(prev => prev + 1)
      setSelectedOptionId(null)
      setWrittenAnswer("")
      setTimeLeft(defaultTime)
    } else {
      setEvaluating(true)
      // Submit Full Attempt
      try {
        const payload = {
          quizId: quiz?.id,
          startedAt: startedAt.toISOString(),
          answers: [...answers, currentAnswer]
        }
        
        const res = await fetch(`${API_URL}/api/attempts`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
          credentials: "include"
        })
        
        if (res.ok) {
          const result = await res.json()
          router.push(`/results/${result.attemptId}`)
        } else {
          router.push("/arena")
        }
      } catch (err) {
        console.error(err)
        router.push("/arena")
      }
    }
  }, [currentQ, questions, currentQuestion, selectedOptionId, writtenAnswer, answers, quiz, startedAt, router])

  useEffect(() => {
    if (loading || finished || evaluating) return
    if (timeLeft <= 0) {
      handleNext()
      return
    }
    const timer = setInterval(() => setTimeLeft(t => t - 1), 1000)
    return () => clearInterval(timer)
  }, [timeLeft, loading, finished, evaluating, handleNext])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 text-neon-cyan animate-spin" />
        <span className="font-mono text-xs tracking-widest text-neon-cyan">INITIALIZING BATTLEFIELD...</span>
      </div>
    )
  }

  if (evaluating) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Zap className="h-8 w-8 text-neon-pink animate-pulse" />
        <span className="font-mono text-xs tracking-widest text-neon-pink text-glow-pink">AI EVALUATING RESULTS...</span>
      </div>
    )
  }

  const timerPercent = (timeLeft / defaultTime) * 100
  const threshold = defaultTime / 2
  const danger = defaultTime / 4
  const timerColor = timeLeft > threshold ? "text-neon-cyan" : timeLeft > danger ? "text-neon-amber" : "text-neon-pink"

  return (
    <div className="relative min-h-screen flex flex-col">
      <div className="absolute inset-0 grid-bg opacity-30" />
      
      {/* Top bar */}
      <div className="relative z-10 border-b border-panel-border bg-panel-bg/80 backdrop-blur-sm">
        <div className="mx-auto max-w-5xl flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3">
            <Swords className="h-4 w-4 text-neon-pink" />
            <div className="flex flex-col">
              <span className="font-mono text-[10px] tracking-[0.2em] text-neon-pink uppercase leading-tight">
                {quiz?.title || "ARENA"}
              </span>
              <span className="font-mono text-[8px] text-muted-foreground uppercase tracking-widest">
                MODE: {mode}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Target className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="font-mono text-xs text-foreground">
                {currentQ + 1}/{questions.length}
              </span>
            </div>
          </div>
        </div>

        <div className="h-1 bg-panel-border">
          <div
            className={`h-full ${timeLeft > 15 ? 'bg-neon-cyan' : timeLeft > 7 ? 'bg-neon-amber' : 'bg-neon-pink'} transition-all duration-1000 ease-linear`}
            style={{ width: `${timerPercent}%` }}
          />
        </div>
      </div>

      {/* Question area */}
      <div className="relative z-10 flex-1 flex flex-col items-center justify-center px-4 py-8">
        <div className="w-full max-w-2xl">
          <div className="flex items-center justify-center mb-6">
            <div className={`flex h-16 w-16 items-center justify-center border-2 ${timeLeft > 15 ? 'border-neon-cyan/40' : 'border-neon-pink/40'}`}>
              <span className={`font-mono text-2xl font-bold ${timerColor}`}>{timeLeft}</span>
            </div>
          </div>

          <h2 className="text-center text-xl font-bold tracking-tight text-foreground sm:text-2xl mb-10">
            {currentQuestion.prompt}
          </h2>

          {currentQuestion.type === "mcq" ? (
            <div className="grid gap-3 sm:grid-cols-2">
              {currentQuestion.options?.map((opt) => (
                <button
                  key={opt.id}
                  onClick={() => setSelectedOptionId(opt.id)}
                  className={`p-5 text-left border transition-all ${
                    selectedOptionId === opt.id 
                      ? "border-neon-cyan bg-neon-cyan/10 shadow-[0_0_15px_rgba(0,240,255,0.2)]" 
                      : "border-panel-border bg-panel-bg/60 hover:border-neon-cyan/40"
                  }`}
                >
                  <span className="font-mono text-sm text-foreground">{opt.text}</span>
                </button>
              ))}
            </div>
          ) : (
            <div className="flex flex-col gap-4">
              <textarea
                value={writtenAnswer}
                onChange={(e) => setWrittenAnswer(e.target.value)}
                placeholder="Type your answer here..."
                className="w-full h-40 bg-deep-bg/60 border border-panel-border p-4 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none"
              />
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground">AI GRADING ENABLED FOR THIS QUESTION</span>
            </div>
          )}

          <div className="mt-10 flex justify-center">
            <button
              onClick={handleNext}
              className="flex items-center gap-2 bg-neon-cyan/90 hover:bg-neon-cyan px-8 py-3 font-mono text-xs font-bold tracking-widest text-deep-bg transition-all"
            >
              {currentQ === questions.length - 1 ? "FINISH BATTLE" : "NEXT QUESTION"}
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

