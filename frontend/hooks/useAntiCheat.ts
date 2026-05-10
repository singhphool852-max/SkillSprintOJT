"use client"

import { useEffect, useRef, useCallback, useState } from "react"
import { API_URL } from "@/lib/api-config"

// ──────────────────────────────────────────────
// Anti-Cheat Hook — fullscreen lock, visibility,
// blur, keyboard shortcut, copy/paste/cut/contextmenu blocking.
//
// KEY DESIGN:
// 1. requestFullscreen() is exported so the PARENT can call it
//    directly inside a click handler (user gesture required by browsers).
// 2. Anti-cheat is "armed" only AFTER fullscreen succeeds + 1.5s grace.
// 3. Copy/paste/cut/contextmenu are always blocked when armed.
// 4. DevTools shortcuts (F12, Ctrl+Shift+I/J, Ctrl+U etc.) are blocked.
// 5. 3 violations → auto-submit.
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
  requestFullscreen: () => Promise<boolean> // caller triggers on user gesture (Join click)
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
  // Returns a promise that resolves to true if fullscreen was entered.
  const requestFullscreen = useCallback(async (): Promise<boolean> => {
    const el = document.documentElement
    try {
      if (el.requestFullscreen) {
        await el.requestFullscreen()
        // Fullscreen entered successfully — arm after grace period
        setTimeout(() => {
          antiCheatArmedRef.current = true
        }, 1500) // 1.5s grace to avoid race with fullscreenchange event
        return true
      }
    } catch {
      // Browser blocked (no user gesture or policy). Still arm anti-cheat
      // after a delay so monitoring works even without fullscreen.
      console.warn("[anti-cheat] Fullscreen request failed, arming anyway")
    }
    // Fallback: arm anyway after delay
    setTimeout(() => {
      antiCheatArmedRef.current = true
    }, 2000)
    return false
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

    // ── Copy/paste/cut blocking — ALWAYS block (even in textarea) ──
    function handleCopy(e: ClipboardEvent) {
      if (!antiCheatArmedRef.current) return
      e.preventDefault()
      logViolation("copy_blocked")
    }

    function handlePaste(e: ClipboardEvent) {
      if (!antiCheatArmedRef.current) return
      e.preventDefault()
      logViolation("paste_blocked")
    }

    function handleCut(e: ClipboardEvent) {
      if (!antiCheatArmedRef.current) return
      e.preventDefault()
      logViolation("cut_blocked")
    }

    // ── Right-click blocking ──
    function handleContextMenu(e: MouseEvent) {
      if (!antiCheatArmedRef.current) return
      e.preventDefault()
      logViolation("contextmenu_blocked")
    }

    // ── Keyboard shortcut blocking ──
    function handleKeyDown(e: KeyboardEvent) {
      if (!antiCheatArmedRef.current) return

      const ctrl = e.ctrlKey || e.metaKey

      // F12
      if (e.key === "F12") {
        e.preventDefault()
        logViolation("keyboard_shortcut")
        return
      }

      if (ctrl) {
        // Ctrl+Shift+I, Ctrl+Shift+J (DevTools)
        if (e.shiftKey && (e.key === "I" || e.key === "i" || e.key === "J" || e.key === "j")) {
          e.preventDefault()
          logViolation("keyboard_shortcut")
          return
        }
        // Ctrl+U (view source), Ctrl+L (address bar), Ctrl+T (new tab),
        // Ctrl+N (new window), Ctrl+W (close tab)
        if (["u", "U", "l", "L", "t", "T", "n", "N", "w", "W"].includes(e.key)) {
          e.preventDefault()
          logViolation("keyboard_shortcut")
          return
        }
        // Ctrl+Tab (switch tab)
        if (e.key === "Tab") {
          e.preventDefault()
          logViolation("keyboard_shortcut")
          return
        }
      }

      // Alt+Tab detection (limited — browser may not fire this)
      if (e.altKey && e.key === "Tab") {
        e.preventDefault()
        logViolation("keyboard_shortcut")
        return
      }
    }

    // ── Register all listeners ──
    document.addEventListener("fullscreenchange", handleFullscreenChange)
    document.addEventListener("visibilitychange", handleVisibilityChange)
    window.addEventListener("blur", handleWindowBlur)
    document.addEventListener("copy", handleCopy as EventListener, true)  // capture phase
    document.addEventListener("paste", handlePaste as EventListener, true)
    document.addEventListener("cut", handleCut as EventListener, true)
    document.addEventListener("contextmenu", handleContextMenu as EventListener, true)
    document.addEventListener("keydown", handleKeyDown, true) // capture phase to intercept before editor

    // ── Cleanup ──
    return () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange)
      document.removeEventListener("visibilitychange", handleVisibilityChange)
      window.removeEventListener("blur", handleWindowBlur)
      document.removeEventListener("copy", handleCopy as EventListener, true)
      document.removeEventListener("paste", handlePaste as EventListener, true)
      document.removeEventListener("cut", handleCut as EventListener, true)
      document.removeEventListener("contextmenu", handleContextMenu as EventListener, true)
      document.removeEventListener("keydown", handleKeyDown, true)

      // Exit fullscreen on cleanup
      if (document.fullscreenElement) {
        document.exitFullscreen().catch(() => {})
      }
      antiCheatArmedRef.current = false
    }
  }, [enabled, logViolation])

  return { violationCount, warningMessage, showWarning, requestFullscreen }
}
