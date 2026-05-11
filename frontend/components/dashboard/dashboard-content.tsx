"use client"

import { useEffect, useState } from "react"
import {
  Activity,
  Award,
  BarChart3,
  Brain,
  ChevronRight,
  Clock,
  Crown,
  Flame,
  Loader2,
  Shield,
  Skull,
  Star,
  Swords,
  Target,
  TrendingDown,
  TrendingUp,
  Trophy,
  Zap,
} from "lucide-react"
import Link from "next/link"
import { PerformanceChart } from "./performance-chart"
import { API_URL } from "@/lib/api-config"

interface DashboardData {
  stats: {
    totalAttempts: number
    highScore: number
    avgScore: number
    totalScore: number
    unmasteredMistakes: number
  }
  globalRank: number
  totalParticipants: number
  tier: string
  topicAnalysis: Array<{
    topicId: string
    topicName: string
    testsTaken: number
    avgScore: number
    maxScore: number
    totalScore: number
    maxPossible: number
    percentage: number
  }>
  weakTopicStats: Array<{
    topicId: string
    topicName: string
    accuracyPercent: number
    weakLevel: string
  }>
  strongPoints: string[]
  weakPoints: string[]
  performanceTrend: Array<{
    testTitle: string
    score: number
    percentage: number
    date: string
  }>
  completedTests: Array<{
    attemptId: string
    testId: string
    testTitle: string
    topicName: string
    score: number
    maxPossible: number
    percentage: number
    isAutoSubmitted: boolean
    submittedAt: string
  }>
  activeTestCount: number
}

const tierConfig: Record<string, { icon: React.ElementType; color: string; glowClass: string }> = {
  APEX: { icon: Crown, color: "text-neon-amber", glowClass: "text-glow-amber" },
  CHAMPION: { icon: Skull, color: "text-neon-pink", glowClass: "text-glow-pink" },
  VETERAN: { icon: Flame, color: "text-neon-amber", glowClass: "text-glow-amber" },
  ELITE: { icon: Star, color: "text-neon-cyan", glowClass: "text-glow-cyan" },
  WARRIOR: { icon: Shield, color: "text-muted-foreground", glowClass: "" },
  ROOKIE: { icon: Target, color: "text-muted-foreground", glowClass: "" },
  UNRANKED: { icon: Target, color: "text-muted-foreground", glowClass: "" },
}


