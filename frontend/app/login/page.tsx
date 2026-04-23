"use client"

import { useState, useEffect, type FormEvent } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import {
  ChevronRight,
  Crosshair,
  Flame,
  Lock,
  LogIn,
  Mail,
  Shield,
  Swords,
  UserPlus,
  Zap,
} from "lucide-react"
import Link from "next/link"
import { useAuth } from "@/hooks/useAuth"

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

export default function LoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { checkAuth, isAuthenticated, isLoading: authLoading } = useAuth()
  const [mode, setMode] = useState<"login" | "signup">(
    searchParams.get("mode") === "signup" ? "signup" : "login"
  )
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [username, setUsername] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [error, setError] = useState("")
  const [success, setSuccess] = useState("")
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!authLoading && isAuthenticated) {
      router.push("/")
    }
  }, [authLoading, isAuthenticated, router])

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError("")
    setSuccess("")
    setLoading(true)

    if (mode === "signup" && password !== confirmPassword) {
      setError("Passwords do not match")
      setLoading(false)
      return
    }

    const endpoint = mode === "login" ? "http://localhost:8080/api/auth/login" : "http://localhost:8080/api/auth/signup"

    try {
      const bodyPayload = mode === "signup" ? { email, password, username } : { email, password }
      
      const res = await fetch(endpoint, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(bodyPayload),
        credentials: "include",
      })

      const data = await res.json()

      if (!res.ok) {
        setError(data.error || "Something went wrong")
        setLoading(false)
        return
      }

      if (mode === "signup") {
        setSuccess("Account created! Initializing session...")
        // Automatically login after signup
        const loginRes = await fetch("http://localhost:8080/api/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
          credentials: "include",
        })
        
        if (loginRes.ok) {
          await checkAuth()
          router.push("/")
          return
        } else {
          setMode("login")
          setSuccess("Account created. Please login.")
          setLoading(false)
          return
        }
      }

      // Login successful — sync state and redirect to hero
      await checkAuth()
      router.push("/")
    } catch (err) {
      console.error(err)
      setError("Connection failed. Ensure backend is running at :8080")
      setLoading(false)
    }
  }

  return (
    <div className="relative min-h-screen bg-deep-bg overflow-hidden flex items-center justify-center">
      {/* Background effects */}
      <div className="absolute inset-0 grid-bg opacity-30" />
      <div className="absolute inset-0 scanlines" />
      <div className="absolute top-1/4 left-1/3 w-[600px] h-[600px] rounded-full bg-neon-cyan/3 blur-[150px]" />
      <div className="absolute bottom-1/4 right-1/3 w-[500px] h-[500px] rounded-full bg-neon-pink/3 blur-[120px]" />

      {/* Content */}
      <div className="relative z-10 w-full max-w-5xl mx-auto px-4 py-8 flex flex-col items-center gap-8">
        {/* Header */}
        <div className="text-center">
          <Link href="/" className="inline-flex items-center gap-2 mb-6 group">
            <div className="relative flex h-8 w-8 items-center justify-center border border-neon-cyan/40 bg-neon-cyan/10">
              <Zap className="h-4 w-4 text-neon-cyan" />
            </div>
            <span className="font-mono text-base font-bold tracking-wider text-foreground">
              SKILL<span className="text-neon-cyan text-glow-cyan">SPRINT</span>
            </span>
          </Link>

          <div className="inline-block border border-neon-cyan/30 bg-neon-cyan/5 px-4 py-1 mb-6">
            <span className="font-mono text-[10px] tracking-[0.3em] text-neon-cyan">
              COMPETITIVE LOGIN
            </span>
          </div>

          <h1 className="text-4xl sm:text-5xl font-black uppercase tracking-tight text-foreground mb-2">
            ENTER YOUR
          </h1>
          <h1 className="text-4xl sm:text-5xl font-black uppercase tracking-tight text-neon-cyan text-glow-cyan mb-4">
            SKILL ZONE
          </h1>
          <p className="text-sm text-muted-foreground max-w-md mx-auto">
            Login to continue your battles, track your progress, and enter the arena.
          </p>
        </div>

        {/* Main 3-column layout */}
        <div className="w-full flex flex-col lg:flex-row items-center lg:items-start justify-center gap-6">
          {/* LEFT - System Online Panel */}
          <div className="w-full max-w-xs lg:w-60 hidden lg:block">
            <div className="relative border border-neon-cyan/30 bg-deep-bg/80 backdrop-blur-md overflow-hidden">
              <HudCorner className="absolute -top-px -left-px text-neon-cyan/70" />
              <HudCorner className="absolute -top-px -right-px text-neon-cyan/70 rotate-90" />
              <HudCorner className="absolute -bottom-px -right-px text-neon-cyan/70 rotate-180" />
              <HudCorner className="absolute -bottom-px -left-px text-neon-cyan/70 -rotate-90" />

              <div className="border-b border-neon-cyan/20 bg-neon-cyan/5 px-5 py-3">
                <div className="flex items-center gap-2">
                  <Shield className="h-3.5 w-3.5 text-neon-cyan" />
                  <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-cyan">
                    SYSTEM ONLINE
                  </span>
                </div>
              </div>

              <div className="flex flex-col gap-0.5 p-4">
                {[
                  { label: "Mode", value: "Ranked", color: "neon-cyan" },
                  { label: "Access", value: "Ready", color: "neon-cyan" },
                  { label: "Players", value: "2,847", color: "neon-cyan" },
                ].map((stat) => (
                  <div
                    key={stat.label}
                    className="flex items-center justify-between border-b border-panel-border/50 py-3 last:border-0"
                  >
                    <span className="font-mono text-[11px] tracking-wider text-muted-foreground">
                      {stat.label}
                    </span>
                    <span className={`font-mono text-sm font-bold text-${stat.color}`}>
                      {stat.value}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* CENTER - Auth form */}
          <div className="w-full max-w-md">
            <div className="relative border border-panel-border bg-deep-bg/90 backdrop-blur-xl overflow-hidden">
              <HudCorner className="absolute -top-px -left-px text-neon-cyan/50" />
              <HudCorner className="absolute -top-px -right-px text-neon-cyan/50 rotate-90" />
              <HudCorner className="absolute -bottom-px -right-px text-neon-cyan/50 rotate-180" />
              <HudCorner className="absolute -bottom-px -left-px text-neon-cyan/50 -rotate-90" />

              {/* Form header */}
              <div className="flex items-center justify-between border-b border-panel-border px-6 py-4">
                <div>
                  <span className="block font-mono text-[9px] tracking-[0.2em] text-muted-foreground">
                    PLAYER AUTHENTICATION
                  </span>
                  <span className="block font-mono text-lg font-bold tracking-wider text-foreground">
                    {mode === "login" ? "LOGIN" : "SIGN UP"}
                  </span>
                </div>
                <div className="flex items-center gap-2 border border-neon-cyan/30 bg-neon-cyan/5 px-3 py-1">
                  <div className="h-1.5 w-1.5 rounded-full bg-neon-cyan animate-pulse-glow" />
                  <span className="font-mono text-[9px] tracking-[0.2em] text-neon-cyan">
                    SECURE
                  </span>
                </div>
              </div>

              {/* Form */}
              <form onSubmit={handleSubmit} className="p-6 flex flex-col gap-5">
                {/* Email */}
                <div className="flex flex-col gap-2">
                  <label className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
                    <Mail className="h-3 w-3 text-neon-cyan/60" />
                    EMAIL
                  </label>
                  <input
                    id="auth-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="player@skillsprint.io"
                    required
                    className="w-full border border-panel-border bg-deep-bg/60 px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 transition-all focus:border-neon-cyan/50 focus:outline-none focus:shadow-[0_0_15px_rgba(0,240,255,0.1)]"
                  />
                </div>

                {/* Username */}
                {mode === "signup" && (
                  <div className="flex flex-col gap-2">
                    <label className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
                      <Crosshair className="h-3 w-3 text-neon-cyan/60" />
                      USERNAME
                    </label>
                    <input
                      id="auth-username"
                      type="text"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      placeholder="rival_x"
                      required
                      className="w-full border border-panel-border bg-deep-bg/60 px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 transition-all focus:border-neon-cyan/50 focus:outline-none focus:shadow-[0_0_15px_rgba(0,240,255,0.1)]"
                    />
                  </div>
                )}

                {/* Password */}
                <div className="flex flex-col gap-2">
                  <label className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
                    <Lock className="h-3 w-3 text-neon-cyan/60" />
                    PASSWORD
                  </label>
                  <input
                    id="auth-password"
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••"
                    required
                    minLength={6}
                    className="w-full border border-panel-border bg-deep-bg/60 px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 transition-all focus:border-neon-cyan/50 focus:outline-none focus:shadow-[0_0_15px_rgba(0,240,255,0.1)]"
                  />
                </div>

                {/* Confirm Password */}
                {mode === "signup" && (
                  <div className="flex flex-col gap-2">
                    <label className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
                      <Lock className="h-3 w-3 text-neon-cyan/60" />
                      CONFIRM PASSWORD
                    </label>
                    <input
                      id="auth-confirm-password"
                      type="password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder="••••••"
                      required
                      minLength={6}
                      className="w-full border border-panel-border bg-deep-bg/60 px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 transition-all focus:border-neon-cyan/50 focus:outline-none focus:shadow-[0_0_15px_rgba(0,240,255,0.1)]"
                    />
                  </div>
                )}

                {/* Error */}
                {error && (
                  <div className="flex items-center gap-2 border border-neon-pink/30 bg-neon-pink/5 px-4 py-2.5">
                    <div className="h-1.5 w-1.5 rounded-full bg-neon-pink animate-pulse" />
                    <span className="font-mono text-[11px] tracking-wider text-neon-pink">
                      {error.toUpperCase()}
                    </span>
                  </div>
                )}

                {/* Success */}
                {success && (
                  <div className="flex items-center gap-2 border border-neon-cyan/30 bg-neon-cyan/5 px-4 py-2.5">
                    <div className="h-1.5 w-1.5 rounded-full bg-neon-cyan animate-pulse" />
                    <span className="font-mono text-[11px] tracking-wider text-neon-cyan">
                      {success.toUpperCase()}
                    </span>
                  </div>
                )}

                {/* Submit */}
                <button
                  id="auth-submit"
                  type="submit"
                  disabled={loading}
                  className="group flex items-center justify-center gap-3 bg-neon-cyan/90 hover:bg-neon-cyan px-6 py-3.5 font-mono text-xs font-bold tracking-widest text-deep-bg transition-all hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {loading ? (
                    <>
                      <div className="h-4 w-4 border-2 border-deep-bg/30 border-t-deep-bg rounded-full animate-spin" />
                      {mode === "login" ? "VERIFYING..." : "CREATING ACCOUNT..."}
                    </>
                  ) : mode === "login" ? (
                    <>
                      ENTER ARENA
                      <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                    </>
                  ) : (
                    <>
                      <UserPlus className="h-4 w-4" />
                      CREATE ACCOUNT
                    </>
                  )}
                </button>

                {/* Toggle mode */}
                <div className="text-center">
                  {mode === "login" ? (
                    <span className="font-mono text-[11px] text-muted-foreground">
                      New challenger?{" "}
                      <button
                        type="button"
                        onClick={() => {
                          setMode("signup")
                          setError("")
                          setSuccess("")
                          setConfirmPassword("")
                        }}
                        className="text-neon-cyan hover:text-neon-cyan/80 underline underline-offset-4 transition-colors"
                      >
                        Create account
                      </button>
                    </span>
                  ) : (
                    <span className="font-mono text-[11px] text-muted-foreground">
                      Already have an account?{" "}
                      <button
                        type="button"
                        onClick={() => {
                          setMode("login")
                          setError("")
                          setSuccess("")
                          setConfirmPassword("")
                        }}
                        className="text-neon-cyan hover:text-neon-cyan/80 underline underline-offset-4 transition-colors"
                      >
                        Login
                      </button>
                    </span>
                  )}
                </div>
              </form>
            </div>
          </div>

          {/* RIGHT - Opponent Feed Panel */}
          <div className="w-full max-w-xs lg:w-60 hidden lg:block">
            <div className="relative border border-neon-pink/30 bg-deep-bg/80 backdrop-blur-md overflow-hidden">
              <HudCorner className="absolute -top-px -left-px text-neon-pink/70" />
              <HudCorner className="absolute -top-px -right-px text-neon-pink/70 rotate-90" />
              <HudCorner className="absolute -bottom-px -right-px text-neon-pink/70 rotate-180" />
              <HudCorner className="absolute -bottom-px -left-px text-neon-pink/70 -rotate-90" />

              <div className="border-b border-neon-pink/20 bg-neon-pink/5 px-5 py-3">
                <div className="flex items-center gap-2">
                  <Swords className="h-3.5 w-3.5 text-neon-pink" />
                  <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-pink">
                    OPPONENT FEED
                  </span>
                </div>
              </div>

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
                      Searching...
                    </span>
                  </div>
                </div>

                <div className="flex flex-col gap-3">
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] tracking-wider text-muted-foreground">
                      Rank
                    </span>
                    <span className="font-mono text-sm font-bold text-neon-amber text-glow-amber">
                      PLATINUM I
                    </span>
                  </div>
                  <div className="h-px bg-panel-border/50" />
                  <div className="flex items-center justify-between">
                    <span className="font-mono text-[11px] tracking-wider text-muted-foreground">
                      Win Streak
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
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
