"use client"

import { Brain, Info, ChevronRight, Zap, Target, BookOpen, Layers } from "lucide-react"
import { useState, useEffect } from "react"

interface SummaryDisplayProps {
  summary: string
  onStart: () => void
  topic: string
  difficulty: string
}

export function SummaryDisplay({ summary, onStart, topic, difficulty }: SummaryDisplayProps) {
  const [typedSummary, setTypedSummary] = useState("")
  const [isTyping, setIsTyping] = useState(true)
  const [sections, setSections] = useState<string[]>([])

  useEffect(() => {
    // Split summary safely
    const safeSummary = summary || ""
    const parts = safeSummary.split("\n").filter(p => p.trim().length > 0)
    setSections(parts)
  }, [summary])

  return (
    <div className="w-full max-w-5xl mx-auto animate-in fade-in slide-in-from-bottom-8 duration-1000">
      {/* Premium Header */}
      <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
        <div>
          <div className="flex items-center gap-3 mb-3">
            <div className="h-[2px] w-8 bg-neon-cyan" />
            <span className="font-mono text-[10px] tracking-[0.4em] text-neon-cyan uppercase font-bold">
              AI COGNITIVE SYNTHESIS
            </span>
          </div>
          <h1 className="text-4xl md:text-5xl font-black tracking-tighter text-foreground uppercase">
            SESSION <span className="text-neon-cyan text-glow-cyan">SYNOPSIS</span>
          </h1>
        </div>
        
        <div className="flex gap-4 p-4 border border-panel-border bg-panel-bg/40 backdrop-blur-md">
          <div className="flex flex-col border-r border-panel-border pr-6">
            <span className="font-mono text-[9px] text-muted-foreground uppercase mb-1">Target Domain</span>
            <span className="font-mono text-sm font-bold text-neon-cyan tracking-wider uppercase">{topic}</span>
          </div>
          <div className="flex flex-col pl-2">
            <span className="font-mono text-[9px] text-muted-foreground uppercase mb-1">Intensity</span>
            <span className="font-mono text-sm font-bold text-neon-pink tracking-wider uppercase">{difficulty}</span>
          </div>
        </div>
      </div>

      <div className="grid lg:grid-cols-3 gap-8">
        {/* Left Column: Visual Data */}
        <div className="lg:col-span-1 space-y-6">
          <div className="border border-panel-border bg-panel-bg/20 p-6 relative overflow-hidden group">
            <div className="absolute top-0 left-0 w-1 h-full bg-neon-cyan/50" />
            <div className="flex items-center gap-4 mb-6">
              <div className="p-2 bg-neon-cyan/10 border border-neon-cyan/20">
                <Brain className="h-5 w-5 text-neon-cyan" />
              </div>
              <span className="font-mono text-xs font-bold tracking-widest text-foreground uppercase">Neural Context</span>
            </div>
            
            <div className="space-y-4">
              <div className="flex items-center justify-between font-mono text-[10px]">
                <span className="text-muted-foreground uppercase">Retention Target</span>
                <span className="text-neon-cyan">85%</span>
              </div>
              <div className="h-1 bg-white/5 rounded-full overflow-hidden">
                <div className="h-full bg-neon-cyan w-[85%] shadow-[0_0_8px_#00e5ff] animate-pulse" />
              </div>
              
              <div className="flex items-center justify-between font-mono text-[10px]">
                <span className="text-muted-foreground uppercase">Complexity Load</span>
                <span className="text-neon-pink">High</span>
              </div>
              <div className="h-1 bg-white/5 rounded-full overflow-hidden">
                <div className="h-full bg-neon-pink w-[70%] shadow-[0_0_8px_#ff00ff] animate-pulse" />
              </div>
            </div>
            
            <div className="mt-8 pt-6 border-t border-panel-border/30">
               <div className="flex items-start gap-3">
                  <Info className="h-3 w-3 text-neon-yellow mt-0.5" />
                  <p className="font-mono text-[9px] text-muted-foreground leading-relaxed uppercase">
                    Our AI has extracted key patterns from your data. Review the synopsis before initializing the training cycle.
                  </p>
               </div>
            </div>
          </div>
          
          <div className="border border-panel-border bg-panel-bg/20 p-6 relative overflow-hidden flex flex-col items-center text-center">
             <Layers className="h-10 w-10 text-neon-cyan/20 mb-4" />
             <h4 className="font-mono text-[10px] font-bold text-foreground mb-1 uppercase tracking-widest">Mastery Objectives</h4>
             <p className="font-mono text-[9px] text-muted-foreground uppercase">Identify logic gaps // Optimize recall speed // Contextual application</p>
          </div>
        </div>

        {/* Right Column: The Summary Content */}
        <div className="lg:col-span-2">
          <div className="border border-panel-border bg-panel-bg/60 backdrop-blur-xl p-8 lg:p-10 relative overflow-hidden min-h-[400px] flex flex-col">
            {/* Cyberpunk decoration */}
            <div className="absolute top-0 right-0 p-4 font-mono text-[120px] font-black text-white/[0.03] tracking-tighter select-none">DATA</div>
            <div className="absolute -top-24 -left-24 w-48 h-48 bg-neon-cyan/5 blur-[80px] rounded-full" />
            
            <div className="relative z-10 flex-1">
              <div className="flex items-center gap-3 mb-8">
                <div className="h-2 w-2 bg-neon-cyan rounded-full animate-pulse" />
                <span className="font-mono text-[10px] tracking-[0.2em] text-muted-foreground uppercase">SYNOPSIS STREAM // INCOMING</span>
              </div>

              <div className="space-y-6">
                {(sections || []).map((section, idx) => (
                  <div 
                    key={idx} 
                    className="flex gap-4 animate-in fade-in slide-in-from-left-4 duration-500"
                    style={{ animationDelay: `${idx * 150}ms` }}
                  >
                    <div className="mt-1.5 h-1.5 w-1.5 shrink-0 bg-neon-cyan shadow-[0_0_5px_#00e5ff]" />
                    <p className="font-mono text-sm md:text-base leading-relaxed text-foreground/90 tracking-tight">
                      {section.replace(/^[*-]\s*/, "")}
                    </p>
                  </div>
                ))}
              </div>
            </div>

            <div className="mt-12 flex flex-col md:flex-row items-center justify-between gap-6 pt-8 border-t border-panel-border/30 relative z-10">
               <div className="flex items-center gap-4">
                  <div className="flex -space-x-2">
                     <div className="h-8 w-8 rounded-full border border-panel-border bg-panel-bg flex items-center justify-center">
                        <Zap className="h-3 w-3 text-neon-pink" />
                     </div>
                     <div className="h-8 w-8 rounded-full border border-panel-border bg-panel-bg flex items-center justify-center">
                        <Target className="h-3 w-3 text-neon-cyan" />
                     </div>
                  </div>
                  <span className="font-mono text-[9px] text-muted-foreground uppercase tracking-widest">
                    AI AGENTS READY FOR DRILL
                  </span>
               </div>

               <button 
                onClick={onStart}
                className="group relative flex items-center gap-4 bg-neon-cyan text-deep-bg px-10 py-5 font-mono text-xs font-black tracking-[0.4em] transition-all hover:bg-white hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] overflow-hidden"
              >
                <div className="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent -translate-x-[100%] group-hover:translate-x-[100%] transition-transform duration-700" />
                INITIALIZE TRAINING
                <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Background Neural Grid */}
      <div className="fixed inset-0 pointer-events-none opacity-[0.03] z-[-1]">
        <div className="absolute inset-0 grid-bg" />
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-neon-cyan/5 to-transparent animate-scan" />
      </div>
    </div>
  )
}
