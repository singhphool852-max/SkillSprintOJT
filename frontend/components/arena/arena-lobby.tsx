"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import {
  ChevronRight,
  Clock,
  Crown,
  Flame,
  Radio,
  Shield,
  Skull,
  Star,
  Swords,
  Users,
  Zap,
} from "lucide-react"
import { API_URL } from "@/lib/api-config"

interface Arena {
  id: string;
  title: string;
  category: { name: string };
  currentPlayers: number;
  maxPlayers: number;
  difficulty: string;
  status: string;
  durationSeconds: number;
}

const difficultyConfig: Record<string, { icon: React.ElementType; color: string }> = {
  WARRIOR: { icon: Shield, color: "text-neon-cyan" },
  ELITE: { icon: Star, color: "text-neon-cyan" },
  VETERAN: { icon: Flame, color: "text-neon-amber" },
  CHAMPION: { icon: Skull, color: "text-neon-pink" },
  APEX: { icon: Crown, color: "text-neon-amber" },
}

const statusConfig: Record<string, { color: string; bg: string }> = {
  live: { color: "text-neon-cyan", bg: "bg-neon-cyan" },
  open: { color: "text-neon-amber", bg: "bg-neon-amber" },
  closed: { color: "text-neon-pink", bg: "bg-neon-pink" },
}

