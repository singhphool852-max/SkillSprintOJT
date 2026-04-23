"use client"

import { useState, useEffect } from "react"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import { useAuth } from "@/hooks/useAuth"
import { Shield, Users, BarChart3, HelpCircle, Zap } from "lucide-react"
import Link from "next/link"
import { StudentsTab } from "@/components/admin/StudentsTab"
import { ArenasTab } from "@/components/admin/ArenasTab"
import { QuestionsTab } from "@/components/admin/QuestionsTab"

const API = "http://localhost:8080/api/admin"

function AdminDashboardContent() {
  const { user, logout } = useAuth()
  const [tab, setTab] = useState<"students" | "arenas" | "questions">("students")
  const [stats, setStats] = useState({ students: 0, arenas: 0, attempts: 0 })

  useEffect(() => {
    fetch(`${API}/stats`, { credentials: "include" })
      .then(r => r.json())
      .then(setStats)
      .catch(() => {})
  }, [])

  const tabs = [
    { key: "students" as const, label: "STUDENTS", icon: Users, color: "neon-cyan" },
    { key: "arenas" as const, label: "ARENAS", icon: BarChart3, color: "neon-pink" },
    { key: "questions" as const, label: "QUESTIONS", icon: HelpCircle, color: "neon-amber" },
  ]

  const statCards = [
    { label: "Students", value: stats.students, color: "text-neon-cyan" },
    { label: "Arenas", value: stats.arenas, color: "text-neon-pink" },
    { label: "Attempts", value: stats.attempts, color: "text-neon-amber" },
  ]

  return (
    <div className="relative min-h-screen bg-deep-bg overflow-hidden">
      <div className="absolute inset-0 grid-bg opacity-20" />
      <div className="absolute inset-0 scanlines" />
      <div className="absolute top-1/4 left-1/3 w-[600px] h-[600px] rounded-full bg-neon-amber/3 blur-[150px]" />

      <div className="relative z-10 max-w-6xl mx-auto px-6 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div className="flex items-center gap-4">
            <Link href="/" className="inline-flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center border border-neon-amber/40 bg-neon-amber/10">
                <Zap className="h-4 w-4 text-neon-amber" />
              </div>
              <span className="font-mono text-base font-bold tracking-wider text-foreground">
                SKILL<span className="text-neon-amber">SPRINT</span>
              </span>
            </Link>
            <div className="flex items-center gap-2 border border-neon-amber/30 bg-neon-amber/5 px-3 py-1">
              <Shield className="h-3 w-3 text-neon-amber" />
              <span className="font-mono text-[9px] tracking-[0.2em] text-neon-amber font-bold">ADMIN</span>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <span className="font-mono text-xs text-muted-foreground">{user?.username}</span>
            <button onClick={logout}
              className="border border-neon-pink/30 bg-neon-pink/5 px-4 py-2 font-mono text-[10px] tracking-[0.2em] text-neon-pink hover:bg-neon-pink/10 transition-colors">
              LOGOUT
            </button>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-8">
          {statCards.map(s => (
            <div key={s.label} className="border border-panel-border bg-deep-bg/80 p-4 text-center">
              <span className={`block font-mono text-2xl font-bold ${s.color}`}>{s.value}</span>
              <span className="font-mono text-[10px] tracking-[0.2em] text-muted-foreground">{s.label.toUpperCase()}</span>
            </div>
          ))}
        </div>

        {/* Tab Navigation */}
        <div className="flex gap-0 mb-6 border-b border-panel-border">
          {tabs.map(t => (
            <button key={t.key} onClick={() => setTab(t.key)}
              className={`flex items-center gap-2 px-6 py-3 font-mono text-xs tracking-[0.15em] transition-colors border-b-2 -mb-px ${
                tab === t.key
                  ? `border-${t.color} text-${t.color} bg-${t.color}/5`
                  : "border-transparent text-muted-foreground hover:text-foreground"
              }`}>
              <t.icon className="h-3.5 w-3.5" />
              {t.label}
            </button>
          ))}
        </div>

        {/* Tab Content */}
        {tab === "students" && <StudentsTab />}
        {tab === "arenas" && <ArenasTab />}
        {tab === "questions" && <QuestionsTab />}
      </div>
    </div>
  )
}

export default function AdminDashboardPage() {
  return (
    <ProtectedRoute requiredRole="admin">
      <AdminDashboardContent />
    </ProtectedRoute>
  )
}
