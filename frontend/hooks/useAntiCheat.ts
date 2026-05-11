"use client"

import { useEffect, useRef, useCallback } from "react"

// ─────────────────────────────────────────────────────────────────────────────
// Anti-Cheat Hook (REWRITTEN FOR ISOLATION)
// ─────────────────────────────────────────────────────────────────────────────

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
  // RULE 1 & 6: Use useRef for all mutable state inside hook, NO module-level variables
  const violationCountRef = useRef(0)
  const isArmedRef = useRef(false)
  const lastViolationTimeRef = useRef(0)
  const cleanupFnsRef = useRef<Array<() => void>>([])

  // Stable refs for callbacks to avoid dependency chain complexity
  const onViolationRef = useRef(onViolation)
  const onAutoSubmitRef = useRef(onAutoSubmit)
  onViolationRef.current = onViolation
  onAutoSubmitRef.current = onAutoSubmit

  // ── VIOLATION HANDLER ──
  const addViolation = useCallback((type: ViolationType) => {
    if (!isArmedRef.current) return

    // Debounce: ignore violations within 1.5s
    const now = Date.now()
    if (now - lastViolationTimeRef.current < 1500) return
    lastViolationTimeRef.current = now

    violationCountRef.current += 1
    const count = violationCountRef.current
    onViolationRef.current(type, count)

    if (count >= maxViolations) {
      // Auto-submit after brief delay
      setTimeout(() => {
        onAutoSubmitRef.current()
        // Final cleanup
        cleanup()
      }, 1000)
    }
  }, [maxViolations])

  // RULE 7: Cleanup function removes exactly the listeners this instance added
  const cleanup = useCallback(() => {
    isArmedRef.current = false
    cleanupFnsRef.current.forEach(fn => fn())
    cleanupFnsRef.current = []
    
    if (document.fullscreenElement) {
      document.exitFullscreen().catch(() => {})
    }
  }, [])

  // RULE 8: Initial fullscreen request per component mount
  useEffect(() => {
    if (!enabled) return

    const requestFS = async () => {
      try {
        if (!document.fullscreenElement) {
          await document.documentElement.requestFullscreen()
        }
        // Grace period before arming to avoid false triggers
        setTimeout(() => {
          isArmedRef.current = true
        }, 1500)
      } catch (err) {
        console.warn("[AntiCheat] Fullscreen request failed:", err)
        // Arm anyway so focus monitoring works
        setTimeout(() => {
          isArmedRef.current = true
        }, 2000)
      }
    }

    requestFS()

    return () => {
      // If we are unmounting, we should ideally exit fullscreen if we were the one who requested it
      // but in Arena, we might be switching views. The cleanup() call from parent handles this.
    }
  }, [enabled])

  // RULE 4: Fullscreen change handler isolation and re-prompt
  useEffect(() => {
    if (!enabled) return

    const handleFullscreenChange = () => {
      if (!document.fullscreenElement && isArmedRef.current) {
        addViolation("fullscreen_exit")
        // RULE 3 & 4: Re-prompt fullscreen
        setTimeout(() => {
          if (!document.fullscreenElement && isArmedRef.current) {
            document.documentElement.requestFullscreen().catch(() => {})
          }
        }, 500)
      }
    }

    document.addEventListener("fullscreenchange", handleFullscreenChange)
    const removeListener = () => document.removeEventListener("fullscreenchange", handleFullscreenChange)
    cleanupFnsRef.current.push(removeListener)

    return removeListener
  }, [enabled, addViolation])

  // RULE 5: Visibility and Blur isolation
  useEffect(() => {
    if (!enabled) return

    const handleVisibilityChange = () => {
      if (document.hidden && isArmedRef.current) {
        addViolation("tab_switch")
      }
    }

    const handleBlur = () => {
      if (isArmedRef.current) {
        addViolation("window_blur")
      }
    }

    document.addEventListener("visibilitychange", handleVisibilityChange)
    window.addEventListener("blur", handleBlur)

    const removeListeners = () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange)
      window.removeEventListener("blur", handleBlur)
    }
    cleanupFnsRef.current.push(removeListeners)

    return removeListeners
  }, [enabled, addViolation])

  // ── Keyboard & Context Menu isolation ──
  useEffect(() => {
    if (!enabled) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isArmedRef.current) return
      const ctrl = e.ctrlKey || e.metaKey

      // RULE: Only intercept if Ctrl/Meta/F12 is involved. Plain characters must pass.
      if (!ctrl && e.key !== "F12" && !e.altKey) return

      // F12
      if (e.key === "F12") {
        e.preventDefault()
        addViolation("keyboard_shortcut")
        return
      }

      if (ctrl) {
        // Ctrl+Shift+I / Ctrl+Shift+J (DevTools)
        if (e.shiftKey && /^[ij]$/i.test(e.key)) {
          e.preventDefault()
          addViolation("keyboard_shortcut")
          return
        }
        // Ctrl+C, Ctrl+V, Ctrl+X
        if (/^[cvx]$/i.test(e.key)) {
          e.preventDefault()
          // violation logged by clipboard events
          return
        }
        // Ctrl+U, Ctrl+T, Ctrl+L, Ctrl+R, Ctrl+N, Ctrl+W
        if (/^[utlrnw]$/i.test(e.key)) {
          e.preventDefault()
          addViolation("keyboard_shortcut")
          return
        }
        // Ctrl+Tab
        if (e.key === "Tab") {
          e.preventDefault()
          addViolation("keyboard_shortcut")
          return
        }
      }

      // Alt+Tab
      if (e.altKey && e.key === "Tab") {
        e.preventDefault()
        addViolation("keyboard_shortcut")
      }
    }

    const handleContextMenu = (e: MouseEvent) => {
      if (isArmedRef.current) {
        e.preventDefault()
        addViolation("contextmenu_blocked")
      }
    }

    window.addEventListener("keydown", handleKeyDown, false)
    document.addEventListener("contextmenu", handleContextMenu, true)

    const removeListeners = () => {
      window.removeEventListener("keydown", handleKeyDown, false)
      document.removeEventListener("contextmenu", handleContextMenu, true)
    }
    cleanupFnsRef.current.push(removeListeners)

    return removeListeners
  }, [enabled, addViolation])

  // ── Clipboard isolation ──
  useEffect(() => {
    if (!enabled) return

    const handleCopy = (e: ClipboardEvent) => {
      if (isArmedRef.current) {
        e.preventDefault()
        addViolation("copy_blocked")
      }
    }
    const handlePaste = (e: ClipboardEvent) => {
      if (isArmedRef.current) {
        e.preventDefault()
        addViolation("paste_blocked")
      }
    }
    const handleCut = (e: ClipboardEvent) => {
      if (isArmedRef.current) {
        e.preventDefault()
        addViolation("cut_blocked")
      }
    }

    document.addEventListener("copy", handleCopy, true)
    document.addEventListener("paste", handlePaste, true)
    document.addEventListener("cut", handleCut, true)

    const removeListeners = () => {
      document.removeEventListener("copy", handleCopy, true)
      document.removeEventListener("paste", handlePaste, true)
      document.removeEventListener("cut", handleCut, true)
    }
    cleanupFnsRef.current.push(removeListeners)

    return removeListeners
  }, [enabled, addViolation])

  return {
    violationCount: violationCountRef.current,
    cleanup,
  }
}