export function ArenaLobby() {
  const router = useRouter()
  const [arenas, setArenas] = useState<Arena[]>([])
  const [loading, setLoading] = useState(true)
  const [joining, setJoining] = useState<string | null>(null)
  const [countdown, setCountdown] = useState(3)

  useEffect(() => {
    async function fetchArenas() {
      try {
        const res = await fetch(`${API_URL}/api/arenas`)
        if (res.ok) {
          const data = await res.json()
          setArenas(data)
        }
      } catch (err) {
        console.error("Failed to fetch arenas:", err)
      } finally {
        setLoading(false)
      }
    }
    fetchArenas()
  }, [])

  useEffect(() => {
    if (!joining) return
    if (countdown <= 0) {
      router.push(`/arena/${joining}/play`)
      return
    }
    const timer = setTimeout(() => setCountdown((c) => c - 1), 1000)
    return () => clearTimeout(timer)
  }, [joining, countdown, router])

  function handleJoin(id: string) {
    setJoining(id)
    setCountdown(3)
  }

  return (
    <div className="relative min-h-screen">
      <div className="absolute inset-0 grid-bg opacity-40" />
      <div className="absolute top-0 right-0 w-[400px] h-[400px] rounded-full bg-neon-pink/5 blur-[120px]" />

      <div className="relative z-10 mx-auto max-w-7xl px-4 py-8 lg:px-8">
        {/* Header */}
        <div className="flex items-center gap-3 mb-2">
          <Radio className="h-4 w-4 text-neon-pink animate-pulse-glow" />
          <span className="font-mono text-[10px] tracking-[0.3em] text-neon-pink">
            LIVE ARENAS
          </span>
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl mb-2">
          SELECT YOUR <span className="text-neon-pink text-glow-pink">BATTLEFIELD</span>
        </h1>
        <p className="text-sm text-muted-foreground mb-8">
          Choose an arena and prepare for battle. Each round lasts 5 minutes with 10 rapid-fire questions.
        </p>

        {/* Joining overlay */}
        {joining && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-deep-bg/90 backdrop-blur-sm">
            <div className="flex flex-col items-center gap-6 text-center">
              <div className="relative h-32 w-32 flex items-center justify-center">
                <div className="absolute inset-0 border-2 border-neon-cyan/30 animate-spin" style={{ animationDuration: "3s" }} />
                <div className="absolute inset-3 border border-neon-pink/20 animate-spin" style={{ animationDuration: "2s", animationDirection: "reverse" }} />
                <span className="font-mono text-5xl font-bold text-neon-cyan text-glow-cyan">
                  {countdown}
                </span>
              </div>
              <div>
                <span className="block font-mono text-[11px] tracking-[0.3em] text-neon-pink mb-2">
                  ENTERING ARENA
                </span>
                <span className="font-mono text-lg font-bold tracking-wider text-foreground">
                  PREPARE FOR COMBAT
                </span>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-1 w-1 rounded-full bg-neon-cyan animate-pulse" />
                <span className="font-mono text-[10px] tracking-widest text-muted-foreground">
                  LOADING QUESTIONS
                </span>
                <div className="h-1 w-1 rounded-full bg-neon-cyan animate-pulse" />
              </div>
            </div>
          </div>
        )}

        {/* Loading state */}
        {loading && (
          <div className="flex flex-col gap-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-24 w-full animate-pulse border border-panel-border bg-panel-bg/20" />
            ))}
          </div>
        )}

        {/* Arena list */}
        <div className="flex flex-col gap-3">
          {arenas.map((arena) => {
            const diffKey = (arena.difficulty || "WARRIOR").toUpperCase()
            const diff = difficultyConfig[diffKey] || difficultyConfig.WARRIOR
            const statusKey = (arena.status || "open").toLowerCase()
            const stat = statusConfig[statusKey] || statusConfig.open
            const isFull = arena.currentPlayers >= arena.maxPlayers
            const DiffIcon = diff.icon

            return (
              <div
                key={arena.id}
                className={`group relative border border-panel-border bg-panel-bg/60 transition-all ${
                  isFull ? "opacity-60" : "hover:border-panel-border/80 hover:bg-panel-bg/80"
                }`}
              >
                <div className="flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
                  {/* Left info */}
                  <div className="flex items-center gap-4 flex-1 min-w-0">
                    <div className={`flex h-10 w-10 items-center justify-center border border-panel-border ${diff.color}`}>
                      <DiffIcon className="h-5 w-5" strokeWidth={1.5} />
                    </div>
                    <div className="min-w-0">
                      <div className="flex items-center gap-3 flex-wrap">
                        <span className="font-mono text-sm font-bold tracking-wider text-foreground">
                          {arena.title.toUpperCase()}
                        </span>
                        <span className={`font-mono text-[10px] tracking-[0.2em] ${diff.color}`}>
                          {arena.difficulty}
                        </span>
                      </div>
                      <span className="font-mono text-[10px] tracking-wider text-muted-foreground uppercase">
                        {arena.category?.name || "GENERAL"}
                      </span>
                    </div>
                  </div>

                  {/* Status indicators */}
                  <div className="flex items-center gap-6 flex-wrap">
                    <div className="flex items-center gap-2">
                      <Users className="h-3.5 w-3.5 text-muted-foreground" />
                      <span className="font-mono text-xs text-foreground">
                        {arena.currentPlayers}/{arena.maxPlayers}
                      </span>
                    </div>

                    <div className="flex items-center gap-2">
                      <div className={`h-1.5 w-1.5 rounded-full ${stat.bg} ${arena.status === "live" ? "animate-pulse-glow" : ""}`} />
                      <span className={`font-mono text-[10px] tracking-wider ${stat.color} uppercase`}>
                        {arena.status}
                      </span>
                    </div>

                    {!isFull && (
                      <div className="flex items-center gap-1.5">
                        <Clock className="h-3 w-3 text-muted-foreground" />
                        <span className="font-mono text-[10px] text-muted-foreground uppercase">
                          {arena.durationSeconds}S
                        </span>
                      </div>
                    )}

                    {/* Join button */}
                    {isFull ? (
                      <span className="font-mono text-[10px] tracking-wider text-muted-foreground px-4 py-2 border border-panel-border">
                        FULL
                      </span>
                    ) : (
                      <button
                        onClick={() => handleJoin(arena.id)}
                        className="group/btn flex items-center gap-2 border border-neon-cyan/50 bg-neon-cyan/10 px-5 py-2 font-mono text-[11px] tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_15px_rgba(0,240,255,0.15)]"
                      >
                        <Swords className="h-3.5 w-3.5" />
                        JOIN
                        <ChevronRight className="h-3 w-3 transition-transform group-hover/btn:translate-x-0.5" />
                      </button>
                    )}
                  </div>
                </div>

                {/* Fill bar */}
                <div className="h-0.5 bg-panel-border">
                  <div
                    className={`h-full ${isFull ? "bg-neon-pink/40" : "bg-neon-cyan/40"} transition-all`}
                    style={{ width: `${(arena.currentPlayers / arena.maxPlayers) * 100}%` }}
                  />
                </div>
              </div>
            )
          })}
        </div>

        {/* Bottom info */}
        <div className="mt-8 flex items-center gap-4">
          <div className="h-px flex-1 bg-panel-border" />
          <div className="flex items-center gap-2">
            <Zap className="h-3 w-3 text-muted-foreground" />
            <span className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">
              {arenas.length} ARENAS AVAILABLE
            </span>
          </div>
          <div className="h-px flex-1 bg-panel-border" />
        </div>
      </div>
    </div>
  )
}

