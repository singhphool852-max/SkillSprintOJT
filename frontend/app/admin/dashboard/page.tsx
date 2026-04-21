"use client"

import { ProtectedRoute } from "@/components/ProtectedRoute"
import { useAuth } from "@/hooks/useAuth"
import { Shield, Users, BarChart3, Settings, Zap } from "lucide-react"
import Link from "next/link"

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

function AdminDashboardContent() {
  const { user, logout } = useAuth()

  const stats = [
    { label: "Total Users", value: "—", icon: Users, color: "neon-cyan" },
    { label: "Active Arenas", value: "5", icon: BarChart3, color: "neon-pink" },
    { label: "System Status", value: "Online", icon: Settings, color: "neon-cyan" },
  ]

  return (
    <div className="relative min-h-screen bg-deep-bg overflow-hidden">
      {/* Background effects */}
      <div className="absolute inset-0 grid-bg opacity-20" />
      <div className="absolute inset-0 scanlines" />
      <div className="absolute top-1/4 left-1/3 w-[600px] h-[600px] rounded-full bg-neon-amber/3 blur-[150px]" />

      <div className="relative z-10 max-w-6xl mx-auto px-6 py-12">
        {/* Header */}
        <div className="flex items-center justify-between mb-10">
          <div className="flex items-center gap-4">
            <Link href="/" className="inline-flex items-center gap-2 group">
              <div className="relative flex h-8 w-8 items-center justify-center border border-neon-amber/40 bg-neon-amber/10">
                <Zap className="h-4 w-4 text-neon-amber" />
              </div>
              <span className="font-mono text-base font-bold tracking-wider text-foreground">
                SKILL<span className="text-neon-amber">SPRINT</span>
              </span>
            </Link>
            <div className="flex items-center gap-2 border border-neon-amber/30 bg-neon-amber/5 px-3 py-1">
              <Shield className="h-3 w-3 text-neon-amber" />
              <span className="font-mono text-[9px] tracking-[0.2em] text-neon-amber font-bold">
                ADMIN PANEL
              </span>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <span className="font-mono text-xs text-muted-foreground tracking-wider">
              {user?.username || user?.email}
            </span>
            <button
              onClick={logout}
              className="border border-neon-pink/30 bg-neon-pink/5 px-4 py-2 font-mono text-[10px] tracking-[0.2em] text-neon-pink hover:bg-neon-pink/10 transition-colors"
            >
              LOGOUT
            </button>
          </div>
        </div>

        {/* Welcome */}
        <div className="mb-10">
          <span className="block font-mono text-[10px] tracking-[0.3em] text-neon-amber mb-2">
            COMMAND CENTER
          </span>
          <h1 className="text-3xl font-black uppercase tracking-tight text-foreground">
            Admin Dashboard
          </h1>
          <p className="text-sm text-muted-foreground mt-2 font-mono">
            Welcome back, <span className="text-neon-amber">{user?.username || "Admin"}</span>. System operational.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-10">
          {stats.map((stat) => (
            <div
              key={stat.label}
              className="relative border border-panel-border bg-deep-bg/80 backdrop-blur-md overflow-hidden"
            >
              <HudCorner className="absolute -top-px -left-px text-neon-cyan/50" />
              <HudCorner className="absolute -top-px -right-px text-neon-cyan/50 rotate-90" />
              <HudCorner className="absolute -bottom-px -right-px text-neon-cyan/50 rotate-180" />
              <HudCorner className="absolute -bottom-px -left-px text-neon-cyan/50 -rotate-90" />

              <div className="p-6">
                <div className="flex items-center gap-3 mb-4">
                  <div className={`flex h-10 w-10 items-center justify-center border border-${stat.color}/30 bg-${stat.color}/10`}>
                    <stat.icon className={`h-5 w-5 text-${stat.color}`} />
                  </div>
                  <span className="font-mono text-[11px] tracking-[0.2em] text-muted-foreground uppercase">
                    {stat.label}
                  </span>
                </div>
                <span className={`font-mono text-2xl font-bold text-${stat.color}`}>
                  {stat.value}
                </span>
              </div>
            </div>
          ))}
        </div>

        {/* Placeholder content */}
        <div className="relative border border-panel-border bg-deep-bg/80 backdrop-blur-md overflow-hidden p-8">
          <HudCorner className="absolute -top-px -left-px text-neon-amber/50" />
          <HudCorner className="absolute -top-px -right-px text-neon-amber/50 rotate-90" />
          <HudCorner className="absolute -bottom-px -right-px text-neon-amber/50 rotate-180" />
          <HudCorner className="absolute -bottom-px -left-px text-neon-amber/50 -rotate-90" />

          <div className="text-center">
            <Shield className="h-12 w-12 text-neon-amber/30 mx-auto mb-4" />
            <h2 className="font-mono text-lg font-bold tracking-wider text-foreground mb-2">
              ADMIN CONTROLS
            </h2>
            <p className="font-mono text-xs text-muted-foreground max-w-md mx-auto">
              Manage users, arenas, quizzes, and system settings from this control center.
              Additional admin features coming soon.
            </p>
          </div>
        </div>
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
