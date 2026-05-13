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

  const { cleanup: antiCheatCleanup, warningLevel, showWarningModal, handleWarningAcknowledge } = useAntiCheat({
    onViolation: handleViolation,
    onAutoSubmit: handleAutoSubmit,
    maxViolations: 3,
    enabled: isTestActive,
  })

  // Handler for return to fullscreen button (now handled by warning modal)
  const handleReturnToFullscreen = useCallback(async () => {
    handleWarningAcknowledge()
  }, [handleWarningAcknowledge])

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

        {showWarningModal && (
          <div className="fixed inset-0 z-[10001] bg-black/95 backdrop-blur-md flex flex-col items-center justify-center p-6 text-center">
            <div className="max-w-md w-full border border-red-500 bg-panel-bg p-8 shadow-2xl">
              <AlertTriangle className="h-16 w-16 text-red-500 mx-auto mb-6 animate-pulse" />
              
              {warningLevel === 1 && (
                <>
                  <h2 className="text-2xl font-bold text-red-400 mb-2 uppercase tracking-tight">
                    ⚠️ Warning 1 of 3
                  </h2>
                  <p className="text-white mb-6 text-sm">
                    You exited fullscreen mode. Return to fullscreen immediately or your test will be auto-submitted after 2 more violations.
                  </p>
                  <button
                    onClick={handleReturnToFullscreen}
                    className="w-full bg-red-600 hover:bg-red-700 text-white py-4 font-mono text-xs font-bold tracking-widest transition-all"
                  >
                    RETURN TO FULLSCREEN
                  </button>
                </>
              )}

              {warningLevel === 2 && (
                <>
                  <h2 className="text-2xl font-bold text-orange-400 mb-2 uppercase tracking-tight">
                    ⚠️ Warning 2 of 3 — FINAL WARNING
                  </h2>
                  <p className="text-white mb-6 text-sm">
                    You switched tabs or left the test window. This is your FINAL warning. One more violation will immediately auto-submit your test.
                  </p>
                  <button
                    onClick={handleWarningAcknowledge}
                    className="w-full bg-orange-600 hover:bg-orange-700 text-white py-4 font-mono text-xs font-bold tracking-widest transition-all"
                  >
                    I UNDERSTAND — CONTINUE TEST
                  </button>
                </>
              )}
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
