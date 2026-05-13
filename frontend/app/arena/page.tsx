"use client"

import { useState, useEffect, useRef, useCallback } from "react"
import { ArenaLobby } from "@/components/arena/arena-lobby"
import { TestArena } from "@/components/arena/test-arena"
import { useAntiCheat } from "@/hooks/useAntiCheat"
import { API_URL } from "@/lib/api-config"
import { AlertTriangle } from "lucide-react"

const AUTOSAVE_KEY = "arena_autosave_code"
const AUTOSAVE_LANG_KEY = "arena_autosave_lang"

export default function ArenaPage() {
  const [isTestActive, setIsTestActive] = useState(false)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [violationToast, setViolationToast] = useState<string | null>(null)
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const editorRef = useRef<HTMLTextAreaElement | null>(null)
  const submitHandlerRef = useRef<(() => void) | null>(null)

  // ── Anti-cheat wired into arena page ──
  const handleViolation = useCallback((type: string, count: number) => {
    // Show toast
    const label = type.replace(/_/g, " ").toUpperCase()
    setViolationToast(`⚠️ Violation ${count}/3: ${label}`)
    if (toastTimerRef.current) clearTimeout(toastTimerRef.current)
    toastTimerRef.current = setTimeout(() => setViolationToast(null), 4000)

    // Fire-and-forget POST
    if (!sessionId) return
    fetch(`${API_URL}/api/arena/violations`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ type, count, sessionId }),
    }).catch(() => {})
  }, [sessionId])

  const handleAutoSubmit = useCallback(() => {
    submitHandlerRef.current?.()
  }, [])

  const { cleanup: antiCheatCleanup, showFullscreenWarning } = useAntiCheat({
    onViolation: handleViolation,
    onAutoSubmit: handleAutoSubmit,
    maxViolations: 3,
    enabled: isTestActive,
  })

  // Handler for return to fullscreen button
  const handleReturnToFullscreen = useCallback(async () => {
    try {
      await document.documentElement.requestFullscreen()
      // The fullscreenchange event in useAntiCheat will clear showFullscreenWarning
      // But add a fallback check after a short delay
      setTimeout(() => {
        if (document.fullscreenElement) {
          // Force re-render if needed by checking the hook state
          console.log('[Arena] Fullscreen re-entered successfully')
        }
      }, 500)
    } catch (err) {
      console.error('[Arena] Failed to re-enter fullscreen:', err)
    }
  }, [])

  // ── beforeunload: warn during active test ──
  useEffect(() => {
    if (!isTestActive) return
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault()
      // Modern browsers show a generic message regardless of returnValue
    }
    window.addEventListener("beforeunload", handler)
    return () => window.removeEventListener("beforeunload", handler)
  }, [isTestActive])

  // ── Auto-save code to localStorage every 30s ──
  useEffect(() => {
    if (!isTestActive) return
    const interval = setInterval(() => {
      if (editorRef.current) {
        localStorage.setItem(AUTOSAVE_KEY, editorRef.current.value)
      }
    }, 30000)
    return () => clearInterval(interval)
  }, [isTestActive])

  // ── Disable text selection outside editor ──
  useEffect(() => {
    if (!isTestActive) return
    document.body.style.userSelect = "none"
    document.body.style.webkitUserSelect = "none"
    return () => {
      document.body.style.userSelect = ""
      document.body.style.webkitUserSelect = ""
    }
  }, [isTestActive])

  // ── Re-focus editor on window focus return ──
  useEffect(() => {
    if (!isTestActive) return
    const handler = () => {
      editorRef.current?.focus()
    }
    window.addEventListener("focus", handler)
    return () => window.removeEventListener("focus", handler)
  }, [isTestActive])

  // ── Handle test activation/deactivation ──
  const handleActiveChange = useCallback((active: boolean, sId?: string) => {
    setIsTestActive(active)
    if (active && sId) {
      setSessionId(sId)
    }
    if (!active) {
      // Cleanup on deactivation
      antiCheatCleanup()
      setSessionId(null)
      localStorage.removeItem(AUTOSAVE_KEY)
      localStorage.removeItem(AUTOSAVE_LANG_KEY)
    }
  }, [antiCheatCleanup])

  // When a test is active: render in fixed fullscreen overlay
  if (isTestActive) {
    return (
      <main className="fixed inset-0 z-[9999] bg-deep-bg overflow-hidden flex flex-col">
        {/* Violation toast — always visible, top-right */}
        {violationToast && (
          <div className="fixed top-4 right-4 z-[10000] bg-red-900/90 border border-red-500 text-red-100 px-5 py-3 rounded-lg shadow-2xl animate-pulse text-sm font-bold">
            {violationToast}
          </div>
        )}

        {showFullscreenWarning && (
          <div className="fixed inset-0 z-[10001] bg-black/95 backdrop-blur-md flex flex-col items-center justify-center p-6 text-center">
            <div className="max-w-md w-full border border-panel-border bg-panel-bg p-8 shadow-2xl">
              <AlertTriangle className="h-16 w-16 text-neon-pink mx-auto mb-6 animate-pulse-glow" />
              <h2 className="text-2xl font-bold text-white mb-2 uppercase tracking-tighter">
                Fullscreen <span className="text-neon-pink text-glow-pink">Required</span>
              </h2>
              <p className="text-sm text-muted-foreground mb-8">
                Exiting fullscreen is a violation of the test policy. 
                You must return to fullscreen mode to continue your session.
              </p>
              <button
                onClick={handleReturnToFullscreen}
                className="w-full bg-neon-cyan/90 hover:bg-neon-cyan text-deep-bg py-4 font-mono text-xs font-bold tracking-[0.2em] transition-all"
              >
                RETURN TO FULLSCREEN
              </button>
            </div>
          </div>
        )}

        <TestArena
          onActiveChange={handleActiveChange}
          editorRef={editorRef}
          submitHandlerRef={submitHandlerRef}
        />
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-deep-bg">
      <div className="pt-20">
        <ArenaLobby />
      </div>

      {/* Coding Tests Section */}
      <div className="relative mx-auto max-w-7xl px-4 lg:px-8 py-4">
        <div className="h-px bg-panel-border" />
      </div>
      <TestArena
        onActiveChange={handleActiveChange}
        editorRef={editorRef}
        submitHandlerRef={submitHandlerRef}
      />
    </main>
  )
}
