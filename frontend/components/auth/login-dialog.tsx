"use client"

import { useState, type FormEvent } from "react"
import { useRouter } from "next/navigation"
import { Lock, LogIn, Shield, Swords, User, X } from "lucide-react"
import { API_URL } from "@/lib/api-config"

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

interface LoginDialogProps {
  open: boolean
  onClose: () => void
}

export function LoginDialog({ open, onClose }: LoginDialogProps) {
  const router = useRouter()
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)

  if (!open) return null

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError("")
    setLoading(true)

    try {
      const res = await fetch(`${API_URL}/api/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      })

      if (!res.ok) {
        const data = await res.json()
        setError(data.error || "Invalid credentials")
        setLoading(false)
        return
      }

      // Login successful - store token for WebSocket use
      const data = await res.json()
      if (data.token) {
        localStorage.setItem('auth_token', data.token)
        console.log('[AUTH] Token stored in localStorage for WebSocket')
      }

      // Navigate to arena
      router.push("/arena")
    } catch {
      setError("Connection failed. Try again.")
      setLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-deep-bg/80 backdrop-blur-md"
        onClick={onClose}
      />

      {/* Scanlines overlay */}
      <div className="absolute inset-0 scanlines pointer-events-none" />

      {/* Dialog */}
      <div className="relative z-10 w-full max-w-md mx-4 animate-in fade-in zoom-in-95 duration-300">
        <div className="relative border border-neon-cyan/40 bg-deep-bg/95 backdrop-blur-xl overflow-hidden">
          {/* HUD Corners */}
          <HudCorner className="absolute -top-px -left-px text-neon-cyan/70" />
          <HudCorner className="absolute -top-px -right-px text-neon-cyan/70 rotate-90" />
          <HudCorner className="absolute -bottom-px -right-px text-neon-cyan/70 rotate-180" />
          <HudCorner className="absolute -bottom-px -left-px text-neon-cyan/70 -rotate-90" />

          {/* Glow effects */}
          <div className="absolute -top-20 left-1/2 -translate-x-1/2 w-[300px] h-[200px] rounded-full bg-neon-cyan/5 blur-[80px] pointer-events-none" />
          <div className="absolute -bottom-20 left-1/2 -translate-x-1/2 w-[200px] h-[150px] rounded-full bg-neon-pink/5 blur-[60px] pointer-events-none" />

          {/* Header */}
          <div className="relative border-b border-neon-cyan/20 bg-neon-cyan/5 px-6 py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center border border-neon-cyan/40 bg-neon-cyan/10">
                  <Shield className="h-4 w-4 text-neon-cyan" />
                </div>
                <div>
                  <span className="block font-mono text-xs font-bold tracking-[0.2em] text-neon-cyan">
                    AUTHENTICATION
                  </span>
                  <span className="font-mono text-[9px] tracking-widest text-muted-foreground">
                    IDENTITY VERIFICATION REQUIRED
                  </span>
                </div>
              </div>
              <button
                onClick={onClose}
                className="flex h-7 w-7 items-center justify-center border border-panel-border text-muted-foreground transition-colors hover:border-neon-pink/40 hover:text-neon-pink hover:bg-neon-pink/10"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
          </div>

          {/* Status bar */}
          <div className="flex items-center gap-2 border-b border-panel-border/50 bg-panel-bg/30 px-6 py-2">
            <div className="h-1.5 w-1.5 rounded-full bg-neon-cyan animate-pulse-glow" />
            <span className="font-mono text-[9px] tracking-[0.2em] text-neon-cyan/80">
              SECURE CHANNEL ACTIVE
            </span>
            <div className="h-px flex-1 bg-panel-border" />
            <span className="font-mono text-[9px] tracking-widest text-muted-foreground">
              v3.2.1
            </span>
          </div>

          {/* Form Body */}
          <form onSubmit={handleSubmit} className="p-6 flex flex-col gap-5">
            {/* Username field */}
            <div className="flex flex-col gap-2">
              <label className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
                <User className="h-3 w-3 text-neon-cyan/60" />
                USERNAME
              </label>
              <div className="relative">
                <input
                  id="login-username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Enter your callsign..."
                  required
                  className="w-full border border-panel-border bg-deep-bg/60 px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 transition-all focus:border-neon-cyan/50 focus:outline-none focus:shadow-[0_0_15px_rgba(0,240,255,0.1)] backdrop-blur-sm"
                />
                <div className="absolute right-3 top-1/2 -translate-y-1/2 h-1.5 w-1.5 rounded-full bg-panel-border" />
              </div>
            </div>

            {/* Password field */}
            <div className="flex flex-col gap-2">
              <label className="flex items-center gap-2 font-mono text-[10px] tracking-[0.2em] text-muted-foreground">
                <Lock className="h-3 w-3 text-neon-cyan/60" />
                PASSWORD
              </label>
              <div className="relative">
                <input
                  id="login-password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Enter your access key..."
                  required
                  className="w-full border border-panel-border bg-deep-bg/60 px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 transition-all focus:border-neon-cyan/50 focus:outline-none focus:shadow-[0_0_15px_rgba(0,240,255,0.1)] backdrop-blur-sm"
                />
                <div className="absolute right-3 top-1/2 -translate-y-1/2 h-1.5 w-1.5 rounded-full bg-panel-border" />
              </div>
            </div>

            {/* Error message */}
            {error && (
              <div className="flex items-center gap-2 border border-neon-pink/30 bg-neon-pink/5 px-4 py-2.5">
                <div className="h-1.5 w-1.5 rounded-full bg-neon-pink animate-pulse" />
                <span className="font-mono text-[11px] tracking-wider text-neon-pink">
                  {error.toUpperCase()}
                </span>
              </div>
            )}

            {/* Submit button */}
            <button
              id="login-submit"
              type="submit"
              disabled={loading}
              className="group flex items-center justify-center gap-3 border-2 border-neon-cyan bg-neon-cyan/10 px-6 py-3.5 font-mono text-xs font-bold tracking-widest text-neon-cyan transition-all hover:bg-neon-cyan/20 hover:shadow-[0_0_30px_rgba(0,240,255,0.3)] disabled:opacity-50 disabled:cursor-not-allowed hud-panel backdrop-blur-sm"
            >
              {loading ? (
                <>
                  <div className="h-4 w-4 border-2 border-neon-cyan/30 border-t-neon-cyan rounded-full animate-spin" />
                  VERIFYING...
                </>
              ) : (
                <>
                  <LogIn className="h-4 w-4" />
                  AUTHENTICATE
                  <Swords className="h-4 w-4 transition-transform group-hover:rotate-12" />
                </>
              )}
            </button>
          </form>

          {/* Footer */}
          <div className="border-t border-panel-border/50 bg-panel-bg/20 px-6 py-3">
            <div className="flex items-center justify-center gap-2">
              <div className="h-px flex-1 bg-panel-border/50" />
              <span className="font-mono text-[9px] tracking-[0.2em] text-muted-foreground/50">
                ENCRYPTED CONNECTION
              </span>
              <div className="h-px flex-1 bg-panel-border/50" />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
