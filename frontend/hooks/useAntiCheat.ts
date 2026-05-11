"use client"

import { useEffect, useRef, useCallback } from "react"

// ──────────────────────────────────────────────
// Anti-Cheat Hook
//
// Forces fullscreen, blocks shortcuts/clipboard/right-click,
// detects tab switches, blur, and fullscreen exits.
// Uses useRef for violation count to avoid unnecessary re-renders.
// Calls onAutoSubmit and removes ALL listeners after maxViolations.
// ──────────────────────────────────────────────

type ViolationType =
  | "fullscreen_exit"
  | "tab_switch"
  | "window_blur"
  | "copy_blocked"
  | "paste_blocked"
  | "cut_blocked"
  | "contextmenu_blocked"
  | "keyboard_shortcut"

interface UseAntiCheatProps {
  onViolation: (type: string, count: number) => void
  onAutoSubmit: () => void
  maxViolations?: number
  enabled?: boolean
}

interface UseAntiCheatReturn {
  violationCount: number
  cleanup: () => void
}

export function useAntiCheat({
  onViolation,
  onAutoSubmit,
  maxViolations = 3,
  enabled = true,
}: UseAntiCheatProps): UseAntiCheatReturn {
  const violationCountRef = useRef(0)
  const armedRef = useRef(false)
  const lastViolationTimeRef = useRef(0)
  const cleanupRef = useRef<(() => void) | null>(null)
  // Stable callback refs to avoid stale closures
  const onViolationRef = useRef(onViolation)
  const onAutoSubmitRef = useRef(onAutoSubmit)

  onViolationRef.current = onViolation
  onAutoSubmitRef.current = onAutoSubmit

  const recordViolation = useCallback(
    (type: ViolationType) => {
      if (!armedRef.current) return
      // Debounce: ignore violations within 1.5s of each other
      const now = Date.now()
      if (now - lastViolationTimeRef.current < 1500) return
      lastViolationTimeRef.current = now

      violationCountRef.current += 1
      const count = violationCountRef.current
      onViolationRef.current(type, count)

      if (count >= maxViolations) {
        // Auto-submit after brief delay so user sees the final warning
        setTimeout(() => {
          onAutoSubmitRef.current()
          cleanupRef.current?.()
        }, 1200)
      }
    },
    [maxViolations]
  )

  const cleanup = useCallback(() => {
    armedRef.current = false
    cleanupRef.current?.()
    if (document.fullscreenElement) {
      document.exitFullscreen().catch(() => {})
    }
  }, [])

  useEffect(() => {
    if (!enabled) {
      armedRef.current = false
      return
    }

    // Request fullscreen
    const el = document.documentElement
    el.requestFullscreen?.()
      .then(() => {
        // Arm after grace period to avoid false trigger from the fullscreenchange event itself
        setTimeout(() => {
          armedRef.current = true
        }, 1500)
      })
      .catch(() => {
        // Browser denied (no user gesture context). Arm anyway so monitoring works.
        setTimeout(() => {
          armedRef.current = true
        }, 2000)
      })

    // ── Fullscreen exit ──
    const handleFullscreenChange = () => {
      if (!document.fullscreenElement && armedRef.current) {
        recordViolation("fullscreen_exit")
        // Attempt to re-enter fullscreen
        setTimeout(() => {
          document.documentElement.requestFullscreen?.().catch(() => {})
        }, 400)
      }
    }

    // ── Tab switch ──
    const handleVisibilityChange = () => {
      if (document.hidden && armedRef.current) {
        recordViolation("tab_switch")
      }
    }

    // ── Window blur (Alt+Tab) ──
    const handleBlur = () => {
      if (armedRef.current) {
        recordViolation("window_blur")
      }
    }

    // ── Clipboard blocking ──
    const handleCopy = (e: Event) => {
      if (!armedRef.current) return
      e.preventDefault()
      recordViolation("copy_blocked")
    }
    const handlePaste = (e: Event) => {
      if (!armedRef.current) return
      e.preventDefault()
      recordViolation("paste_blocked")
    }
    const handleCut = (e: Event) => {
      if (!armedRef.current) return
      e.preventDefault()
      recordViolation("cut_blocked")
    }

    // ── Right-click ──
    const handleContextMenu = (e: Event) => {
      if (!armedRef.current) return
      e.preventDefault()
      recordViolation("contextmenu_blocked")
    }

    // ── Keyboard shortcuts ──
    const handleKeyDown = (e: KeyboardEvent) => {
      if (!armedRef.current) return
      const ctrl = e.ctrlKey || e.metaKey

      // F12
      if (e.key === "F12") {
        e.preventDefault()
        recordViolation("keyboard_shortcut")
        return
      }

      if (ctrl) {
        // Ctrl+Shift+I / Ctrl+Shift+J (DevTools)
        if (e.shiftKey && /^[ij]$/i.test(e.key)) {
          e.preventDefault()
          recordViolation("keyboard_shortcut")
          return
        }
        // Ctrl+C, Ctrl+V, Ctrl+X — already caught by clipboard events, but also block here
        if (/^[cvx]$/i.test(e.key)) {
          e.preventDefault()
          // clipboard event handlers will log the violation
          return
        }
        // Ctrl+U (view source), Ctrl+T (new tab), Ctrl+L (address bar),
        // Ctrl+R (reload), Ctrl+N (new window), Ctrl+W (close tab)
        if (/^[utlrnw]$/i.test(e.key)) {
          e.preventDefault()
          recordViolation("keyboard_shortcut")
          return
        }
        // Ctrl+Tab
        if (e.key === "Tab") {
          e.preventDefault()
          recordViolation("keyboard_shortcut")
          return
        }
      }

      // Alt+Tab (limited — most OSes handle this before the browser)
      if (e.altKey && e.key === "Tab") {
        e.preventDefault()
        recordViolation("keyboard_shortcut")
      }
    }

    // Register all listeners in capture phase for maximum intercept priority
    document.addEventListener("fullscreenchange", handleFullscreenChange)
    document.addEventListener("visibilitychange", handleVisibilityChange)
    window.addEventListener("blur", handleBlur)
    document.addEventListener("copy", handleCopy, true)
    document.addEventListener("paste", handlePaste, true)
    document.addEventListener("cut", handleCut, true)
    document.addEventListener("contextmenu", handleContextMenu, true)
    document.addEventListener("keydown", handleKeyDown, true)

    const removeListeners = () => {
      document.removeEventListener("fullscreenchange", handleFullscreenChange)
      document.removeEventListener("visibilitychange", handleVisibilityChange)
      window.removeEventListener("blur", handleBlur)
      document.removeEventListener("copy", handleCopy, true)
      document.removeEventListener("paste", handlePaste, true)
      document.removeEventListener("cut", handleCut, true)
      document.removeEventListener("contextmenu", handleContextMenu, true)
      document.removeEventListener("keydown", handleKeyDown, true)
      armedRef.current = false
    }

    cleanupRef.current = removeListeners

    return () => {
      removeListeners()
      if (document.fullscreenElement) {
        document.exitFullscreen().catch(() => {})
      }
    }
  }, [enabled, recordViolation])

  return {
    violationCount: violationCountRef.current,
    cleanup,
  }
}
