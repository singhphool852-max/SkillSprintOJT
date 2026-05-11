"use client"

import { useState, useEffect, useRef, useCallback } from "react"
import { ArenaLobby } from "@/components/arena/arena-lobby"
import { TestArena } from "@/components/arena/test-arena"
import { useAntiCheat } from "@/hooks/useAntiCheat"
import { API_URL } from "@/lib/api-config"

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

  const { cleanup: antiCheatCleanup } = useAntiCheat({
    onViolation: handleViolation,
    onAutoSubmit: handleAutoSubmit,
    maxViolations: 3,
    enabled: isTestActive,
  })

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
