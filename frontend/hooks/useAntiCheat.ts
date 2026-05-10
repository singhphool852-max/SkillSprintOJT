"use client"

import { useEffect, useRef, useCallback, useState } from "react"
import { API_URL } from "@/lib/api-config"

// ──────────────────────────────────────────────
// Anti-Cheat Hook — fullscreen lock, visibility,
// blur, keyboard shortcut, copy/paste blocking.
//
// KEY FIX: Anti-cheat is "armed" only AFTER fullscreen
// is successfully established + a 1s grace period.
// This prevents false violations on initial join.
// ──────────────────────────────────────────────

interface AntiCheatOptions {
  attemptId: string
  testId: string
  remainingSeconds: number
  onAutoSubmit: () => void
  enabled: boolean // only active when test is in progress
}

interface AntiCheatState {
  violationCount: number
  warningMessage: string | null
  showWarning: boolean
  requestFullscreen: () => void // caller triggers on user gesture (Join click)
}

const VIOLATION_LABELS: Record<string, string> = {
  fullscreen_exit: "FULLSCREEN EXIT DETECTED",
  tab_switch: "TAB SWITCH DETECTED",
  window_blur: "WINDOW FOCUS LOST",
  copy_blocked: "COPY ATTEMPT BLOCKED",
  paste_blocked: "PASTE ATTEMPT BLOCKED",
  cut_blocked: "CUT ATTEMPT BLOCKED",
  contextmenu_blocked: "RIGHT-CLICK BLOCKED",
  keyboard_shortcut: "BLOCKED SHORTCUT DETECTED",
}

