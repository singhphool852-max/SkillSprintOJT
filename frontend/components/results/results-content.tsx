"use client"

import { useEffect, useState, useMemo } from "react"
import Link from "next/link"
import {
  ChevronRight,
  Clock,
  Shield,
  Swords,
  Target,
  Trophy,
  Zap,
  Loader2,
  CheckCircle,
  XCircle,
  MessageSquare,
  ArrowUpRight,
  Activity,
  Award,
  RotateCcw,
  Home,
  ArrowRight
} from "lucide-react"
import { AIDebrief } from "@/components/train/ai-debrief"
import { recommendDifficulty, analyzeWeaknesses } from "@/lib/training-history"
import { API_URL } from "@/lib/api-config"

interface Attempt {
  id: string;
  score: number;
  totalQuestions: number;
  startedAt: string;
  completedAt: string;
  quiz?: { title: string; arenaId?: string };
}

interface Answer {
  id: string;
  questionId: string;
  isCorrect: boolean;
  score: number;
  feedback: string;
  explanation: string;
  evaluatedBy: string;
  writtenAnswer?: string;
}

export function ResultsContent({ id }: { id?: string }) {
  const [loading, setLoading] = useState(true)
  const [attempt, setAttempt] = useState<Attempt | null>(null)
  const [answers, setAnswers] = useState<Answer[]>([])
  const [isLocalResult, setIsLocalResult] = useState(false)

  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  useEffect(() => {
    async function fetchResults() {
      // 1. PRIMARY: Read local session data first
      let localRendered = false
      const localData = sessionStorage.getItem("skillsprint_train_result")
      
      if (localData) {
        try {
          const data = JSON.parse(localData)
          setAttempt(data.attempt)
          setAnswers(data.answers)
          setIsLocalResult(true)
          setLoading(false)
          localRendered = true
        } catch (e) {
          console.error("Local data parsing failed", e)
        }
      }

      // 2. ENHANCEMENT: If we have an attemptId that is not 'local', try to fetch the real backend data
      // This happens silently behind the scenes if we already rendered local data.
      if (id && id !== 'local') {
        try {
          const res = await fetch(`${API_URL}/api/attempts/${id}`, { credentials: "include" })
          if (res.ok) {
            const data = await res.json()
            setAttempt(data.attempt)
            setAnswers(data.answers)
            setIsLocalResult(false)
            if (!localRendered) setLoading(false)
            return
          }
        } catch (err) {
          console.error("Backend fetch failed. Fallback UI will persist.", err)
        }
      }

      // 3. COMPLETE FAILURE: No local data, and backend data failed/missing
      if (!localRendered) {
        setErrorMsg("Session data not found. Please start a new session.")
        setLoading(false)
      }
    }
    
    fetchResults()
  }, [id])

  // Analytics
  const analytics = useMemo(() => {
    if (!attempt || answers.length === 0) return null

    const correctCount = answers.filter(a => a.isCorrect).length
    const incorrectCount = attempt.totalQuestions - correctCount
    const accuracy = Math.round((correctCount / attempt.totalQuestions) * 100)
    
    const start = new Date(attempt.startedAt).getTime()
    const end = new Date(attempt.completedAt).getTime()
    const totalTimeSec = Math.round((end - start) / 1000)
    
    // speedScore: normalized against 30s per question baseline
    const expectedTime = attempt.totalQuestions * 30
    const speedFactor = Math.max(0.2, 1 - (totalTimeSec / expectedTime))
    const speedScore = Math.round(speedFactor * 100)

    let strongAreas = accuracy > 70 ? [attempt.quiz?.title || "Current Domain"] : []
    let weakAreas = accuracy < 70 ? [attempt.quiz?.title || "Current Domain"] : []
    let revisionConcept = "Syntax & Semantics"

    // Integrate with local history
    const weaknesses = analyzeWeaknesses()
    const diffRec = recommendDifficulty(accuracy, "Medium") // Baseline medium

    let smartFeedback = ""
    if (speedScore > 80 && accuracy > 80) {
       smartFeedback = `[ SYSTEM NOTE ]: Exceptional metrics. Sync optimization 100%. \n[ TACTICAL ]: Increasing system pressure is recommended.`
       revisionConcept = "Advanced Architectures"
       strongAreas.push("Syntactic Velocity", "Pattern Recognition")
    } else if (accuracy > 70) {
       smartFeedback = `[ SYSTEM NOTE ]: Acceptable pass rate identified. Routine stable. \n[ TACTICAL ]: Sustain current parameters.`
       revisionConcept = "State Management"
       strongAreas.push("Core Logic")
    } else if (speedScore > 80 && accuracy < 50) {
       smartFeedback = `[ WARNING ]: Reckless inputs detected. Velocity does not excuse syntax degradation. \n[ TACTICAL ]: Downgrading difficulty to foundational nodes.`
       revisionConcept = "Core Principles"
       weakAreas.push("Accuracy Threshold", "Logic Verification")
    } else {
       smartFeedback = `[ SYSTEM NOTE ]: Accuracy below optimal threshold. \n[ TACTICAL ]: ${weaknesses}`
       revisionConcept = "Debugging Fundamentals"
       weakAreas.push("Domain Knowledge", "Constraint Handling")
    }

    return { accuracy, correctCount, incorrectCount, totalTimeSec, speedScore, strongAreas, weakAreas, revisionConcept, smartFeedback, diffRec }
  }, [attempt, answers])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6">
        <div className="relative">
          <Loader2 className="h-12 w-12 text-neon-cyan animate-spin opacity-20" />
          <Zap className="absolute inset-3.5 h-5 w-5 text-neon-cyan animate-pulse" />
        </div>
        <div className="text-center space-y-2">
          <span className="font-mono text-xs tracking-[0.4em] text-neon-cyan uppercase">DECRYPTING MISSION DATA...</span>
          <div className="h-1 w-48 bg-panel-border overflow-hidden">
             <div className="h-full bg-neon-cyan animate-progress" />
          </div>
        </div>
      </div>
    )
  }

  if (!attempt || !analytics) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[70vh] gap-6 px-4">
        <div className="flex flex-col items-center gap-4 text-center max-w-md p-8 bg-transparent">
          <p className="font-mono text-sm leading-relaxed text-muted-foreground uppercase">
            {errorMsg || "Session data not found. Please start a new session."}
          </p>
          <div className="mt-6 flex flex-col w-full gap-3">
             <Link 
               href="/train"
               className="flex items-center justify-center gap-3 border border-panel-border bg-transparent px-6 py-4 font-mono text-xs font-bold tracking-[0.2em] text-white hover:bg-white/5 transition-all"
             >
               <RotateCcw className="h-4 w-4" />
               START NEW SESSION
             </Link>
             <Link 
               href="/dashboard"
               className="flex items-center justify-center gap-3 border border-neon-cyan/20 bg-neon-cyan/5 px-6 py-4 font-mono text-xs font-bold tracking-[0.2em] text-neon-cyan hover:bg-neon-cyan/10 transition-all"
             >
               <Home className="h-4 w-4" />
               RETURN TO HUB
             </Link>
          </div>
        </div>
      </div>
    )
  }

  const isWin = analytics.accuracy > 70

  return (
    <div className="relative min-h-screen pb-20 overflow-hidden">
      <div className="absolute inset-0 grid-bg opacity-20" />
      
      <div className="relative z-10 mx-auto max-w-5xl px-4 py-8">
        {/* Header HUD */}
        <div className="relative mb-16 pt-8 text-center">
            {/* Offline Notification */}
            {isLocalResult && (
              <div className="flex justify-center mb-6">
                <div className="inline-flex items-center gap-2 border border-neon-yellow/30 bg-neon-yellow/10 px-4 py-1.5 rounded-full">
                  <Shield className="h-3 w-3 text-neon-yellow" />
                  <span className="font-mono text-[9px] font-black tracking-widest text-neon-yellow uppercase">Showing local session results</span>
                </div>
              </div>
            )}
            
            {/* Background elements */}
            <div className={`absolute top-0 left-1/2 -translate-x-1/2 w-96 h-96 blur-[120px] opacity-20 ${isWin ? 'bg-neon-cyan' : 'bg-neon-pink'}`} />

            <div className={`inline-flex items-center gap-3 border px-6 py-2 mb-8 ${
                isWin ? "border-neon-cyan/30 bg-neon-cyan/10" : "border-neon-pink/30 bg-neon-pink/10"
            }`}>
                <Trophy className={`h-5 w-5 ${isWin ? "text-neon-cyan" : "text-neon-pink"}`} />
                <span className={`font-mono text-xs tracking-[0.5em] font-black italic ${isWin ? "text-neon-cyan" : "text-neon-pink"}`}>
                {isWin ? "NEURAL ASCENSION COMPLETE" : "MISSION COMPROMISED"}
                </span>
            </div>

            <div className="relative">
                <h1 className="text-5xl lg:text-7xl font-black uppercase tracking-tighter text-foreground mb-4 italic">
                    {attempt.quiz?.title || "ARENA_BATTLE"}
                </h1>
                <div className="flex flex-col items-center gap-2">
                    <div className="flex items-baseline gap-4">
                        <span className={`text-8xl lg:text-9xl font-black tracking-tighter ${isWin ? 'text-neon-cyan text-glow-cyan' : 'text-neon-pink text-glow-pink'}`}>
                          {attempt.score}
                        </span>
                        <span className="font-mono text-xl text-muted-foreground">PTS</span>
                    </div>
                </div>

                {/* Smart Feedback Banner */}
                <div className="mt-8 max-w-2xl mx-auto p-4 border border-panel-border bg-panel-bg/60 backdrop-blur-md">
                    <div className="flex items-center justify-center gap-3">
                        <MessageSquare className={`h-5 w-5 ${isWin ? "text-neon-cyan" : "text-neon-yellow"}`} />
                        <p className="font-mono text-sm leading-relaxed text-foreground">
                            {analytics.smartFeedback}
                        </p>
                    </div>
                </div>
            </div>
        </div>

        {/* Telemetry Grid */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-16">
            {[
                { label: "QUERIES", val: attempt.totalQuestions, icon: Activity, color: 'text-white' },
                { label: "STABLE_SYNCS", val: analytics.correctCount, icon: CheckCircle, color: 'text-neon-cyan' },
                { label: "ACCURACY", val: `${analytics.accuracy}%`, icon: Target, color: isWin ? 'text-neon-cyan' : 'text-neon-pink' },
                { label: "TIME_ELAPSED", val: `${analytics.totalTimeSec}s`, icon: Clock, color: 'text-neon-yellow' }
            ].map((stat, i) => (
                <div key={i} className="group relative border border-panel-border bg-panel-bg/20 p-6 backdrop-blur-md transition-all hover:bg-panel-bg/40">
                    <stat.icon className={`h-5 w-5 mb-4 ${stat.color} opacity-80`} />
                    <div className="flex flex-col">
                        <span className="font-mono text-[10px] text-muted-foreground uppercase tracking-widest mb-1">{stat.label}</span>
                        <span className={`text-3xl font-black tracking-tight ${stat.color}`}>{stat.val}</span>
                    </div>
                </div>
            ))}
        </div>

        {/* Neural Analysis Sections */}
        <div className="grid lg:grid-cols-2 gap-8 mb-16">
            {/* Strengths */}
            <div className="relative space-y-4 border-l-2 border-neon-cyan/40 pl-6 py-4 bg-neon-cyan/5">
                <div className="flex items-center gap-2 text-neon-cyan">
                    <Zap className="h-4 w-4 fill-current" />
                    <span className="font-mono text-xs font-bold uppercase tracking-widest">Neural Strengths</span>
                </div>
                <div className="flex flex-wrap gap-2">
                    {analytics.strongAreas.map(area => (
                        <span key={area} className="px-3 py-1 border border-neon-cyan/20 bg-neon-cyan/10 font-mono text-[10px] text-neon-cyan uppercase">
                           {area}
                        </span>
                    ))}
                    {analytics.strongAreas.length === 0 && <span className="text-[10px] font-mono text-muted-foreground uppercase">Scanning data...</span>}
                </div>
            </div>

            {/* Vulnerabilities */}
            <div className="relative space-y-4 border-l-2 border-neon-pink/40 pl-6 py-4 bg-neon-pink/5">
                <div className="flex items-center gap-2 text-neon-pink">
                    <Shield className="h-4 w-4 fill-current" />
                    <span className="font-mono text-xs font-bold uppercase tracking-widest">System Vulnerabilities</span>
                </div>
                <div className="flex flex-wrap gap-2">
                    {analytics.weakAreas.map(area => (
                        <span key={area} className="px-3 py-1 border border-neon-pink/20 bg-neon-pink/10 font-mono text-[10px] text-neon-pink uppercase">
                           {area}
                        </span>
                    ))}
                    {analytics.weakAreas.length === 0 && <span className="text-[10px] font-mono text-muted-foreground uppercase">No anomalies detected.</span>}
                </div>
            </div>
        </div>

        {/* Detailed Logs Overlay */}
        <div className="space-y-4 mb-20">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-3">
                    <div className="h-px w-8 bg-neon-cyan" />
                    <span className="font-mono text-[10px] tracking-[0.3em] font-bold text-foreground uppercase">SESSION_LOGS</span>
                </div>
                <span className="font-mono text-[9px] text-muted-foreground uppercase">Integrity Guaranteed</span>
            </div>

            {answers.map((ans, idx) => (
                <div key={ans.id} className={`border border-panel-border transition-all ${ans.isCorrect ? 'bg-panel-bg/20' : 'bg-neon-pink/[0.02] border-neon-pink/10'}`}>
                    <div className="flex items-center justify-between p-4 border-b border-panel-border/30">
                        <div className="flex items-center gap-4">
                            <span className="font-mono text-[10px] text-muted-foreground">ID_0{idx+1}</span>
                            <div className={`h-1.5 w-1.5 rounded-full ${ans.isCorrect ? 'bg-neon-cyan' : 'bg-neon-pink'} shadow-[0_0_8px_currentColor]`} />
                            <span className={`font-mono text-[10px] font-bold ${ans.isCorrect ? 'text-neon-cyan' : 'text-neon-pink'} uppercase`}>
                                {ans.isCorrect ? 'SYNC_STABLE' : 'DATA_LOSS'}
                            </span>
                        </div>
                    </div>
                    
                    <div className="p-4 space-y-4">
                        {ans.writtenAnswer && (
                            <div className="border-l-2 border-panel-border/50 pl-4 py-2">
                                <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-widest block mb-1">User Neural Input</span>
                                <div className="font-mono text-xs text-foreground bg-deep-bg/50 p-3 overflow-x-auto whitespace-pre-wrap">
                                    {ans.writtenAnswer}
                                </div>
                            </div>
                        )}
                        <AIDebrief 
                            isCorrect={ans.isCorrect}
                            feedback={ans.feedback}
                            explanation={ans.explanation}
                        />
                    </div>
                </div>
            ))}
        </div>

        {/* Neural Optimization Path */}
        <div className="mb-16 animate-in slide-in-from-bottom-8 duration-700">
            <div className="relative border border-neon-yellow/30 bg-neon-yellow/5 p-8 overflow-hidden">
                <div className="absolute top-0 right-0 p-4 opacity-10">
                    <Shield className="h-20 w-20 text-neon-yellow" />
                </div>
                <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-8">
                    <div className="space-y-4">
                        <div className="flex items-center gap-3">
                            <div className="h-2 w-2 rounded-full bg-neon-yellow animate-pulse" />
                            <span className="font-mono text-xs font-black tracking-[0.3em] text-neon-yellow uppercase">NEURAL_REOPTIMIZATION_PATH</span>
                        </div>
                        <h3 className="text-2xl font-bold text-foreground tracking-tight">
                            Critical Concept Revision: <span className="text-neon-yellow uppercase">{analytics.revisionConcept}...</span>
                        </h3>
                        <p className="text-sm text-muted-foreground max-w-xl leading-relaxed">
                            Based on your behavioral synchronicity, the system recommends a deep-dive into this technical domain before your next evolution.
                        </p>
                    </div>
                    <Link
                        href="/train"
                        className="flex items-center justify-center gap-4 border border-neon-yellow bg-neon-yellow/10 px-8 py-5 font-mono text-xs font-black tracking-widest text-neon-yellow hover:bg-neon-yellow/20 transition-all group"
                    >
                        INITIALIZE REAPPLY
                        <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                    </Link>
                </div>
            </div>
        </div>

        {/* Final Action Bar */}
        <div className="sticky bottom-8 z-20 flex flex-col sm:flex-row gap-4 justify-center bg-deep-bg/90 backdrop-blur-xl p-4 border border-panel-border shadow-[0_0_50px_rgba(0,0,0,0.5)]">
            <Link
                href="/train"
                className="group flex flex-1 items-center justify-center gap-3 border border-neon-cyan/50 bg-neon-cyan/10 py-4 font-mono text-xs font-black tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 shadow-[0_0_20px_rgba(0,240,255,0.1)]"
            >
                <RotateCcw className="h-4 w-4" />
                {analytics.diffRec.action === "Retry Same Level" ? "RETRY SESSION" : "TRAIN SAME DOMAIN"}
            </Link>
            <Link
                href={`/train`}
                className={`group flex flex-1 items-center justify-center gap-3 border py-4 font-mono text-xs font-black tracking-widest transition-all ${
                  analytics.diffRec.direction === "UP" 
                    ? "border-neon-pink/50 bg-neon-pink/10 text-neon-pink hover:bg-neon-pink/20" 
                    : analytics.diffRec.direction === "DOWN"
                    ? "border-neon-yellow/50 bg-neon-yellow/10 text-neon-yellow hover:bg-neon-yellow/20"
                    : "border-muted-foreground/30 bg-muted-foreground/5 text-muted-foreground hover:bg-muted-foreground/10"
                }`}
            >
                <Target className="h-4 w-4" />
                {analytics.diffRec.action.toUpperCase()}
            </Link>
            <Link
                href="/train"
                className="group flex flex-1 items-center justify-center gap-3 border border-neon-yellow/50 bg-neon-yellow/10 py-4 font-mono text-xs font-black tracking-widest text-neon-yellow transition-all hover:bg-neon-yellow/20"
            >
                <Swords className="h-4 w-4" />
                TRAIN ANOTHER
            </Link>
            <Link
                href="/dashboard"
                className="group flex items-center justify-center gap-4 border border-panel-border bg-panel-bg/40 px-8 py-4 font-mono text-xs font-bold tracking-widest text-muted-foreground hover:text-foreground transition-all"
            >
                <Home className="h-4 w-4" />
                HUB
            </Link>
        </div>
      </div>
    </div>
  )
}