export function DashboardContent() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [user, setUser] = useState<{ username?: string; email?: string } | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchDashboard = async () => {
      try {
        // Fetch user info
        const meRes = await fetch(`${API_URL}/api/auth/me`, { credentials: "include" })
        if (meRes.ok) {
          const meData = await meRes.json()
          setUser({ username: meData.username, email: meData.email })
        }

        // Fetch full dashboard
        const res = await fetch(`${API_URL}/api/dashboard/full`, { credentials: "include" })
        if (res.ok) {
          const json = await res.json()
          setData(json)
        }
      } catch (err) {
        console.error("Dashboard Fetch Error:", err)
      } finally {
        setLoading(false)
      }
    }
    fetchDashboard()
  }, [])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-4">
        <Loader2 className="h-8 w-8 text-neon-cyan animate-spin" />
        <span className="font-mono text-xs tracking-widest text-neon-cyan">LOADING COMMAND CENTER...</span>
      </div>
    )
  }

  if (!data || !user) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[60vh] gap-6 text-center">
        <Shield className="h-12 w-12 text-neon-pink mb-2" />
        <h2 className="text-xl font-bold text-foreground">ACCESS DENIED</h2>
        <p className="text-muted-foreground text-sm max-w-xs">Please login to access your command center and track your progress.</p>
        <Link href="/login" className="px-8 py-3 bg-neon-cyan font-mono text-xs font-bold text-deep-bg tracking-widest hover:bg-neon-cyan/90 transition-all">LOGIN NOW</Link>
      </div>
    )
  }

  const displayName = user.username && !user.username.includes('@') 
    ? user.username 
    : (user.email?.split('@')[0] || "OPERATOR")

  const tierInfo = tierConfig[data.tier] || tierConfig.UNRANKED
  const TierIcon = tierInfo.icon

  return (
    <div className="relative min-h-screen">
      <div className="absolute inset-0 grid-bg opacity-40" />

      <div className="relative z-10 mx-auto max-w-7xl px-4 py-8 lg:px-8">
        {/* Header */}
        <div className="flex flex-col gap-4 mb-8 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <div className="h-2 w-2 rounded-full bg-neon-cyan animate-pulse-glow" />
              <span className="font-mono text-[10px] tracking-[0.3em] text-neon-cyan uppercase">
                COMMAND CENTER // ONLINE
              </span>
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
              WELCOME BACK, <span className="text-neon-cyan text-glow-cyan uppercase">{displayName}</span>
            </h1>
          </div>
          <Link
            href="/arena"
            className="group flex items-center gap-3 border border-neon-cyan bg-neon-cyan/10 px-6 py-3 font-mono text-xs tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_20px_rgba(0,240,255,0.2)]"
          >
            <Swords className="h-4 w-4" />
            ENTER ARENA
            <ChevronRight className="h-3 w-3 transition-transform group-hover:translate-x-1" />
          </Link>
        </div>

        {/* Top stats row */}
        <div className="grid gap-3 grid-cols-2 lg:grid-cols-4 mb-6">
          <StatCard
            label="TESTS COMPLETED"
            value={data.stats.totalAttempts.toString()}
            icon={Activity}
            color="cyan"
            sub="SUBMITTED"
          />
          <StatCard
            label="PEAK SCORE"
            value={data.stats.highScore.toString()}
            icon={Crown}
            color="amber"
            sub="PERSONAL BEST"
          />
          <StatCard
            label="AVG PERFORMANCE"
            value={data.stats.avgScore ? data.stats.avgScore.toFixed(1) : "0"}
            icon={Target}
            color="pink"
            sub="MEAN SCORE"
          />
          <StatCard
            label="TOTAL SCORE"
            value={data.stats.totalScore.toString()}
            icon={Flame}
            color="cyan"
            sub="CUMULATIVE XP"
          />
        </div>

        <div className="grid gap-6 lg:grid-cols-3 mb-6">
          {/* Chart (Left) */}
          <div className="lg:col-span-2 border border-panel-border bg-panel-bg/60 p-6 flex flex-col">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <Activity className="h-4 w-4 text-neon-cyan" />
                <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                  SCORE PROGRESSION
                </span>
              </div>
            </div>
            <div className="flex-1 min-h-[300px]">
              <PerformanceChart data={data.performanceTrend} />
            </div>
          </div>

          {/* Sidebar (Right) */}
          <div className="flex flex-col gap-6">
            {/* Rank Card */}
            <div className="flex-1 border border-panel-border bg-panel-bg/60 p-6">
              <div className="flex items-center gap-3 mb-6">
                <Zap className="h-4 w-4 text-neon-cyan" />
                <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                  RANKING SYSTEM
                </span>
              </div>
              <div className="flex flex-col gap-4">
                <div className="group p-4 border border-neon-yellow/20 bg-neon-yellow/5 transition-all duration-300 hover:border-neon-yellow/60 hover:shadow-[0_0_20px_rgba(251,191,36,0.1)]">
                  <span className="block font-mono text-[9px] text-neon-yellow/60 group-hover:text-neon-yellow mb-1 uppercase transition-colors">CURRENT TIER</span>
                  <div className="flex items-center gap-3">
                    <TierIcon className={`h-5 w-5 ${tierInfo.color}`} />
                    <span className={`font-mono text-xl font-bold ${tierInfo.color} ${tierInfo.glowClass} uppercase transition-all`}>{data.tier}</span>
                  </div>
                </div>
                {data.globalRank > 0 && (
                  <div className="flex items-center justify-between px-4 py-3 border border-panel-border/50 bg-secondary/10">
                    <span className="font-mono text-[10px] text-muted-foreground uppercase">GLOBAL RANK</span>
                    <span className="font-mono text-lg font-bold text-neon-cyan">#{data.globalRank}<span className="text-xs text-muted-foreground ml-1">/ {data.totalParticipants}</span></span>
                  </div>
                )}
                <p className="text-[10px] font-mono text-muted-foreground leading-relaxed">
                  Complete more tests with high efficiency to upgrade your classification status.
                </p>
              </div>
            </div>

            <Link
              href="/leaderboard"
              className="flex-1 flex items-center justify-center gap-3 border border-panel-border bg-panel-bg/40 px-8 py-4 font-mono text-xs tracking-widest text-muted-foreground hover:border-foreground/20 hover:text-foreground transition-all uppercase"
            >
              <Trophy className="h-4 w-4" />
              VIEW LEADERBOARD
            </Link>
          </div>
        </div>

        {/* Adaptive Practice Section (NEW) */}
        <div className="grid gap-6 lg:grid-cols-3 mb-6">
          <div className="lg:col-span-2 border border-panel-border bg-panel-bg/60 p-6">
            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <Brain className="h-4 w-4 text-neon-cyan" />
                <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                  ADAPTIVE INTELLIGENCE
                </span>
              </div>
              <div className="px-2 py-0.5 border border-neon-cyan/20 bg-neon-cyan/5">
                <span className="font-mono text-[9px] text-neon-cyan uppercase tracking-wider">AI POWERED</span>
              </div>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <div className="border border-panel-border/30 bg-secondary/5 p-4 flex flex-col justify-between">
                <div>
                  <h3 className="font-mono text-sm font-bold text-foreground mb-1 uppercase">PRACTICE MISTAKES</h3>
                  <p className="text-[10px] text-muted-foreground font-mono leading-relaxed mb-4">
                    Target your weaknesses by retrying {data.stats.unmasteredMistakes || 0} questions you previously answered incorrectly.
                  </p>
                </div>
                <Link
                  href="/train?mode=mistakes"
                  className="w-full py-2 bg-neon-pink/10 border border-neon-pink/30 text-neon-pink font-mono text-[10px] tracking-widest text-center hover:bg-neon-pink/20 transition-all uppercase"
                >
                  START RECOVERY SESSION
                </Link>
              </div>

              <div className="border border-neon-cyan/20 bg-neon-cyan/5 p-4 flex flex-col justify-between">
                <div>
                  <h3 className="font-mono text-sm font-bold text-neon-cyan mb-1 uppercase">ADAPTIVE TRAINING</h3>
                  <p className="text-[10px] text-muted-foreground font-mono leading-relaxed mb-4">
                    SkillSprint AI will generate a personalized path with similar questions based on your weak areas.
                  </p>
                </div>
                <Link
                  href="/train?mode=adaptive"
                  className="w-full py-2 bg-neon-cyan text-deep-bg font-mono text-[10px] font-bold tracking-widest text-center hover:bg-neon-cyan/90 transition-all uppercase"
                >
                  START ADAPTIVE SESSION
                </Link>
              </div>
            </div>
          </div>

          <div className="border border-panel-border bg-panel-bg/60 p-6">
            <div className="flex items-center gap-3 mb-6">
              <TrendingDown className="h-4 w-4 text-neon-pink" />
              <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                WEAK AREAS
              </span>
            </div>
            <div className="flex flex-col gap-3">
              {data.weakTopicStats?.map((topic) => (
                <div key={topic.topicId} className="p-3 border border-panel-border/30 bg-secondary/5">
                  <div className="flex items-center justify-between mb-1">
                    <span className="font-mono text-[11px] font-bold text-foreground uppercase">{topic.topicName}</span>
                    <span className="font-mono text-[10px] text-neon-pink uppercase">{topic.accuracyPercent.toFixed(0)}% ACC</span>
                  </div>
                  <div className="w-full h-1 bg-panel-border/30 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-neon-pink" 
                      style={{ width: `${topic.accuracyPercent}%` }}
                    />
                  </div>
                </div>
              ))}
              {(!data.weakTopicStats || data.weakTopicStats.length === 0) && (
                <p className="font-mono text-[10px] text-muted-foreground text-center py-4">NO CRITICAL WEAK AREAS IDENTIFIED YET.</p>
              )}
            </div>
          </div>
        </div>

        {/* Topic Analysis + Strengths/Weaknesses Row */}
        {data.topicAnalysis.length > 0 && (
          <div className="grid gap-6 lg:grid-cols-3 mb-6">
            {/* Topic Breakdown */}
            <div className="lg:col-span-2 border border-panel-border bg-panel-bg/60 p-6">
              <div className="flex items-center gap-3 mb-6">
                <BarChart3 className="h-4 w-4 text-neon-cyan" />
                <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                  TOPIC-WISE ANALYSIS
                </span>
              </div>
              <div className="flex flex-col gap-3">
                {data.topicAnalysis.map((topic) => (
                  <div key={topic.topicId || topic.topicName} className="flex items-center gap-4 px-4 py-3 border border-panel-border/30 bg-secondary/5 hover:bg-secondary/10 transition-all">
                    <div className="flex-1 min-w-0">
                      <span className="block font-mono text-xs font-bold text-foreground uppercase truncate">{topic.topicName}</span>
                      <span className="font-mono text-[10px] text-muted-foreground">{topic.testsTaken} test{topic.testsTaken !== 1 ? 's' : ''} // Avg: {topic.avgScore.toFixed(1)}</span>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="w-32 h-2 bg-panel-border/30 overflow-hidden">
                        <div
                          className="h-full transition-all duration-500"
                          style={{
                            width: `${Math.min(topic.percentage, 100)}%`,
                            backgroundColor: topic.percentage >= 70 ? '#00f0ff' : topic.percentage >= 40 ? '#fbbf24' : '#ff2d6f',
                          }}
                        />
                      </div>
                      <span className={`font-mono text-sm font-bold min-w-[48px] text-right ${
                        topic.percentage >= 70 ? 'text-neon-cyan' : topic.percentage >= 40 ? 'text-neon-yellow' : 'text-neon-pink'
                      }`}>
                        {topic.percentage.toFixed(0)}%
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Strengths & Weaknesses */}
            <div className="flex flex-col gap-6">
              {/* Strong Points */}
              <div className="flex-1 border border-panel-border bg-panel-bg/60 p-6">
                <div className="flex items-center gap-3 mb-4">
                  <TrendingUp className="h-4 w-4 text-neon-cyan" />
                  <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                    STRONG POINTS
                  </span>
                </div>
                {data.strongPoints.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    {data.strongPoints.map((point) => (
                      <div key={point} className="flex items-center gap-3 px-3 py-2 border border-neon-cyan/15 bg-neon-cyan/5">
                        <div className="h-1.5 w-1.5 rounded-full bg-neon-cyan" />
                        <span className="font-mono text-xs text-neon-cyan uppercase">{point}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="font-mono text-[10px] text-muted-foreground">Complete more tests to identify strengths.</p>
                )}
              </div>

              {/* Weak Points */}
              <div className="flex-1 border border-panel-border bg-panel-bg/60 p-6">
                <div className="flex items-center gap-3 mb-4">
                  <TrendingDown className="h-4 w-4 text-neon-pink" />
                  <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                    NEEDS IMPROVEMENT
                  </span>
                </div>
                {data.weakPoints.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    {data.weakPoints.map((point) => (
                      <div key={point} className="flex items-center gap-3 px-3 py-2 border border-neon-pink/15 bg-neon-pink/5">
                        <div className="h-1.5 w-1.5 rounded-full bg-neon-pink" />
                        <span className="font-mono text-xs text-neon-pink uppercase">{point}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="font-mono text-[10px] text-muted-foreground">Keep performing well across all topics!</p>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Recent Completed Tests */}
        <div className="border border-panel-border bg-panel-bg/60 p-6">
          <div className="flex items-center gap-3 mb-6">
            <Clock className="h-4 w-4 text-neon-pink" />
            <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
              RECENT BATTLE LOGS
            </span>
          </div>

          <div className="flex flex-col gap-2">
            {data.completedTests.map((test) => (
              <Link
                key={test.attemptId}
                href={`/results/${test.attemptId}`}
                className="flex items-center gap-4 border border-panel-border/50 bg-secondary/20 px-4 py-3 transition-all hover:bg-secondary/30"
              >
                <div className="flex h-8 w-8 items-center justify-center border border-neon-cyan/30 text-neon-cyan">
                  <Trophy className="h-3.5 w-3.5" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs font-bold tracking-wider text-foreground truncate uppercase">
                      {test.testTitle}
                    </span>
                    <span className="font-mono text-[9px] px-2 py-0.5 border border-panel-border/30 text-muted-foreground uppercase">
                      {test.topicName}
                    </span>
                    {test.isAutoSubmitted && (
                      <span className="font-mono text-[9px] px-2 py-0.5 bg-neon-pink/10 border border-neon-pink/20 text-neon-pink uppercase">
                        AUTO
                      </span>
                    )}
                  </div>
                  <span className="font-mono text-[10px] text-muted-foreground uppercase">
                    Score: {test.score}/{test.maxPossible} ({test.percentage.toFixed(0)}%) // {test.submittedAt ? new Date(test.submittedAt).toLocaleDateString() : 'N/A'}
                  </span>
                </div>
                <span className={`font-mono text-sm font-bold ${
                  test.percentage >= 70 ? 'text-neon-cyan' : test.percentage >= 40 ? 'text-neon-yellow' : 'text-neon-pink'
                }`}>
                  {test.percentage.toFixed(0)}%
                </span>
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              </Link>
            ))}
            {data.completedTests.length === 0 && (
              <div className="py-10 text-center border border-dashed border-panel-border text-muted-foreground font-mono text-xs uppercase">
                NO BATTLE LOGS FOUND. ENTER THE ARENA TO START.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function StatCard({
  label,
  value,
  icon: Icon,
  color,
  sub,
}: {
  label: string
  value: string
  icon: React.ElementType
  color: "cyan" | "pink" | "amber"
  sub: string
}) {
  const colorClasses = {
    cyan: {
      border: "border-neon-cyan/20 group-hover:border-neon-cyan",
      text: "text-neon-cyan/70 group-hover:text-neon-cyan",
      glow: "group-hover:text-glow-cyan",
      shadow: "group-hover:shadow-[0_0_20px_rgba(0,240,255,0.15)]"
    },
    pink: {
      border: "border-neon-pink/20 group-hover:border-neon-pink",
      text: "text-neon-pink/70 group-hover:text-neon-pink",
      glow: "group-hover:text-glow-pink",
      shadow: "group-hover:shadow-[0_0_20px_rgba(255,45,111,0.15)]"
    },
    amber: {
      border: "border-neon-yellow/20 group-hover:border-neon-yellow",
      text: "text-neon-yellow/70 group-hover:text-neon-yellow",
      glow: "group-hover:text-glow-yellow",
      shadow: "group-hover:shadow-[0_0_20px_rgba(255,184,0,0.15)]"
    }
  }

  const fallbackStyle = {
    border: "border-panel-border",
    text: "text-muted-foreground",
    glow: "",
    shadow: ""
  }

  const current = colorClasses[color as keyof typeof colorClasses] ?? fallbackStyle

  return (
    <div className={`group border ${current.border} bg-panel-bg/60 p-4 transition-all duration-300 hover:bg-panel-bg/80 ${current.shadow}`}>
      <div className="flex items-center justify-between mb-3">
        <span className="font-mono text-[10px] tracking-[0.2em] text-muted-foreground uppercase">
          {label}
        </span>
        <Icon className={`h-4 w-4 ${current.text.split(' ')[0]} transition-colors`} />
      </div>
      <div className={`font-mono text-2xl font-bold ${current.text} ${current.glow} transition-all`}>
        {value}
      </div>
      <span className="font-mono text-[10px] tracking-wider text-muted-foreground uppercase">
        {sub}
      </span>
    </div>
  )
}

