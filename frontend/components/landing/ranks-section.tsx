"use client"

import Image from "next/image"
import Link from "next/link"
import { Crown, Flame, Shield, Skull, Star, Sword, Swords, Zap } from "lucide-react"

const ranks = [
  { name: "ROOKIE", icon: Shield, rating: "TOP 100%", color: "neon-muted", border: "border-muted-foreground/20", glow: "shadow-[0_0_15px_rgba(107,107,138,0.2)]" },
  { name: "WARRIOR", icon: Sword, rating: "TOP 75%", color: "neon-cyan", border: "border-neon-cyan/20", glow: "shadow-[0_0_15px_rgba(0,240,255,0.2)]" },
  { name: "ELITE", icon: Star, rating: "TOP 50%", color: "neon-cyan", border: "border-neon-cyan/20", glow: "shadow-[0_0_15px_rgba(0,240,255,0.2)]" },
  { name: "VETERAN", icon: Flame, rating: "TOP 30%", color: "neon-yellow", border: "border-neon-yellow/20", glow: "shadow-[0_0_15px_rgba(255,184,0,0.2)]" },
  { name: "CHAMPION", icon: Skull, rating: "TOP 15%", color: "neon-pink", border: "border-neon-pink/20", glow: "shadow-[0_0_15px_rgba(255,45,111,0.2)]" },
  { name: "APEX", icon: Crown, rating: "TOP 5%", color: "neon-yellow", border: "border-neon-yellow/20", glow: "shadow-[0_0_15px_rgba(255,184,0,0.2)]" },
]

export function RanksSection() {
  return (
    <section className="relative bg-background py-24 lg:py-32 overflow-hidden">
      {/* Ambient image accent */}
      <div className="absolute inset-y-0 left-0 w-1/3 opacity-[0.04] hidden lg:block">
        <Image
          src="https://hebbkx1anhila5yf.public.blob.vercel-storage.com/image-HEb82YGGhzok3kcILKzoCSYySBrIGL.png"
          alt=""
          fill
          className="object-cover object-right"
        />
        <div className="absolute inset-0 bg-gradient-to-l from-background via-background/80 to-transparent" />
      </div>

      <div className="absolute inset-0 grid-bg opacity-30" />
      <div className="absolute bottom-0 left-0 w-[500px] h-[500px] rounded-full bg-neon-amber/3 blur-[120px]" />

      <div className="relative z-10 mx-auto max-w-7xl px-4 lg:px-8">
        {/* Section header */}
        <div className="flex items-center gap-4 mb-4">
          <Swords className="h-4 w-4 text-neon-yellow" />
          <span className="font-mono text-[11px] tracking-[0.3em] text-neon-yellow">
            RANK SYSTEM
          </span>
          <div className="h-px flex-1 bg-panel-border" />
        </div>

        <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl text-balance">
          CLIMB THE <span className="text-neon-yellow text-glow-yellow">RANKS</span>
        </h2>
        <p className="mt-4 max-w-xl text-sm leading-relaxed text-muted-foreground lg:text-base">
          Every correct answer pushes you closer to the top. Every mistake drops you back.
          Only the sharpest minds reach Apex.
        </p>

        {/* Ranks grid */}
        <div className="mt-16 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {ranks.map((rank, i) => {
            const isYellow = rank.color === "neon-yellow"
            const colorClass = rank.color === "neon-muted" ? "text-muted-foreground" : `text-${rank.color}`
            const borderClass = isYellow ? "border-neon-yellow/10" : rank.border
            const hoverBorderClass = isYellow ? "hover:border-neon-yellow/60" : `hover:${rank.border.split('/')[0]}`

            return (
              <div
                key={rank.name}
                className={`group relative flex items-center gap-4 border ${borderClass} ${hoverBorderClass} bg-panel-bg/40 p-5 transition-all duration-300 hover:bg-panel-bg/60 ${rank.glow.replace('0.2', '0.05')} hover:shadow-[0_0_15px_rgba(251,191,36,0.1)] backdrop-blur-sm`}
              >
                <div className={`flex h-10 w-10 items-center justify-center border ${borderClass} ${hoverBorderClass} bg-white/5 ${colorClass} transition-colors duration-300`}>
                  <rank.icon className="h-5 w-5" strokeWidth={1.5} />
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <span className={`font-mono text-sm font-bold tracking-widest ${colorClass} opacity-80 group-hover:opacity-100 transition-opacity`}>
                      {rank.name}
                    </span>
                    <span className="font-mono text-[10px] text-muted-foreground">
                      {`TIER ${i + 1}`}
                    </span>
                  </div>
                  <span className="font-mono text-xs text-muted-foreground">
                    {rank.rating} SR
                  </span>
                </div>
              </div>
            )
          })}
        </div>

        {/* CTA */}
        <div className="mt-16 flex flex-col items-center gap-6">
          <div className="flex items-center gap-4">
            <div className="h-px w-16 bg-panel-border" />
            <span className="font-mono text-[10px] tracking-[0.3em] text-muted-foreground">
              ARE YOU READY?
            </span>
            <div className="h-px w-16 bg-panel-border" />
          </div>
          <Link
            href="/arena"
            className="group relative flex items-center gap-3 border border-neon-cyan bg-neon-cyan/10 px-10 py-4 font-mono text-sm tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_30px_rgba(0,240,255,0.2)]"
          >
            <Zap className="h-4 w-4" />
            START YOUR FIRST ARENA
          </Link>
        </div>
      </div>
    </section>
  )
}
