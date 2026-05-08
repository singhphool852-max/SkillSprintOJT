"use client"

import { useEffect, useState } from "react"
import {
  Award,
  Crown,
  Flame,
  Loader2,
  Shield,
  Skull,
  Star,
  Target,
  Trophy,
} from "lucide-react"
import { API_URL } from "@/lib/api-config"

interface LeaderboardEntry {
  rank: number
  userId: string
  username: string
  totalScore: number
  testsCompleted: number
  avgPercentage: number
  highScore: number
  tier: string
}

interface LeaderboardData {
  entries: LeaderboardEntry[]
  totalUsers: number
}

const tierConfig: Record<string, { icon: React.ElementType; color: string; borderColor: string }> = {
  APEX: { icon: Crown, color: "text-neon-amber", borderColor: "border-neon-amber/30" },
  CHAMPION: { icon: Skull, color: "text-neon-pink", borderColor: "border-neon-pink/30" },
  VETERAN: { icon: Flame, color: "text-neon-amber", borderColor: "border-neon-amber/20" },
  ELITE: { icon: Star, color: "text-neon-cyan", borderColor: "border-neon-cyan/20" },
  WARRIOR: { icon: Shield, color: "text-neon-cyan", borderColor: "border-neon-cyan/15" },
  ROOKIE: { icon: Target, color: "text-muted-foreground", borderColor: "border-panel-border" },
}