export function useAntiCheat({
  attemptId,
  testId,
  remainingSeconds,
  onAutoSubmit,
  enabled,
}: AntiCheatOptions): AntiCheatState {
  const [violationCount, setViolationCount] = useState(0)
  const [warningMessage, setWarningMessage] = useState<string | null>(null)
  const [showWarning, setShowWarning] = useState(false)
  const remainingRef = useRef(remainingSeconds)
  const violationCountRef = useRef(0)
  const autoSubmitCalledRef = useRef(false)

  // ── ARMED FLAG: prevents false triggers during setup ──
  const antiCheatArmedRef = useRef(false)

  // Debounce: prevent duplicate violations from rapid event bursts
  const lastViolationTimeRef = useRef(0)

  // Keep remaining time ref in sync
  useEffect(() => {
    remainingRef.current = remainingSeconds
  }, [remainingSeconds])

  // Log violation to backend
  const logViolation = useCallback(
    async (violationType: string) => {
      // GUARD: only count violations when anti-cheat is armed
      if (!antiCheatArmedRef.current) return

      // Debounce: ignore violations within 2s of the last one
      const now = Date.now()
      if (now - lastViolationTimeRef.current < 2000) return
      lastViolationTimeRef.current = now

      const newCount = violationCountRef.current + 1
      violationCountRef.current = newCount
      setViolationCount(newCount)

      // Show warning
      const label = VIOLATION_LABELS[violationType] || violationType.toUpperCase()
      if (newCount === 1) {
        setWarningMessage(`⚠️ WARNING 1/3: ${label}. This is your first warning. Further violations will result in auto-submission.`)
      } else if (newCount === 2) {
        setWarningMessage(`🚨 FINAL WARNING 2/3: ${label}. One more violation will AUTO-SUBMIT your test!`)
      } else {
        setWarningMessage(`❌ VIOLATION 3/3: ${label}. Your test is being auto-submitted.`)
      }
      setShowWarning(true)

      // Auto-dismiss warning after 4s (except on auto-submit)
      if (newCount < 3) {
        setTimeout(() => setShowWarning(false), 4000)
      }

      // Log to backend
      try {
        await fetch(`${API_URL}/api/arena/violations`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            attemptId,
            testId,
            violationType,
            remainingTime: remainingRef.current,
          }),
        })
      } catch (e) {
        // Silent fail — don't break the test flow
      }

      // Auto-submit on 3rd violation
      if (newCount >= 3 && !autoSubmitCalledRef.current) {
        autoSubmitCalledRef.current = true
        setTimeout(() => {
          onAutoSubmit()
        }, 1500) // Brief delay so user sees the message
      }
    },
    [attemptId, testId, onAutoSubmit]
  )

  // Request fullscreen — exported for parent to call from user gesture (Join button)
  const requestFullscreen = useCallback(() => {
    const el = document.documentElement
    if (el.requestFullscreen) {
      el.requestFullscreen()
        .then(() => {
          // Fullscreen entered successfully — arm after grace period
          setTimeout(() => {
            antiCheatArmedRef.current = true
          }, 1500) // 1.5s grace to avoid race with fullscreenchange event
        })
        .catch(() => {
          // Browser blocked (no user gesture). Still arm anti-cheat
          // after a delay so monitoring works even without fullscreen.
          setTimeout(() => {
            antiCheatArmedRef.current = true
          }, 2000)
        })
    } else {
      // Fullscreen API not supported — arm anyway
      setTimeout(() => {
        antiCheatArmedRef.current = true
      }, 1500)
    }
  }, [])

  useEffect(() => {
    if (!enabled) {
      // Reset armed state when disabled
      antiCheatArmedRef.current = false
      return
    }

    // NOTE: We do NOT call requestFullscreen() here.
    // The parent component calls requestFullscreen() from the Join button handler
    // so it's triggered by a user gesture (required by browsers).

    // ── Fullscreen exit detection ──
    function handleFullscreenChange() {
      if (!document.fullscreenElement && antiCheatArmedRef.current && violationCountRef.current < 3) {
        logViolation("fullscreen_exit")
        // Try to re-enter fullscreen
        setTimeout(() => {
          document.documentElement.requestFullscreen?.().catch(() => {})
        }, 500)
      }
    }

    // ── Tab switch detection ──
    function handleVisibilityChange() {
      if (document.hidden && antiCheatArmedRef.current && violationCountRef.current < 3) {
        logViolation("tab_switch")
      }
    }

    // ── Window blur detection ──
    function handleWindowBlur() {
      if (antiCheatArmedRef.current && violationCountRef.current < 3) {
        logViolation("window_blur")
      }
    }

    // ── Copy/paste/cut/contextmenu blocking ──
    function handleCopy(e: ClipboardEvent) {
      // Allow inside code editor textarea
      if ((e.target as HTMLElement)?.tagName === "TEXTAREA") return
      e.preventDefault()
      logViolation("copy_blocked")
    }

    function handlePaste(e: ClipboardEvent) {
      // Allow inside code editor textarea
      if ((e.target as HTMLElement)?.tagName === "TEXTAREA") return
      e.preventDefault()
      logViolation("paste_blocked")
    }

    function handleCut(e: ClipboardEvent) {
      if ((e.target as HTMLElement)?.tagName === "TEXTAREA") return
      e.preventDefault()
      logViolation("cut_blocked")
    }

    function handleContextMenu(e: MouseEvent) {
      e.preventDefault()
    }

    // ── Keyboard shortcut blocking ──
    function handleKeyDown(e: KeyboardEvent) {
      const ctrl = e.ctrlKey || e.metaKey

      // F12
      if (e.key === "F12") {
        e.preventDefault()
        logViolation("keyboard_shortcut")
        return
      }

      if (ctrl) {
        // Ctrl+Shift+I, Ctrl+Shift+J
        if (e.shiftKey && (e.key === "I" || e.key === "i" || e.key === "J" || e.key === "j")) {
          e.preventDefault()
          logViolation("keyboard_shortcut")
          return
        }
        // Ctrl+U, Ctrl+L, Ctrl+T, Ctrl+N, Ctrl+W
        if (["u", "U", "l", "L", "t", "T", "n", "N", "w", "W"].includes(e.key)) {
          e.preventDefault()
          logViolation("keyboard_shortcut")
          return
        }
        // Ctrl+Tab
        if (e.key === "Tab") {
          e.preventDefault()
          logViolation("keyboard_shortcut")
          return
        }
      }
    }

    // ── Register all listeners ──
    document.addEventListener("fullscreenchange", handleFullscreenChange)
    document.addEventListener("visibilitychange", handleVisibilityChange)
    window.addEventListener("blur", handleWindowBlur)
    document.addEventListener("copy", handleCopy as EventListener)
    document.addEventListener("paste", handlePaste as EventListener)
    document.addEventListener("cut", handleCut as EventListener)
    document.addEventListener("contextmenu", handleContextMenu as EventListener)
    document.addEventListener("keydown", handleKeyDown)

    // ── Cleanup ──
    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange)
      document.removeEventListener("visibilitychange", handleVisibilityChange)
      window.removeEventListener("blur", handleWindowBlur)
      document.removeEventListener("copy", handleCopy as EventListener)
      document.removeEventListener("paste", handlePaste as EventListener)
      document.removeEventListener("cut", handleCut as EventListener)
      document.removeEventListener("contextmenu", handleContextMenu as EventListener)
      document.removeEventListener("keydown", handleKeyDown)

      // Exit fullscreen on cleanup
      if (document.fullscreenElement) {
        document.exitFullscreen().catch(() => {})
      }
      antiCheatArmedRef.current = false
    }
  }, [enabled, logViolation])

  return { violationCount, warningMessage, showWarning, requestFullscreen }
}
