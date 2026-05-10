"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import Image from "next/image"
import { ChevronRight, Crosshair, Flame, Shield, Swords, Trophy, Zap } from "lucide-react"

function AnimatedCounter({ target, suffix = "" }: { target: number; suffix?: string }) {
  const [count, setCount] = useState(0)
  useEffect(() => {
    let frame: number
    const duration = 2000
    const start = performance.now()
    function animate(now: number) {
      const elapsed = now - start
      const progress = Math.min(elapsed / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      setCount(Math.floor(eased * target))
      if (progress < 1) frame = requestAnimationFrame(animate)
    }
    frame = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(frame)
  }, [target])
  return (
    <span>
      {count.toLocaleString()}
      {suffix}
    </span>
  )
}

function HudCorner({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      width="20"
      height="20"
      viewBox="0 0 20 20"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path d="M0 0H20V2H2V20H0V0Z" fill="currentColor" />
    </svg>
  )
}

const leaderboardData = [
  { rank: 1, name: "SYNERGY", points: 4820, icon: "cyan" },
  { rank: 2, name: "SHADOWFOX", points: 4510, icon: "pink" },
  { rank: 3, name: "NEONBLADE", points: 4365, icon: "yellow" },
]

export function HeroSection() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <section className="relative min-h-screen overflow-hidden bg-deep-bg">
      {/* Anime background image - atmospheric layer */}
      <div className="absolute inset-0">
        <Image
          src="https://hebbkx1anhila5yf.public.blob.vercel-storage.com/image-HEb82YGGhzok3kcILKzoCSYySBrIGL.png"
          alt=""
          fill
          className="object-cover object-center"
          priority
        />
        {/* Heavy dark overlay for readability */}
        <div className="absolute inset-0 bg-deep-bg/70" />
        {/* Gradient overlays for edge blending */}
        <div className="absolute inset-0 bg-gradient-to-t from-deep-bg via-deep-bg/40 to-deep-bg/60" />
        <div className="absolute inset-0 bg-gradient-to-b from-deep-bg/80 via-transparent to-deep-bg" />
        {/* Side vignettes */}
        <div className="absolute inset-0 bg-gradient-to-r from-deep-bg/80 via-transparent to-deep-bg/80" />
      </div>

      {/* Grid background overlay */}
      <div className="absolute inset-0 grid-bg opacity-40" />

      {/* Scanlines */}
      <div className="absolute inset-0 scanlines" />

      {/* Ambient glow effects */}
      <div className="absolute top-1/3 left-1/4 w-[500px] h-[500px] rounded-full bg-neon-cyan/5 blur-[150px]" />
      <div className="absolute top-1/3 right-1/4 w-[400px] h-[400px] rounded-full bg-neon-pink/5 blur-[120px]" />

      {/* Content layer */}
      <div className="relative z-10 mx-auto flex min-h-screen max-w-7xl flex-col px-4 pt-20 lg:px-8">

        {/* Top status bar */}
        <div
          className={`flex items-center justify-between py-4 transition-all duration-700 ${mounted ? "opacity-100" : "opacity-0"}`}
        >
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 border border-neon-cyan/30 bg-neon-cyan/5 px-3 py-1 backdrop-blur-sm">
              <div className="h-1.5 w-1.5 rounded-full bg-neon-cyan animate-pulse-glow" />
              <span className="font-mono text-[10px] tracking-[0.2em] text-neon-cyan">
                SYSTEM ONLINE
              </span>
            </div>
            <div className="hidden h-px flex-1 max-w-24 bg-panel-border sm:block" />
            <span className="hidden font-mono text-[10px] tracking-widest text-muted-foreground sm:block">
              v3.2.1
            </span>
          </div>
          <div className="flex items-center gap-3">
            <span className="font-mono text-[10px] tracking-widest text-muted-foreground">
              PLAYERS ONLINE:
            </span>
            <span className="font-mono text-[10px] font-bold text-neon-cyan text-glow-cyan">
              {mounted ? <AnimatedCounter target={2847} /> : "0"}
            </span>
          </div>
        </div>

        {/* Main content: 3-column layout */}
        <div className="flex flex-1 flex-col items-center justify-center gap-6 py-8 lg:flex-row lg:items-center lg:gap-6">

          {/* LEFT: Your Stats HUD Panel */}
          <div
            className={`w-full max-w-xs transition-all duration-700 delay-200 lg:w-72 ${mounted ? "opacity-100 translate-x-0" : "opacity-0 -translate-x-8"}`}
          >
            <div className="relative border border-neon-cyan/30 bg-deep-bg/80 backdrop-blur-md overflow-hidden">
              <HudCorner className="absolute -top-px -left-px text-neon-cyan/70" />
              <HudCorner className="absolute -top-px -right-px text-neon-cyan/70 rotate-90" />
              <HudCorner className="absolute -bottom-px -right-px text-neon-cyan/70 rotate-180" />
              <HudCorner className="absolute -bottom-px -left-px text-neon-cyan/70 -rotate-90" />

              {/* Panel header */}
              <div className="border-b border-neon-cyan/20 bg-neon-cyan/5 px-5 py-3">
                <div className="flex items-center gap-2">
                  <Shield className="h-3.5 w-3.5 text-neon-cyan" />
                  <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-cyan">
                    YOUR STATS
                  </span>
                </div>
              </div>

              {/* Stats list */}
              <div className="flex flex-col gap-0.5 p-4">
                {[
                  { label: "SPEED", value: "92%", icon: Zap, color: "neon-cyan" },
                  { label: "ACCURACY", value: "87%", icon: Crosshair, color: "neon-cyan" },
                  { label: "WINS", value: "315", icon: Trophy, color: "neon-yellow" },
                  { label: "RANK", value: "DIAMOND II", icon: Shield, color: "neon-pink" },
                ].map((stat) => (
                  <div
                    key={stat.label}
                    className="flex items-center justify-between border-b border-panel-border/50 py-3 last:border-0"
                  >
                    <div className="flex items-center gap-3">
                      <stat.icon className={`h-3.5 w-3.5 text-${stat.color}`} />
                      <span className="font-mono text-[11px] tracking-wider text-muted-foreground">
                        {stat.label}
                      </span>
                    </div>
                    <span className={`font-mono text-sm font-bold text-${stat.color}`}>
                      {stat.value}
                    </span>
                  </div>
                ))}
              </div>

              {/* Train Now button */}
              <div className="px-4 pb-4">
                <Link
                  href="/arena"
                  className="flex items-center justify-center gap-2 border border-neon-cyan/40 bg-neon-cyan/10 px-4 py-3 font-mono text-xs tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_20px_rgba(0,240,255,0.15)]"
                >
                  <Zap className="h-3.5 w-3.5" />
                  ENTER ARENA
                </Link>
              </div>
            </div>
          </div>

          {/* CENTER: Main Headline */}
          <div
            className={`flex-1 text-center transition-all duration-700 delay-100 ${mounted ? "opacity-100 translate-y-0 scale-100" : "opacity-0 translate-y-6 scale-95"}`}
          >
            <span className="mb-4 inline-block font-mono text-[10px] tracking-[0.3em] text-neon-pink text-glow-pink">
              {"// COMPETITIVE MICROLEARNING"}
            </span>
            <h1 className="text-balance">
              <span className="block text-4xl font-black uppercase leading-[1.05] tracking-tight text-foreground sm:text-5xl lg:text-6xl xl:text-7xl">
                Master Your
              </span>
              <span className="block text-4xl font-black uppercase leading-[1.05] tracking-tight sm:text-5xl lg:text-6xl xl:text-7xl">
                <span className="text-neon-cyan text-glow-cyan">Skills.</span>
              </span>
              <span className="block text-4xl font-black uppercase leading-[1.05] tracking-tight text-foreground sm:text-5xl lg:text-6xl xl:text-7xl">
                Win The <span className="text-neon-pink text-glow-pink">Race.</span>
              </span>
            </h1>
            <p className="mx-auto mt-6 max-w-md text-sm leading-relaxed text-muted-foreground lg:text-base">
              Enter the arena. Face rapid-fire knowledge challenges.
              Climb the ranks. Prove your dominance.
            </p>

            <div className="mt-8 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
              <Link
                href="/arena"
                className="group flex items-center justify-center gap-3 border border-neon-cyan bg-neon-cyan/10 px-8 py-3.5 font-mono text-xs tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_30px_rgba(0,240,255,0.2)] hud-panel backdrop-blur-sm"
              >
                <Swords className="h-4 w-4" />
                ENTER ARENA
                <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </Link>
              <Link
                href="/dashboard"
                className="flex items-center justify-center gap-3 border border-panel-border bg-deep-bg/50 px-8 py-3.5 font-mono text-xs tracking-widest text-muted-foreground transition-all hover:border-foreground/20 hover:text-foreground backdrop-blur-sm"
              >
                VIEW PROFILE
              </Link>
            </div>
          </div>

          {/* RIGHT: Opponent HUD Panel */}
          <div
            className={`w-full max-w-xs transition-all duration-700 delay-300 lg:w-72 ${mounted ? "opacity-100 translate-x-0" : "opacity-0 translate-x-8"}`}
          >
            <div className="relative border border-neon-pink/30 bg-deep-bg/80 backdrop-blur-md overflow-hidden">
              <HudCorner className="absolute -top-px -left-px text-neon-pink/70" />
              <HudCorner className="absolute -top-px -right-px text-neon-pink/70 rotate-90" />
              <HudCorner className="absolute -bottom-px -right-px text-neon-pink/70 rotate-180" />
              <HudCorner className="absolute -bottom-px -left-px text-neon-pink/70 -rotate-90" />

              {/* Panel header */}
              <div className="border-b border-neon-pink/20 bg-neon-pink/5 px-5 py-3">
                <div className="flex items-center gap-2">
                  <Swords className="h-3.5 w-3.5 text-neon-pink" />
                  <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-pink">
                    OPPONENT
                  </span>
                </div>
              </div>

              {/* Opponent info */}
              <div className="p-5">
                <div className="flex items-center gap-3 mb-4">
                  <div className="flex h-10 w-10 items-center justify-center border border-neon-pink/40 bg-neon-pink/10">
                    <Crosshair className="h-5 w-5 text-neon-pink" />
                  </div>
                  <div>
                    <span className="block font-mono text-sm font-bold tracking-wider text-foreground">
                      RIVAL_X
                    </span>
                    <span className="font-mono text-[10px] tracking-wider text-muted-foreground">
                      SEARCHING...
                    </span>
                  </div>
                </div>

                <div className="flex flex-col gap-3">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] tracking-wider text-muted-foreground">
                      RANK
                    </span>
                    <span className="font-mono text-sm font-bold text-neon-yellow text-glow-yellow">
                      PLATINUM I
                    </span>
                  </div>
                  <div className="h-px bg-panel-border/50" />
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] tracking-wider text-muted-foreground">
                      WIN STREAK
                    </span>
                    <div className="flex items-center gap-1.5">
                      <Flame className="h-3 w-3 text-neon-pink" />
                      <span className="font-mono text-sm font-bold text-neon-pink text-glow-pink">
                        7
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Enter Match button */}
              <div className="px-4 pb-4">
                <Link
                  href="/arena"
                  className="w-full flex items-center justify-center gap-2 border border-neon-pink/60 bg-neon-pink/10 px-4 py-3 font-mono text-xs font-bold tracking-widest text-neon-pink transition-all hover:bg-neon-pink/20 hover:shadow-[0_0_20px_rgba(255,45,111,0.15)] animate-border-glow"
                >
                  <Swords className="h-3.5 w-3.5" />
                  ENTER MATCH
                </Link>
              </div>
            </div>
          </div>
        </div>

        {/* Bottom: Live Leaderboard */}
        <div
          className={`mx-auto mb-6 w-full max-w-lg transition-all duration-700 delay-400 ${mounted ? "opacity-100 translate-y-0" : "opacity-0 translate-y-8"}`}
        >
          <div className="relative border border-panel-border bg-deep-bg/80 backdrop-blur-md overflow-hidden">
            <HudCorner className="absolute -top-px -left-px text-neon-amber/50" />
            <HudCorner className="absolute -top-px -right-px text-neon-amber/50 rotate-90" />

            {/* Leaderboard header */}
            <div className="border-b border-panel-border bg-panel-bg/60 px-5 py-2.5">
              <div className="flex items-center justify-center gap-2">
                <div className="h-px flex-1 bg-neon-amber/20" />
                <Trophy className="h-3 w-3 text-neon-yellow" />
                <span className="font-mono text-[10px] font-bold tracking-[0.2em] text-neon-yellow">
                  LIVE LEADERBOARD
                </span>
                <Trophy className="h-3 w-3 text-neon-yellow" />
                <div className="h-px flex-1 bg-neon-yellow/20" />
              </div>
            </div>

            {/* Leaderboard entries */}
            <div className="flex flex-col">
              {leaderboardData.map((entry) => (
                <div
                  key={entry.rank}
                  className="flex items-center gap-4 border-b border-panel-border/30 px-5 py-2.5 last:border-0"
                >
                  <span className={`font-mono text-sm font-bold ${
                    entry.rank === 1 ? "text-neon-yellow text-glow-yellow" :
                    entry.rank === 2 ? "text-muted-foreground" : "text-muted-foreground"
                  }`}>
                    {entry.rank}
                  </span>
                  <div className={`flex h-6 w-6 items-center justify-center border ${
                    entry.icon === "cyan" ? "border-neon-cyan/30 text-neon-cyan" :
                    entry.icon === "pink" ? "border-neon-pink/30 text-neon-pink" :
                    "border-neon-yellow/30 text-neon-yellow"
                  }`}>
                    <Zap className="h-3 w-3" />
                  </div>
                  <span className="flex-1 font-mono text-xs font-bold tracking-widest text-foreground">
                    {entry.name}
                  </span>
                  <span className="font-mono text-xs text-muted-foreground">
                    <span className={`font-bold ${
                      entry.rank === 1 ? "text-neon-yellow" :
                      entry.rank === 2 ? "text-foreground" : "text-foreground"
                    }`}>{entry.points.toLocaleString()}</span> PTS
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Bottom Navigation Bar */}
        <div
          className={`mb-4 transition-all duration-700 delay-500 ${mounted ? "opacity-100" : "opacity-0"}`}
        >
          <div className="flex items-center justify-center gap-1 overflow-x-auto">
            {[
              { label: "ARENA CHALLENGES", href: "/arena", icon: Flame },
              { label: "BOOSTERS", href: "/dashboard", icon: Zap },
              { label: "QUICKPLAY", href: "/arena", icon: Swords, active: true },
              { label: "LEADERBOARD", href: "/leaderboard", icon: Trophy },
              { label: "ACHIEVEMENTS", href: "/achievements", icon: Shield },
            ].map((item) => (
              <Link
                key={item.label}
                href={item.href}
                className={`flex items-center gap-2 px-4 py-2.5 font-mono text-[10px] tracking-widest transition-all whitespace-nowrap ${
                  item.active
                    ? "border border-neon-cyan/40 bg-neon-cyan/10 text-neon-cyan"
                    : "border border-panel-border/50 bg-deep-bg/60 text-muted-foreground hover:text-foreground hover:border-panel-border backdrop-blur-sm"
                }`}
              >
                <item.icon className="h-3 w-3" />
                <span className="hidden sm:inline">{item.label}</span>
              </Link>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
