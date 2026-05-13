"use client"

import { useEffect, useRef, useCallback } from "react"

type ViolationType =
  | "fullscreen_exit"
  | "tab_switch"
  | "window_blur"
  | "blocked_key"

interface UseAntiCheatProps {
  onViolation: (type: string, count: number) => void
  onAutoSubmit: () => void
  onShowFullscreenWarning: (show: boolean) => void
  maxViolations?: number
  enabled?: boolean
}

interface UseAntiCheatReturn {
  cleanup: () => void
}

export function useAntiCheat({
  onViolation,
  onAutoSubmit,
  onShowFullscreenWarning,
  maxViolations = 3,
  enabled = true,
}: UseAntiCheatProps): UseAntiCheatReturn {
  // CRITICAL: Use useRef for violation count to avoid stale closure bug
  const violationCountRef = useRef(0)
  const isArmedRef = useRef(false)
  const lastViolationTimeRef = useRef(0)
  
  // Stable refs for callbacks
  const onViolationRef = useRef(onViolation)
  const onAutoSubmitRef = useRef(onAutoSubmit)
  const onShowFullscreenWarningRef = useRef(onShowFullscreenWarning)
  onViolationRef.current = onViolation
  onAutoSubmitRef.current = onAutoSubmit
  onShowFullscreenWarningRef.current = onShowFullscreenWarning

  // ── VIOLATION HANDLER ──
  const addViolation = useCallback((type: ViolationType) => {
    if (!isArmedRef.current) return

    // Debounce: ignore violations within 500ms
    const now = Date.now()
    if (now - lastViolationTimeRef.current < 500) return
    lastViolationTimeRef.current = now

    violationCountRef.current += 1
    const count = violationCountRef.current

    console.log(`[AntiCheat] Violation ${count}/${maxViolations}: ${type}`)

    // Always log to backend
    onViolationRef.current(type, count)

    // Auto-submit on max violations
    if (count >= maxViolations) {
      console.log('[AntiCheat] Max violations reached - auto-submitting')
      onAutoSubmitRef.current()
      cleanup()
    }
  }, [maxViolations])

  // ── CLEANUP ──
  const cleanup = useCallback(() => {
    isArmedRef.current = false
    onShowFullscreenWarningRef.current(false)
    
    if (document.fullscreenElement) {
      document.exitFullscreen().catch(() => {})
    }
  }, [])

  // ── INITIAL FULLSCREEN REQUEST ──
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
          console.log('[AntiCheat] Armed and monitoring')
        }, 1500)
      } catch (err) {
        console.warn("[AntiCheat] Fullscreen request failed:", err)
        // Arm anyway so monitoring works
        setTimeout(() => {
          isArmedRef.current = true
        }, 2000)
      }
    }

    requestFS()
  }, [enabled])

  // ── FULLSCREEN CHANGE HANDLER ──
  useEffect(() => {
    if (!enabled) return

    const handleFullscreenChange = () => {
      if (!document.fullscreenElement && isArmedRef.current) {
        // Exited fullscreen - try to auto re-enter immediately
        console.log('[AntiCheat] Fullscreen exit detected - attempting auto re-enter')
        
        document.documentElement.requestFullscreen()
          .then(() => {
            // Success - fullscreen restored automatically, no violation needed
            console.log('[AntiCheat] Auto re-enter fullscreen succeeded')
            onShowFullscreenWarningRef.current(false)
          })
          .catch(() => {
            // Browser blocked auto fullscreen (requires user gesture)
            // Count as violation AND show warning button
            console.log('[AntiCheat] Auto re-enter failed - counting violation and showing warning')
            addViolation("fullscreen_exit")
            onShowFullscreenWarningRef.current(true)
          })
      } else if (document.fullscreenElement) {
        // Entered fullscreen - hide warning
        console.log('[AntiCheat] Fullscreen entered')
        onShowFullscreenWarningRef.current(false)
      }
    }

    document.addEventListener("fullscreenchange", handleFullscreenChange)
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange)
  }, [enabled, addViolation])

  // ── VISIBILITY CHANGE HANDLER (tab switch) ──
  useEffect(() => {
    if (!enabled) return

    const handleVisibilityChange = () => {
      if (document.hidden && isArmedRef.current) {
        console.log('[AntiCheat] Tab switch detected')
        addViolation("tab_switch")
        // Do NOT pause counting or show blocking modal
        // Just count and continue
      }
    }

    document.addEventListener("visibilitychange", handleVisibilityChange)
    return () => document.removeEventListener("visibilitychange", handleVisibilityChange)
  }, [enabled, addViolation])

  // ── WINDOW BLUR HANDLER (Alt+Tab) ──
  useEffect(() => {
    if (!enabled) return

    const handleBlur = () => {
      if (isArmedRef.current) {
        console.log('[AntiCheat] Window blur detected')
        addViolation("window_blur")
        // Keep counting, do not pause
      }
    }

    window.addEventListener("blur", handleBlur)
    return () => window.removeEventListener("blur", handleBlur)
  }, [enabled, addViolation])

  // ── BLOCKED KEYS HANDLER ──
  useEffect(() => {
    if (!enabled) return

    const handleKeyDown = (e: KeyboardEvent) => {
      if (!isArmedRef.current) return

      // Block common cheating keys
      const blockedKeys = [
        'F12', // DevTools
        'I', // Ctrl+Shift+I (DevTools)
        'J', // Ctrl+Shift+J (Console)
        'C', // Ctrl+Shift+C (Inspect)
        'U', // Ctrl+U (View Source)
      ]

      const isCtrlShift = (e.ctrlKey || e.metaKey) && e.shiftKey
      const isCtrlU = (e.ctrlKey || e.metaKey) && e.key === 'U'

      if (
        e.key === 'F12' ||
        (isCtrlShift && ['I', 'J', 'C'].includes(e.key.toUpperCase())) ||
        isCtrlU
      ) {
        e.preventDefault()
        addViolation("blocked_key")
      }
    }

    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [enabled, addViolation])

  // ── CONTEXT MENU BLOCK ──
  useEffect(() => {
    if (!enabled) return

    const handleContextMenu = (e: MouseEvent) => {
      if (isArmedRef.current) {
        e.preventDefault()
      }
    }

    document.addEventListener("contextmenu", handleContextMenu)
    return () => document.removeEventListener("contextmenu", handleContextMenu)
  }, [enabled])

  return {
    cleanup,
  }
}
