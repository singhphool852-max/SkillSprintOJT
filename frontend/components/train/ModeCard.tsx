"use client"

import { LucideIcon, Play } from "lucide-react"

interface ModeCardProps {
  title: string
  description: string
  icon: LucideIcon
  duration: string
  difficulty: "Beginner" | "Intermediate" | "Advanced" | "Adaptive" | "Variable"
  color: "cyan" | "pink" | "yellow"
  isPremium?: boolean
  isActive?: boolean
  onClick?: () => void
}

export function ModeCard({ title, description, icon: Icon, duration, difficulty, color, isPremium, isActive, onClick }: ModeCardProps) {
  const themes = {
    cyan: "border-neon-cyan/20 text-neon-cyan bg-neon-cyan/5 hover:border-neon-cyan hover:shadow-[0_0_20px_rgba(0,240,255,0.15)]",
    pink: "border-neon-pink/20 text-neon-pink bg-neon-pink/5 hover:border-neon-pink hover:shadow-[0_0_20px_rgba(255,45,111,0.15)]",
    yellow: "border-neon-yellow/10 text-neon-yellow bg-neon-yellow/5 hover:border-neon-yellow/60 hover:shadow-[0_0_20px_rgba(251,191,36,0.1)]"
  }

  return (
    <div 
      onClick={onClick}
      className={`group flex flex-col border ${themes[color]} ${isActive ? 'bg-white/[0.03] border-opacity-100 scale-[1.02]' : ''} p-6 transition-all duration-300 relative overflow-hidden h-full cursor-pointer ${isPremium ? 'glow-pink border-neon-pink/40' : ''}`}
    >
      {isPremium && (
        <div className="absolute -right-8 -top-8 h-24 w-24 bg-neon-pink/10 blur-3xl rounded-full pointer-events-none group-hover:bg-neon-pink/20 transition-all" />
      )}

      <div className="flex items-start justify-between mb-4 relative z-10">
        <div className={`flex ${isPremium ? 'h-16 w-16' : 'h-12 w-12'} shrink-0 items-center justify-center border border-inherit bg-white/5 transition-all duration-300`}>
          <Icon className={`${isPremium ? 'h-8 w-8' : 'h-6 w-6'}`} strokeWidth={1.5} />
        </div>
        {isPremium && (
          <span className="font-mono text-[8px] font-bold tracking-[0.2em] text-neon-pink border border-neon-pink/30 px-2 py-0.5 uppercase">
            PREMIUM
          </span>
        )}
      </div>

      <div className="flex-1 relative z-10 min-w-0">
        <div className="flex items-center gap-3 mb-1.5">
          <h3 className={`font-mono ${isPremium ? 'text-sm' : 'text-xs'} font-bold tracking-[0.2em] text-foreground uppercase truncate`}>
            {title}
          </h3>
          <span className="font-mono text-[8px] text-muted-foreground uppercase py-0.5 px-2 border border-panel-border bg-panel-bg/40">
            {difficulty}
          </span>
        </div>
        <p className={`text-[11px] leading-relaxed text-muted-foreground/80 group-hover:text-muted-foreground transition-colors mb-4 ${isPremium ? 'line-clamp-3' : 'line-clamp-2'}`}>
          {description}
        </p>
      </div>

      <div className="mt-auto pt-4 flex items-center justify-between relative z-10 border-t border-panel-border/30">
        <div className="flex items-center gap-3 text-[9px] font-mono text-muted-foreground uppercase tracking-widest">
          <span className="flex items-center gap-1.5">
             <div className={`h-1 w-1 rounded-full ${isPremium ? 'bg-neon-pink' : 'bg-inherit'}`} />
             {duration}
          </span>
        </div>
        <div className="flex items-center gap-2 group/btn">
          <span className="font-mono text-[9px] font-bold tracking-[0.2em] text-muted-foreground group-hover:text-foreground transition-colors uppercase">
            START MODE
          </span>
          <div className="flex h-7 w-7 items-center justify-center border border-panel-border bg-panel-bg/40 text-muted-foreground group-hover:bg-foreground group-hover:text-deep-bg transition-all">
            <Play className="h-2.5 w-2.5 fill-current" />
          </div>
        </div>
      </div>
    </div>
  )
}
