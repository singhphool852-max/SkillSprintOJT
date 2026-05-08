"use client"

import { useState } from "react"
import { X, Target, Zap, ChevronRight, Info, Brain, ShieldAlert } from "lucide-react"
import { SynthesisLoadingHUD } from "./SynthesisLoadingHUD"
import { API_URL } from "@/lib/api-config"

interface SessionSetupPanelProps {
  topic: string
  mode: string
  onClose: () => void
  onStart: (quizId: string) => void
}

export function SessionSetupPanel({ topic, mode, onClose, onStart }: SessionSetupPanelProps) {
  const [difficulty, setDifficulty] = useState<string>("Medium")
  const [count, setCount] = useState<number>(10)
  const [isGenerating, setIsGenerating] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleStart = async () => {
    if (isGenerating) return // Prevent double-clicks
    setIsGenerating(true)
    setError(null)

    try {
      const normalizedTopic = topic.toLowerCase()
      console.log("[TargetMode] Initializing Generation:", { topic: normalizedTopic, difficulty, count })

      const res = await fetch(`${API_URL}/api/training/generate`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          topic: normalizedTopic,
          difficulty: difficulty.toLowerCase(),
          count
        }),
        credentials: "include"
      })

      if (res.ok) {
        const data = await res.json()
        onStart(data.sessionId || data.session_id)
      } else {
        const errData = await res.json().catch(() => ({}))
        setError(errData.error || "Neural synthesis interrupted")
        setIsGenerating(false)
      }
    } catch (err: any) {
      console.error(err)
      setError(err?.message || "Server unreachable. Please try again.")
      setIsGenerating(false)
    }
  }

  const difficulties = ["Easy", "Medium", "Hard"]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-deep-bg/80 backdrop-blur-md p-4 animate-in fade-in duration-300">
      {isGenerating && <SynthesisLoadingHUD topic={topic} difficulty={difficulty} />}
      
      <div className="relative w-full max-w-lg border border-neon-cyan/30 bg-panel-bg/90 shadow-[0_0_50px_rgba(0,240,255,0.1)] p-8 lg:p-10 overflow-hidden">
        {/* Decorative elements */}
        <div className="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-neon-cyan/50 to-transparent" />
        <div className="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
          <Target className="h-32 w-32 -mr-12 -mt-12" />
        </div>

        {/* Close Button */}
        <button 
          onClick={onClose}
          className="absolute top-4 right-4 h-8 w-8 flex items-center justify-center text-muted-foreground hover:text-foreground transition-colors"
        >
          <X className="h-4 w-4" />
        </button>

        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="h-6 w-6 flex items-center justify-center border border-neon-cyan bg-neon-cyan/10">
              <Zap className="h-3 w-3 text-neon-cyan" />
            </div>
            <span className="font-mono text-[10px] tracking-[0.3em] text-neon-cyan uppercase">
              SESSION CONFIGURATION
            </span>
          </div>
          <h2 className="text-2xl font-bold tracking-tight text-foreground uppercase">
             {mode} // <span className="text-neon-cyan text-glow-cyan">{topic}</span>
          </h2>
        </div>

        <div className="space-y-8">
          {/* Difficulty Selector */}
          <div>
            <label className="block font-mono text-[10px] text-muted-foreground uppercase tracking-widest mb-4">
              INTENSITY LEVEL
            </label>
            <div className="grid grid-cols-3 gap-3">
              {difficulties.map((d) => {
                const isActive = difficulty === d
                return (
                  <button
                    key={d}
                    onClick={() => setDifficulty(d)}
                    className={`py-3 border font-mono text-[10px] font-bold tracking-widest transition-all ${
                      isActive 
                        ? 'border-neon-cyan bg-neon-cyan/20 text-neon-cyan shadow-[0_0_15px_rgba(0,240,255,0.2)]' 
                        : 'border-panel-border bg-white/5 text-muted-foreground hover:border-neon-cyan/40 hover:text-foreground'
                    }`}
                  >
                    {d.toUpperCase()}
                  </button>
                )
              })}
            </div>
          </div>

          {/* Question Count */}
          <div>
            <div className="flex justify-between items-end mb-4">
              <label className="block font-mono text-[10px] text-muted-foreground uppercase tracking-widest">
                QUESTION STACK SIZE
              </label>
              <span className="font-mono text-xl font-bold text-neon-cyan">{count}</span>
            </div>
            <input 
              type="range"
              min="5"
              max="15"
              step="1"
              value={count}
              onChange={(e) => setCount(parseInt(e.target.value))}
              className="w-full h-1.5 bg-white/10 rounded-none appearance-none cursor-pointer accent-neon-cyan hover:accent-neon-cyan transition-all"
            />
            <div className="flex justify-between mt-2 font-mono text-[8px] text-muted-foreground uppercase">
              <span>Min 5</span>
              <span>Max 15</span>
            </div>
          </div>

          {error && (
            <div className="flex items-center gap-3 p-4 border border-neon-pink/30 bg-neon-pink/5">
               <ShieldAlert className="h-4 w-4 text-neon-pink" />
               <p className="font-mono text-[10px] text-neon-pink uppercase tracking-widest font-bold">{error}</p>
            </div>
          )}

          <div className="flex gap-3 p-4 border border-panel-border bg-white/[0.02]">
            <Info className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
            <p className="text-[10px] leading-relaxed text-muted-foreground uppercase font-mono">
              The AI will generate a specialized {difficulty} set of {count} questions focused on {topic} fundamentals and edge cases.
            </p>
          </div>

          {/* Action */}
          <button 
            disabled={isGenerating}
            onClick={handleStart}
            className="group w-full flex items-center justify-center gap-3 border border-neon-cyan bg-neon-cyan/10 py-5 font-mono text-xs font-bold tracking-[0.3em] text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_25px_rgba(0,240,255,0.2)] disabled:opacity-50"
          >
            {isGenerating ? "SYNTHESIZING..." : "INITIALIZE NEURAL LINK"}
            <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
          </button>
        </div>
      </div>
    </div>
  )
}