export function LeaderboardContent() {
  const [data, setData] = useState<LeaderboardData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchLeaderboard() {
      try {
        const res = await fetch(`${API_URL}/api/leaderboard/global`)
        if (res.ok) {
          const json = await res.json()
          setData(json)
        }
      } catch (err) {
        console.error(err)
      } finally {
        setLoading(false)
      }
    }
    fetchLeaderboard()
  }, [])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 text-neon-cyan animate-spin" />
        <span className="font-mono text-xs tracking-widest text-neon-cyan">CALCULATING DOMINANCE...</span>
      </div>
    )
  }

  const entries = data?.entries || []

  return (
    <div className="relative min-h-screen pb-20">
      <div className="absolute inset-0 grid-bg opacity-30" />
      
      <div className="relative z-10 mx-auto max-w-5xl px-4 py-8">
        {/* Header */}
        <div className="flex items-center gap-3 mb-2">
          <Trophy className="h-4 w-4 text-neon-amber" />
          <span className="font-mono text-[10px] tracking-[0.3em] text-neon-amber uppercase">
            GLOBAL RANKINGS
          </span>
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl mb-2">
          THE <span className="text-neon-amber text-glow-amber">ELITE</span> OPERATIVES
        </h1>
        {data && data.totalUsers > 0 && (
          <p className="font-mono text-[10px] text-muted-foreground mb-10 uppercase">
            {data.totalUsers} OPERATIVE{data.totalUsers !== 1 ? 'S' : ''} RANKED // SORTED BY TOTAL SCORE
          </p>
        )}

        {/* Top 3 podium */}
        {entries.length > 0 && (
          <div className="grid gap-4 sm:grid-cols-3 mb-12">
            {entries.slice(0, 3).map((player, i) => {
              const tier = tierConfig[player.tier] || tierConfig.ROOKIE
              const TierIcon = tier.icon
              
              return (
                <div
                  key={player.userId}
                  className={`relative border p-8 text-center bg-panel-bg/40 ${
                    i === 0 ? "border-neon-amber/50 scale-105 shadow-[0_0_30px_rgba(255,180,0,0.1)]" : "border-panel-border"
                  }`}
                >
                  <div className="flex items-center justify-center mb-4">
                    <div className={`flex h-12 w-12 items-center justify-center border ${tier.borderColor} ${tier.color}`}>
                      <TierIcon className="h-6 w-6" />
                    </div>
                  </div>
                  <span className={`block font-mono text-3xl font-black ${
                    i === 0 ? "text-neon-amber text-glow-amber" : "text-foreground"
                  } mb-1`}>
                    #{i + 1}
                  </span>
                  <span className="block font-mono text-sm font-bold tracking-widest text-foreground mb-1 uppercase">
                    {player.username}
                  </span>
                  <span className={`block font-mono text-lg font-bold ${tier.color}`}>
                    {player.totalScore} XP
                  </span>
                  <div className="mt-2 flex items-center justify-center gap-4 font-mono text-[10px] text-muted-foreground">
                    <span>{player.testsCompleted} TEST{player.testsCompleted !== 1 ? 'S' : ''}</span>
                    <span>•</span>
                    <span>BEST: {player.highScore}</span>
                  </div>
                </div>
              )
            })}
          </div>
        )}

        {/* Full List */}
        <div className="border border-panel-border bg-panel-bg/20 overflow-hidden">
          <div className="flex items-center gap-4 bg-panel-bg/60 border-b border-panel-border px-6 py-4">
            <span className="w-12 font-mono text-[10px] tracking-widest text-muted-foreground uppercase">RANK</span>
            <span className="flex-1 font-mono text-[10px] tracking-widest text-muted-foreground uppercase">OPERATIVE</span>
            <span className="w-16 text-center font-mono text-[10px] tracking-widest text-muted-foreground uppercase hidden sm:block">TESTS</span>
            <span className="w-16 text-center font-mono text-[10px] tracking-widest text-muted-foreground uppercase hidden sm:block">BEST</span>
            <span className="w-16 text-center font-mono text-[10px] tracking-widest text-muted-foreground uppercase hidden sm:block">TIER</span>
            <span className="w-24 text-right font-mono text-[10px] tracking-widest text-muted-foreground uppercase">TOTAL SCORE</span>
          </div>

          <div className="divide-y divide-panel-border/30">
            {entries.map((player) => {
              const tier = tierConfig[player.tier] || tierConfig.ROOKIE
              return (
                <div
                  key={player.userId}
                  className="flex items-center gap-4 px-6 py-4 transition-all hover:bg-secondary/10"
                >
                <span className={`w-12 font-mono text-sm font-bold ${
                  player.rank <= 3 ? "text-neon-amber" : "text-muted-foreground"
                }`}>
                  #{player.rank}
                </span>
                <div className="flex-1 flex items-center gap-3">
                   <div className={`h-2 w-2 rounded-full ${
                     player.rank <= 3 ? 'bg-neon-amber' : player.rank <= 10 ? 'bg-neon-cyan/60' : 'bg-neon-cyan/30'
                   }`} />
                   <span className="font-mono text-xs font-bold tracking-wider text-foreground uppercase">
                     {player.username}
                   </span>
                </div>
                <span className="w-16 text-center font-mono text-xs text-muted-foreground hidden sm:block">
                  {player.testsCompleted}
                </span>
                <span className="w-16 text-center font-mono text-xs text-muted-foreground hidden sm:block">
                  {player.highScore}
                </span>
                <span className={`w-16 text-center font-mono text-[10px] font-bold hidden sm:block ${tier.color} uppercase`}>
                  {player.tier}
                </span>
                <span className="w-24 text-right font-mono text-sm font-bold text-neon-cyan">
                  {player.totalScore}
                </span>
              </div>
              )
            })}
            {entries.length === 0 && (
              <div className="py-20 text-center text-muted-foreground font-mono text-xs">NO PERFORMANCE DATA RECORDED YET.</div>
            )}
          </div>
        </div>

        <div className="mt-12 p-6 border border-panel-border bg-panel-bg/40 text-center">
            <Award className="h-6 w-6 text-neon-cyan mx-auto mb-3" />
            <h3 className="font-mono text-sm font-bold text-foreground mb-2">WANT TO CLIMB THE RANKS?</h3>
            <p className="text-xs text-muted-foreground max-w-md mx-auto leading-relaxed">
              Every test in the arena contributes to your global rating. High scores across multiple tests are the key to reaching the APEX tier.
            </p>
        </div>
      </div>
    </div>
  )
}

