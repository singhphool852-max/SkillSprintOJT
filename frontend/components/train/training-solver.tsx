"use client"

import { useState, useEffect, useCallback, useMemo } from "react"
import { useRouter } from "next/navigation"
import { 
  ChevronRight, 
  ChevronLeft,
  Clock, 
  Target, 
  Zap, 
  Loader2, 
  Brain,
  Info,
  ShieldAlert,
  Database
} from "lucide-react"
import { saveSessionMetrics } from "@/lib/training-history"
import { QuestionRenderer } from "./QuestionRenderer"
import { API_URL } from "@/lib/api-config"

interface Question {
  id: string
  prompt: string
  type: string
  options?: { id: string, text: string }[]
  template?: string
  hint?: string
  maxScore: number
  _source?: string
}

interface TrainingSolverProps {
  initialQuestions: Question[]
  topic: string
  mode: string
  difficulty: string
  count: number
  arenaId: string
}

export function TrainingSolver({ initialQuestions, topic, mode, difficulty, count, arenaId }: TrainingSolverProps) {
  const router = useRouter()
  const [currentQ, setCurrentQ] = useState(0)
  const [answers, setAnswers] = useState<Record<string, string>>({})
  const [evaluations, setEvaluations] = useState<Record<string, any>>({})
  const [questionLogs, setQuestionLogs] = useState<any[]>([])
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [evaluating, setEvaluating] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [submissionError, setSubmissionError] = useState<string | null>(null)
  const [startedAt] = useState(new Date())
  const [lastQuestionSync, setLastQuestionSync] = useState(new Date())

  const defaultTime = useMemo(() => {
    const m = mode.toUpperCase()
    if (m.includes("SPEED")) return 15
    if (m.includes("PRACTICE")) return 90
    return 45
  }, [mode])

  const [timeLeft, setTimeLeft] = useState(defaultTime)

  const currentQuestion = initialQuestions[currentQ]
  const currentEvaluation = evaluations[currentQuestion?.id]
  const isAnswered = !!answers[currentQuestion?.id]

  const handleSubmitAnswer = async () => {
    // 1. Pre-flight: check answer exists
    if (!isAnswered) {
      setValidationError("Please provide an input before submitting.")
      return
    }
    if (isSubmitting) return

    // 2. Pre-flight: validate required payload fields
    if (!currentQuestion?.id) {
      setValidationError("Question data is incomplete. Cannot verify.")
      console.warn("[Verify] Blocked: currentQuestion.id is missing")
      return
    }

    setValidationError(null)
    setIsSubmitting(true)

    const userAnswer = answers[currentQuestion.id] || ""
    const isMCQ = currentQuestion.type?.toLowerCase() === "mcq"
    
    // Resolve MCQ ID to actual text for backend consistency
    let finalAnswerValue = userAnswer
    if (isMCQ && currentQuestion.options) {
      const selected = currentQuestion.options.find(o => o.id === userAnswer)
      if (selected) finalAnswerValue = selected.text
    }

    const payload = {
      questionId: currentQuestion.id,
      selectedOptionId: isMCQ ? finalAnswerValue : "",
      writtenAnswer: !isMCQ ? finalAnswerValue : ""
    }

    console.log("[Verify] Selected:", userAnswer, "->", finalAnswerValue)
    console.log("[Verify] Payload Sent:", payload)

    // 3. Check if this is a DB-backed training question (from POST /api/train/session)
    //    These have _answer metadata embedded by the play page mapper
    const dbAnswer = (currentQuestion as any)._answer
    if (dbAnswer !== undefined) {
      // Helper: normalize answer text for reliable comparison
      const normalize = (s: string) => s.trim().toLowerCase().replace(/\s+/g, ' ')

      let isCorrect = false
      let correctOptionId: string | undefined

      if (isMCQ && currentQuestion.options) {
        // userAnswer is the option ID (e.g. "OPT_13_2")
        // dbAnswer is the option text (e.g. "3NF")
        // Find which option the user selected
        const selectedOpt = currentQuestion.options.find((o: any) => o.id === userAnswer)
        const selectedText = selectedOpt?.text || ""

        // Find the correct option by normalized text match
        const correctOpt = currentQuestion.options.find(
          (o: any) => normalize(o.text) === normalize(dbAnswer)
        )
        correctOptionId = correctOpt?.id

        // Compare the selected option text against the correct answer text
        isCorrect = normalize(selectedText) === normalize(dbAnswer)

        console.log("[Verify] question id:", currentQuestion.id)
        console.log("[Verify] raw user answer:", selectedText)
        console.log("[Verify] raw correct answer:", dbAnswer)
        console.log("[Verify] normalized user:", normalize(selectedText))
        console.log("[Verify] normalized correct:", normalize(dbAnswer))
        console.log("[Verify] result:", isCorrect ? "CORRECT" : "INCORRECT")
      } else {
        // For non-MCQ, do a basic normalized comparison
        isCorrect = normalize(userAnswer).includes(normalize(dbAnswer).substring(0, 30))

        console.log("[Verify] question id:", currentQuestion.id)
        console.log("[Verify] raw user answer:", userAnswer.substring(0, 50))
        console.log("[Verify] raw correct answer:", dbAnswer.substring(0, 50))
        console.log("[Verify] result:", isCorrect ? "CORRECT" : "INCORRECT")
      }

      const localResult = {
        isCorrect,
        feedback: isCorrect
          ? "Correct! Well done."
          : "Incorrect. Review the explanation below.",
        explanation: (currentQuestion as any).explanation || "",
        correctOptionId
      }
      setEvaluations(prev => ({ ...prev, [currentQuestion.id]: localResult }))
      setIsSubmitting(false)
      return
    }

    // 4. Attempt backend verification
    console.log("[Verify] Sending payload:", JSON.stringify(payload))
    try {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 12000) // 12s timeout
      const res = await fetch(`${API_URL}/api/training/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        credentials: "include",
        signal: controller.signal
      })
      clearTimeout(timeout)

      console.log("[Verify] Response status:", res.status)

      // 5. Parse response body safely
      let body: any = null
      try {
        const text = await res.text()
        if (text && text.trim().length > 0) {
          body = JSON.parse(text)
        }
      } catch (parseErr) {
        console.warn("[Verify] Failed to parse response body:", parseErr)
      }

      // 6. Classify response
      if (res.ok && body) {
        // Validate the shape minimally
        const evaluation = {
          isCorrect: !!body.isCorrect,
          feedback: body.feedback || "Verification complete.",
          explanation: body.explanation || "",
          correctOptionId: body.correctOptionId || undefined
        }
        setEvaluations(prev => ({ ...prev, [currentQuestion.id]: evaluation }))
        console.log("[Verify] Success. Result:", evaluation.isCorrect ? "CORRECT" : "INCORRECT")
        console.log("[Verify] Backend Body:", body)
        return
      }

      // Non-OK responses — register a fallback evaluation so they can proceed!
      const fallbackEval = {
        isCorrect: false,
        feedback: "System verification unavailable. Answer recorded offline.",
        explanation: "Backend unreachable or unidentifiable status."
      }
      setEvaluations(prev => ({ ...prev, [currentQuestion.id]: fallbackEval }))

      if (res.status === 401 || res.status === 403) {
        setValidationError("Session expired. Local validation used.")
        console.warn("[Verify] Auth failure:", res.status)
      } else if (res.status === 404) {
        setValidationError("Question not found on server. Local validation used.")
        console.warn("[Verify] Question not found in DB:", currentQuestion.id)
      } else if (res.status >= 500) {
        setValidationError("Server error. Local validation used.")
        console.error("[Verify] Server error:", res.status, body)
      } else {
        setValidationError("Could not verify answer right now. Local validation used.")
        console.warn("[Verify] Unexpected status:", res.status, body)
      }

    } catch (err: any) {
      // 7. Network / timeout / abort errors
      const fallbackEval = {
        isCorrect: false,
        feedback: "Server offline or timed out. Answer recorded locally.",
        explanation: "Proceed to next question. Accuracy tracking is paused."
      }
      setEvaluations(prev => ({ ...prev, [currentQuestion.id]: fallbackEval }))

      if (err?.name === "AbortError") {
        setValidationError("Verification timed out. Answer saved.")
        console.warn("[Verify] Request timed out after 12s")
      } else {
        setValidationError("Server unavailable. Answer saved.")
        console.error("[Verify] Network error:", err?.message || err)
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const handlePrev = useCallback(() => {
    setValidationError(null)
    if (currentQ > 0) {
      setCurrentQ(prev => prev - 1)
      setTimeLeft(defaultTime) // Keep timer consistent per question entry
    }
  }, [currentQ, defaultTime])

  const handleNext = useCallback(async () => {
    setValidationError(null)
    const timeTaken = Math.round((new Date().getTime() - lastQuestionSync.getTime()) / 1000)
    const log = {
      id: currentQuestion?.id,
      timeTaken,
      isCorrect: evaluations[currentQuestion?.id]?.isCorrect || false
    }
    setQuestionLogs(prev => [...prev, log])
    setLastQuestionSync(new Date())

    if (currentQ < initialQuestions.length - 1) {
      setCurrentQ(prev => prev + 1)
      setTimeLeft(defaultTime)
    } else {
      if (evaluating) return
      setEvaluating(true)
      try {
        const totalDuration = Math.round((new Date().getTime() - startedAt.getTime()) / 1000)
        const correctCount = initialQuestions.filter(q => evaluations[q.id]?.isCorrect).length
        const accuracy = Math.round((correctCount / initialQuestions.length) * 100)
        
        const payload = {
          quizId: arenaId,
          startedAt: startedAt.toISOString(),
          totalTime: totalDuration,
          accuracy,
          answers: initialQuestions.map(q => ({
            questionId: q.id,
            writtenAnswer: answers[q.id] || "",
            selectedOptionId: q.type === "mcq" ? answers[q.id] : ""
          }))
        }

        // Build a matching result shape to be used as a fallback if the backend is down
        const localFallbackResult = {
          attempt: {
            id: "LOCAL_FALLBACK",
            score: correctCount * 10,
            totalQuestions: initialQuestions.length,
            startedAt: startedAt.toISOString(),
            completedAt: new Date().toISOString(),
            quiz: { title: topic, arenaId: arenaId }
          },
          answers: initialQuestions.map(q => ({
            id: "ANS_" + q.id,
            questionId: q.id,
            isCorrect: evaluations[q.id]?.isCorrect || false,
            score: evaluations[q.id]?.isCorrect ? 10 : 0,
            feedback: evaluations[q.id]?.feedback || "Offline verification recorded.",
            explanation: evaluations[q.id]?.explanation || "",
            evaluatedBy: "LOCAL_SYSTEM",
            writtenAnswer: answers[q.id] || ""
          }))
        }

        // Store for adaptive AI recommendations
        saveSessionMetrics({
          topic,
          difficulty: difficulty || "Medium",
          accuracy,
          timeTaken: totalDuration,
          date: new Date().toISOString()
        })

        sessionStorage.setItem("skillsprint_train_result", JSON.stringify(localFallbackResult))
        
        // Fire and forget backend sync! Instantly navigate so the user isn't blocked.
        fetch(`${API_URL}/api/attempts`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
          credentials: "include"
        }).catch(err => console.warn("Background sync failed:", err))
        
        router.push('/results')
      } catch (err) {
        console.error("Critical submission error, backend unreachable:", err)
        router.push('/results')
      }
    }
  }, [currentQ, initialQuestions, answers, evaluations, arenaId, startedAt, lastQuestionSync, defaultTime, evaluating, router])

  useEffect(() => {
    if (evaluating || currentEvaluation) return
    if (timeLeft <= 0) {
      handleNext()
      return
    }
    const timer = setInterval(() => setTimeLeft(t => t - 1), 1000)
    return () => clearInterval(timer)
  }, [timeLeft, evaluating, currentEvaluation, handleNext])

  if (!currentQuestion) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-6 bg-deep-bg p-4 text-center">
        <ShieldAlert className="h-16 w-16 text-neon-pink animate-pulse" />
        <div className="space-y-2">
          <h2 className="text-2xl font-black italic tracking-widest text-foreground uppercase text-glow-pink">SYSTEM ANOMALY</h2>
          <p className="font-mono text-xs text-muted-foreground uppercase tracking-widest">Neural query undefined. End the session.</p>
        </div>
        <button 
          onClick={() => router.push("/train")}
          className="border border-panel-border px-8 py-3 font-mono text-xs font-bold tracking-widest text-muted-foreground uppercase hover:bg-white/5 transition-all mt-4"
        >
          Return to Hub
        </button>
      </div>
    )
  }

  const progressPercent = ((currentQ + 1) / initialQuestions.length) * 100
  const timerPercent = (timeLeft / defaultTime) * 100

  return (
    <div className="relative min-h-screen flex flex-col bg-deep-bg">
      <div className="absolute inset-0 grid-bg opacity-20 pointer-events-none" />

      {/* Top Header */}
      <div className="relative z-10 border-b border-panel-border bg-panel-bg/80 backdrop-blur-md">
        <div className="mx-auto max-w-7xl flex items-center justify-between px-4 py-4">
          <div className="flex items-center gap-4">
            <div className="flex h-10 w-10 items-center justify-center border border-neon-cyan/40 bg-neon-cyan/5">
               <Brain className="h-5 w-5 text-neon-cyan" />
            </div>
            <div className="flex flex-col">
              <span className="font-mono text-[10px] tracking-[0.2em] text-neon-cyan uppercase leading-tight">
                {topic} // PRACTICE SESSION
              </span>
              <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-widest">
                {mode.replace('_', ' ')} — {difficulty}
              </span>
            </div>
          </div>

          <div className="flex items-center gap-8">
            <div className="flex items-center gap-2">
              <Target className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="font-mono text-xs text-foreground uppercase tracking-tighter">
                QUESTION {currentQ + 1} <span className="text-muted-foreground">OF</span> {initialQuestions.length}
              </span>
            </div>
            
            <div className={`flex items-center gap-3 px-4 py-2 border ${timeLeft < 10 ? 'border-neon-pink/40 bg-neon-pink/5' : 'border-panel-border'}`}>
               <Clock className={`h-4 w-4 ${timeLeft < 10 ? 'text-neon-pink animate-pulse' : 'text-muted-foreground'}`} />
               <span className={`font-mono text-sm font-bold ${timeLeft < 10 ? 'text-neon-pink' : 'text-foreground'}`}>
                 {timeLeft < 10 ? `0:0${timeLeft}` : `0:${timeLeft}`}
               </span>
            </div>
          </div>
        </div>

        {/* Global Progress Bar */}
        <div className="h-[2px] bg-panel-border overflow-hidden">
          <div 
            className="h-full bg-neon-cyan transition-all duration-1000 ease-out shadow-[0_0_10px_#00e5ff]"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
      </div>

      {/* Main content */}
      <div className="relative z-10 flex-1 flex flex-col items-center py-12 px-4 pb-32">
        {evaluating ? (
          <div className="flex flex-col items-center justify-center min-h-[50vh] gap-6 text-center animate-in fade-in duration-500">
             <div className="relative h-20 w-20">
                <Loader2 className="absolute inset-0 h-20 w-20 text-neon-cyan animate-spin opacity-20" />
                <Zap className="absolute inset-5 h-10 w-10 text-neon-pink animate-pulse" />
             </div>
             <div className="space-y-2">
                <h2 className="text-2xl font-black italic tracking-widest text-foreground uppercase shadow-neon-glow">EVALUATING ANSWER</h2>
                <p className="font-mono text-xs text-muted-foreground uppercase tracking-widest">Grading response integrity...</p>
             </div>
          </div>
        ) : (
          <div className="w-full max-w-4xl animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Question Card */}
            <div className="mb-8">
               <div className="flex items-center gap-3 mb-4">
                  <span className="font-mono text-[9px] px-2 py-1 bg-neon-cyan/10 text-neon-cyan border border-neon-cyan/20 uppercase tracking-widest">
                    {currentQuestion.type.replace("_", " ")}
                  </span>
                  {currentQuestion._source === "ai" ? (
                    <span className="font-mono text-[9px] px-2 py-1 bg-neon-pink/10 text-neon-pink border border-neon-pink/20 uppercase tracking-widest flex items-center gap-1.5">
                      <Zap className="h-2 w-2 fill-current" /> AI SYNTHESIZED
                    </span>
                  ) : (
                    <span className="font-mono text-[9px] px-2 py-1 bg-white/5 text-muted-foreground border border-white/10 uppercase tracking-widest flex items-center gap-1.5">
                      <Database className="h-2 w-2" /> VAULT SOURCE
                    </span>
                  )}
                  <div className="h-px flex-1 bg-panel-border/30" />
               </div>
               <h2 className="text-2xl font-bold tracking-tight text-foreground leading-tight mb-8">
                 {currentQuestion.prompt}
               </h2>
            </div>

            {/* Renderer */}
            <div className="mb-12">
               <QuestionRenderer 
                 question={currentQuestion} 
                 answer={answers[currentQuestion.id] || ""}
                 isLocked={!!currentEvaluation}
                 result={currentEvaluation}
                 onChange={(val) => setAnswers(prev => ({ ...prev, [currentQuestion.id]: val }))}
               />
            </div>

            {/* Bottom Controls */}
            <div className="flex flex-col gap-4 pt-8 border-t border-panel-border/30">
               {validationError && (
                 <div className="flex items-center gap-3 p-3 bg-neon-pink/10 border border-neon-pink/30 animate-in zoom-in-95 duration-200">
                   <Info className="h-4 w-4 text-neon-pink shrink-0" />
                   <span className="font-mono text-[10px] text-neon-pink font-bold uppercase tracking-widest leading-none mt-0.5">
                     {validationError}
                   </span>
                 </div>
               )}

               <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                     <div className={`h-1.5 w-1.5 rounded-full ${isAnswered ? 'bg-neon-cyan animate-pulse' : 'bg-muted-foreground opacity-20'}`} />
                     <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-tighter">
                       {isAnswered ? 'Answer selected' : 'Waiting for input...'}
                     </span>
                  </div>
                  
                  <div className="flex items-center gap-4">
                     {currentQ > 0 && (
                       <button
                         onClick={handlePrev}
                         disabled={evaluating}
                         className="group relative flex items-center gap-2 border border-panel-border bg-transparent px-6 py-4 font-mono text-xs font-bold tracking-[0.3em] text-muted-foreground transition-all hover:bg-white/5 hover:text-foreground disabled:opacity-50"
                       >
                         <ChevronLeft className="h-4 w-4 transition-transform group-hover:-translate-x-1" />
                         PREV
                       </button>
                     )}

                     {!currentEvaluation ? (
                       <button
                         disabled={isSubmitting}
                         onClick={handleSubmitAnswer}
                         className={`group relative flex items-center gap-3 border ${!isAnswered ? 'border-panel-border text-muted-foreground opacity-50 shadow-none' : 'border-neon-cyan bg-neon-cyan/10 text-neon-cyan hover:bg-neon-cyan/20 hover:shadow-[0_0_25px_rgba(0,240,255,0.2)]'} px-10 py-4 font-mono text-xs font-bold tracking-[0.3em] transition-all`}
                       >
                         {isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : "SUBMIT ANSWER"}
                         {!isSubmitting && <Zap className={`h-3.5 w-3.5 fill-current ${!isAnswered ? 'opacity-50' : ''}`} />}
                       </button>
                     ) : (
                       <button
                         onClick={handleNext}
                         className="group relative flex items-center gap-3 border border-foreground bg-foreground px-10 py-4 font-mono text-xs font-bold tracking-[0.3em] text-deep-bg transition-all hover:bg-white hover:shadow-[0_0_25px_rgba(255,255,255,0.2)]"
                       >
                         {currentQ === initialQuestions.length - 1 ? "FINISH SESSION" : "NEXT QUESTION"}
                         <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                       </button>
                     )}
                  </div>
               </div>
            </div>
          </div>
        )}
      </div>

      {/* Decorative timer shadow */}
      <div 
        className="fixed top-0 left-0 w-full h-[1px] bg-neon-cyan/50 pointer-events-none transition-all duration-1000 ease-linear z-50"
        style={{ width: `${timerPercent}%` }}
      />
    </div>
  )
}

