"use client"

import { useEffect, useState } from "react"
import { BarChart3, FileText, FolderOpen, Shield, Users } from "lucide-react"
import { API_URL } from "@/lib/api-config"

const API = API_URL

interface AdminStats {
  totalTests: number
  publishedTests: number
  activeTests: number
  totalTopics: number
  totalAttempts: number
  totalUsers: number
}

export default function AdminAnalyticsPage() {
  const [stats, setStats] = useState<AdminStats | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchStats() {
      try {
        const res = await fetch(`${API}/api/admin/analytics`, { credentials: "include" })
        if (res.ok) {
          setStats(await res.json())
        }
      } catch (e) {
        console.error("Analytics fetch failed:", e)
      } finally {
        setLoading(false)
      }
    }
    fetchStats()
  }, [])

  const statCards = stats
    ? [
        { label: "TOTAL TESTS", value: stats.totalTests, icon: FileText, color: "neon-cyan" },
        { label: "PUBLISHED", value: stats.publishedTests, icon: FileText, color: "neon-green" },
        { label: "ACTIVE NOW", value: stats.activeTests, icon: BarChart3, color: "neon-pink" },
        { label: "TOPICS", value: stats.totalTopics, icon: FolderOpen, color: "neon-amber" },
        { label: "ATTEMPTS", value: stats.totalAttempts, icon: BarChart3, color: "neon-cyan" },
        { label: "USERS", value: stats.totalUsers, icon: Users, color: "neon-pink" },
      ]
    : []

  return (
    <div className="relative min-h-screen">
      <div className="absolute inset-0 grid-bg opacity-20" />

      <div className="relative z-10 px-8 py-8">
        <div className="flex items-center gap-3 mb-2">
          <Shield className="h-4 w-4 text-neon-cyan animate-pulse-glow" />
          <span className="font-mono text-[10px] tracking-[0.3em] text-neon-cyan">
            ADMIN ANALYTICS
          </span>
        </div>
        <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl mb-8">
          PLATFORM <span className="text-neon-cyan text-glow-cyan">OVERVIEW</span>
        </h1>

        {loading ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div key={i} className="h-28 animate-pulse border border-panel-border bg-panel-bg/20" />
            ))}
          </div>
        ) : stats ? (
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
            {statCards.map((card) => {
              const Icon = card.icon
              return (
                <div
                  key={card.label}
                  className="border border-panel-border bg-panel-bg/60 backdrop-blur-sm p-5 hover:border-neon-cyan/30 transition-colors"
                >
                  <div className="flex items-center gap-2 mb-3">
                    <Icon className={`h-3.5 w-3.5 text-${card.color}`} />
                    <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                      {card.label}
                    </span>
                  </div>
                  <span className={`block text-3xl font-bold font-mono text-${card.color}`}>
                    {card.value}
                  </span>
                </div>
              )
            })}
          </div>
        ) : (
          <div className="border border-panel-border bg-panel-bg/40 p-10 text-center">
            <span className="font-mono text-xs text-muted-foreground">
              UNABLE TO LOAD ANALYTICS
            </span>
          </div>
        )}

        {/* Common Mistakes Section (NEW) */}
        {!loading && (
          <div className="mt-12">
            <div className="flex items-center gap-3 mb-6">
              <BarChart3 className="h-4 w-4 text-neon-pink" />
              <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                COMMON MISTAKES ANALYTICS
              </span>
            </div>

            <MistakesTable />
          </div>
        )}
      </div>
    </div>
  )
}

function MistakesTable() {
  const [mistakes, setMistakes] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchMistakes() {
      try {
        const res = await fetch(`${API_URL}/api/admin/analytics/mistakes`, { credentials: "include" })
        if (res.ok) setMistakes(await res.json())
      } catch (e) {
        console.error("Mistakes fetch failed:", e)
      } finally {
        setLoading(false)
      }
    }
    fetchMistakes()
  }, [])

  if (loading) return <div className="h-40 animate-pulse border border-panel-border bg-panel-bg/20" />

  return (
    <div className="border border-panel-border bg-panel-bg/40">
      <div className="grid grid-cols-[1fr_120px_100px_100px] gap-4 px-5 py-3 border-b border-panel-border bg-panel-bg/60 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
        <span>QUESTION TITLE</span>
        <span>DOMAIN</span>
        <span className="text-right">FAILURES</span>
        <span className="text-right">FAILURE RATE</span>
      </div>
      <div className="flex flex-col">
        {mistakes.map((m, i) => (
          <div key={i} className="grid grid-cols-[1fr_120px_100px_100px] gap-4 px-5 py-4 border-b border-panel-border/30 last:border-0 hover:bg-neon-pink/5 transition-colors items-center">
            <span className="font-mono text-sm font-bold text-foreground uppercase truncate">{m.questionTitle}</span>
            <span className="font-mono text-[10px] text-neon-amber uppercase">{m.topicId}</span>
            <span className="font-mono text-sm font-bold text-neon-pink text-right">{m.failureCount}</span>
            <span className="font-mono text-sm font-bold text-neon-pink text-right">{m.failureRate ? m.failureRate.toFixed(1) : 0}%</span>
          </div>
        ))}
        {mistakes.length === 0 && (
          <div className="p-10 text-center font-mono text-[10px] text-muted-foreground uppercase">
            NO ANALYTICS DATA RECORDED YET.
          </div>
        )}
      </div>
    </div>
  )
}

