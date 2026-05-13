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
  const [showFullscreenWarning, setShowFullscreenWarning] = useState(false)
  const [violationCount, setViolationCount] = useState(0)
  const toastTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const editorRef = useRef<HTMLTextAreaElement | null>(null)
  const submitHandlerRef = useRef<(() => void) | null>(null)

  // ── Anti-cheat wired into arena page ──
  const handleViolation = useCallback((type: string, count: number) => {
    // Update violation count state
    setViolationCount(count)

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

  const handleShowFullscreenWarning = useCallback((show: boolean) => {
    setShowFullscreenWarning(show)
  }, [])

  const { cleanup: antiCheatCleanup } = useAntiCheat({
    onViolation: handleViolation,
    onAutoSubmit: handleAutoSubmit,
    onShowFullscreenWarning: handleShowFullscreenWarning,
    maxViolations: 3,
    enabled: isTestActive,
  })

  // Handler for return to fullscreen button
  const handleReturnToFullscreen = useCallback(async () => {
    try {
      await document.documentElement.requestFullscreen()
      setShowFullscreenWarning(false)
    } catch (e) {
      // Could not enter fullscreen
      // Button click IS a user gesture so this should succeed
      console.error('Fullscreen failed:', e)
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

        {/* Fullscreen warning modal — OVERLAY ONLY, does not replace test UI */}
        {showFullscreenWarning && (
          <div 
            style={{
              position: 'fixed',
              inset: 0,
              zIndex: 99999,
              backgroundColor: 'rgba(0, 0, 0, 0.85)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <div style={{
              maxWidth: '500px',
              padding: '2rem',
              backgroundColor: '#1a1a1a',
              border: '2px solid #ef4444',
              borderRadius: '8px',
              textAlign: 'center',
            }}>
              <AlertTriangle 
                style={{
                  width: '64px',
                  height: '64px',
                  color: '#ef4444',
                  margin: '0 auto 1.5rem',
                }}
              />
              <h2 style={{
                fontSize: '1.5rem',
                fontWeight: 'bold',
                color: '#ef4444',
                marginBottom: '1rem',
                textTransform: 'uppercase',
              }}>
                ⚠ FULLSCREEN REQUIRED
              </h2>
              <p style={{
                color: '#ffffff',
                marginBottom: '0.5rem',
                fontSize: '1.125rem',
                fontWeight: 'bold',
              }}>
                Violation {violationCount}/3
              </p>
              <p style={{
                color: '#d1d5db',
                marginBottom: '0.75rem',
                fontSize: '0.875rem',
              }}>
                Return to fullscreen to continue your test.
              </p>
              <p style={{
                color: '#fbbf24',
                marginBottom: '1.5rem',
                fontSize: '0.875rem',
                fontWeight: 'bold',
              }}>
                Further tab switches will count as additional violations.
              </p>
              <button
                onClick={handleReturnToFullscreen}
                style={{
                  width: '100%',
                  padding: '1rem',
                  backgroundColor: '#dc2626',
                  color: '#ffffff',
                  fontWeight: 'bold',
                  fontSize: '0.875rem',
                  letterSpacing: '0.1em',
                  border: 'none',
                  borderRadius: '4px',
                  cursor: 'pointer',
                  textTransform: 'uppercase',
                }}
                onMouseEnter={(e) => e.currentTarget.style.backgroundColor = '#b91c1c'}
                onMouseLeave={(e) => e.currentTarget.style.backgroundColor = '#dc2626'}
              >
                RETURN TO FULLSCREEN
              </button>
            </div>
          </div>
        )}

        {/* Test UI renders BEHIND the overlay always */}
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
