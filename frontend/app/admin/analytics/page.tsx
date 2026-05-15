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
  activeArenaUsers: number
}

// Error Boundary Component
function ErrorFallback({ error, resetErrorBoundary }: { error: Error; resetErrorBoundary: () => void }) {
  return (
    <div className="min-h-screen flex items-center justify-center p-8">
      <div className="border border-red-500/30 bg-red-950/20 p-8 max-w-lg">
        <h2 className="text-xl font-bold text-red-400 mb-4">Analytics Failed to Load</h2>
        <p className="text-sm text-gray-400 mb-4">{error.message}</p>
        <button
          onClick={resetErrorBoundary}
          className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-mono text-xs uppercase tracking-wider transition-colors"
        >
          Retry
        </button>
      </div>
    </div>
  )
}

export default function AdminAnalyticsPage() {
  const [stats, setStats] = useState<AdminStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function fetchStats() {
      try {
        console.log('[ANALYTICS] Fetching stats from:', `${API}/api/admin/analytics`)
        const res = await fetch(`${API}/api/admin/analytics`, { credentials: "include" })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        console.log('[ANALYTICS] Stats received:', data)
        setStats(data ?? {})
        setError(null)
      } catch (e) {
        console.error("[ANALYTICS] fetch failed:", e)
        setError(e instanceof Error ? e.message : "Failed to load analytics")
        // Set safe defaults so page doesn't crash
        setStats({
          totalTests: 0,
          publishedTests: 0,
          activeTests: 0,
          totalTopics: 0,
          totalAttempts: 0,
          totalUsers: 0,
          activeArenaUsers: 0,
        })
      } finally {
        setLoading(false)
      }
    }
    fetchStats()
  }, [])

  const statCards = stats
    ? [
        { label: "TOTAL TESTS", value: stats.totalTests ?? 0, icon: FileText, color: "neon-cyan" },
        { label: "PUBLISHED", value: stats.publishedTests ?? 0, icon: FileText, color: "neon-green" },
        { label: "ACTIVE NOW", value: stats.activeTests ?? 0, icon: BarChart3, color: "neon-pink" },
        { label: "TOPICS", value: stats.totalTopics ?? 0, icon: FolderOpen, color: "neon-amber" },
        { label: "ATTEMPTS", value: stats.totalAttempts ?? 0, icon: BarChart3, color: "neon-cyan" },
        { 
          label: "ACTIVE IN ARENA", 
          value: stats.activeArenaUsers ?? 0, 
          icon: Users, 
          color: "neon-pink",
          subtitle: `${stats.totalUsers ?? 0} total users`
        },
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

        {error && (
          <div className="mb-6 border border-yellow-500/30 bg-yellow-950/20 p-4">
            <p className="font-mono text-xs text-yellow-400">⚠ {error}</p>
          </div>
        )}

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
                  {card.subtitle && (
                    <div className="text-xs text-gray-500 mt-2 font-mono">
                      {card.subtitle}
                    </div>
                  )}
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

        {/* Anti-Cheat Violations Section */}
        {!loading && (
          <div className="mt-12">
            <div className="flex items-center gap-3 mb-6">
              <Shield className="h-4 w-4 text-neon-pink" />
              <span className="font-mono text-[11px] tracking-[0.2em] text-foreground uppercase">
                ANTI-CHEAT VIOLATIONS
              </span>
            </div>

            <ViolationsTable />
          </div>
        )}
      </div>
    </div>
  )
}

interface ViolationRow {
  userName: string
  userEmail: string
  testTitle: string
  violationCount: number
  fullscreenExits: number
  tabSwitches: number
  lastViolation: string
}

function ViolationsTable() {
  const [violations, setViolations] = useState<ViolationRow[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function fetchViolations() {
      try {
        console.log('[VIOLATIONS] Fetching from:', `${API_URL}/api/admin/analytics/mistakes`)
        const res = await fetch(`${API_URL}/api/admin/analytics/mistakes`, { credentials: "include" })
        if (!res.ok) {
          console.error('[VIOLATIONS] HTTP error:', res.status)
          throw new Error(`HTTP ${res.status}`)
        }
        const data = await res.json()
        console.log('[VIOLATIONS] Data received:', data)
        console.log('[VIOLATIONS] Mistakes array:', data?.mistakes)
        setViolations(data?.mistakes ?? [])
      } catch (e) {
        console.error("[VIOLATIONS] fetch failed:", e)
        setViolations([])
      } finally {
        setLoading(false)
      }
    }
    fetchViolations()
  }, [])

  if (loading) {
    return <div className="h-40 animate-pulse border border-panel-border bg-panel-bg/20" />
  }

  if (!violations || violations.length === 0) {
    return (
      <div className="border border-panel-border bg-panel-bg/40 p-10 text-center">
        <Shield className="h-8 w-8 text-gray-600 mx-auto mb-3" />
        <p className="font-mono text-[10px] text-muted-foreground uppercase">
          No violation data yet.
        </p>
        <p className="font-mono text-[9px] text-gray-600 mt-2">
          Data appears when students exit fullscreen or switch tabs during tests.
        </p>
      </div>
    )
  }

  return (
    <div className="border border-panel-border bg-panel-bg/40 overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-panel-border bg-panel-bg/60">
              <th className="text-left py-3 px-4 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                STUDENT
              </th>
              <th className="text-left py-3 px-4 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                TEST
              </th>
              <th className="text-center py-3 px-4 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                TOTAL WARNINGS
              </th>
              <th className="text-center py-3 px-4 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                FULLSCREEN EXITS
              </th>
              <th className="text-center py-3 px-4 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                TAB SWITCHES
              </th>
              <th className="text-left py-3 px-4 font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                LAST VIOLATION
              </th>
            </tr>
          </thead>
          <tbody>
            {violations.map((row, i) => (
              <tr
                key={i}
                className="border-b border-panel-border/30 last:border-0 hover:bg-neon-pink/5 transition-colors"
              >
                <td className="py-3 px-4">
                  <div className="font-medium text-white text-sm">{row.userName}</div>
                  <div className="text-xs text-gray-500 font-mono">{row.userEmail}</div>
                </td>
                <td className="py-3 px-4 text-gray-300">{row.testTitle}</td>
                <td className="py-3 px-4 text-center">
                  <span
                    className={`px-2 py-1 rounded font-bold text-xs font-mono ${
                      row.violationCount >= 3
                        ? "bg-red-900 text-red-300"
                        : row.violationCount === 2
                        ? "bg-yellow-900 text-yellow-300"
                        : "bg-gray-700 text-gray-300"
                    }`}
                  >
                    {row.violationCount}
                  </span>
                </td>
                <td className="py-3 px-4 text-center text-orange-400 font-mono font-bold">
                  {row.fullscreenExits}
                </td>
                <td className="py-3 px-4 text-center text-yellow-400 font-mono font-bold">
                  {row.tabSwitches}
                </td>
                <td className="py-3 px-4 text-gray-500 text-xs font-mono">
                  {row.lastViolation ? new Date(row.lastViolation).toLocaleString() : "N/A"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

