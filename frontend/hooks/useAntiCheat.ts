"use client"

import { useEffect, useRef, useCallback, useState } from "react"

type ViolationType =
  | "fullscreen_exit"
  | "tab_switch"
  | "window_blur"

interface UseAntiCheatProps {
  onViolation: (type: string, count: number) => void
  onAutoSubmit: () => void
  maxViolations?: number
  enabled?: boolean
}

interface UseAntiCheatReturn {
  violationCount: number
  warningLevel: number
  showWarningModal: boolean
  handleWarningAcknowledge: () => void
  cleanup: () => void
}

export function useAntiCheat({
  onViolation,
  onAutoSubmit,
  maxViolations = 3,
  enabled = true,
}: UseAntiCheatProps): UseAntiCheatReturn {
  // CRITICAL: Use useRef for violation count to avoid stale closure bug
  const violationCountRef = useRef(0)
  const isArmedRef = useRef(false)
  const lastViolationTimeRef = useRef(0)
  const isShowingWarningRef = useRef(false)
  
  const [warningLevel, setWarningLevel] = useState(0)
  const [showWarningModal, setShowWarningModal] = useState(false)

  // Stable refs for callbacks
  const onViolationRef = useRef(onViolation)
  const onAutoSubmitRef = useRef(onAutoSubmit)
  onViolationRef.current = onViolation
  onAutoSubmitRef.current = onAutoSubmit

  // ── VIOLATION HANDLER ──
  const addViolation = useCallback((type: ViolationType) => {
    if (!isArmedRef.current) return

    // Debounce: ignore violations within 1 second
    const now = Date.now()
    if (now - lastViolationTimeRef.current < 1000) return
    lastViolationTimeRef.current = now

    violationCountRef.current += 1
    const count = violationCountRef.current

    console.log(`[AntiCheat] Violation ${count}/${maxViolations}: ${type}`)

    // Always log to backend
    onViolationRef.current(type, count)

    if (count === 1) {
      // First violation - show warning 1
      setWarningLevel(1)
      setShowWarningModal(true)
      isShowingWarningRef.current = true
    } else if (count === 2) {
      // Second violation - show warning 2 (final warning)
      setWarningLevel(2)
      setShowWarningModal(true)
      isShowingWarningRef.current = true
    } else if (count >= maxViolations) {
      // Third violation - AUTO SUBMIT IMMEDIATELY
      setShowWarningModal(false)
      isShowingWarningRef.current = false
      console.log('[AntiCheat] Max violations reached - auto-submitting')
      
      // Brief delay to show final toast, then auto-submit
      setTimeout(() => {
        onAutoSubmitRef.current()
      }, 500)
    }
  }, [maxViolations])

  // ── WARNING MODAL ACKNOWLEDGE ──
  const handleWarningAcknowledge = useCallback(() => {
    setShowWarningModal(false)
    isShowingWarningRef.current = false
    
    // If warning 1 (fullscreen exit), request fullscreen again
    if (warningLevel === 1 && !document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(err => {
        console.warn('[AntiCheat] Failed to re-enter fullscreen:', err)
      })
    }
  }, [warningLevel])

  // ── CLEANUP ──
  const cleanup = useCallback(() => {
    isArmedRef.current = false
    setShowWarningModal(false)
    isShowingWarningRef.current = false
    
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
      // Ignore fullscreen changes caused by warning modal
      if (isShowingWarningRef.current) {
        console.log('[AntiCheat] Fullscreen change ignored (warning modal active)')
        return
      }

      if (!document.fullscreenElement && isArmedRef.current) {
        // Exited fullscreen - violation
        console.log('[AntiCheat] Fullscreen exit detected')
        addViolation("fullscreen_exit")
      } else if (document.fullscreenElement) {
        // Entered fullscreen - clear any warnings
        console.log('[AntiCheat] Fullscreen entered')
      }
    }

    document.addEventListener("fullscreenchange", handleFullscreenChange)
    return () => document.removeEventListener("fullscreenchange", handleFullscreenChange)
  }, [enabled, addViolation])

  // ── VISIBILITY CHANGE HANDLER (tab switch) ──
  useEffect(() => {
    if (!enabled) return

    const handleVisibilityChange = () => {
      if (document.hidden && isArmedRef.current && !isShowingWarningRef.current) {
        console.log('[AntiCheat] Tab switch detected')
        addViolation("tab_switch")
      }
    }

    document.addEventListener("visibilitychange", handleVisibilityChange)
    return () => document.removeEventListener("visibilitychange", handleVisibilityChange)
  }, [enabled, addViolation])

  // NOTE: Removed window blur listener to prevent false positives
  // (clicking address bar, opening DevTools, etc.)

  return {
    violationCount: violationCountRef.current,
    warningLevel,
    showWarningModal,
    handleWarningAcknowledge,
    cleanup,
  }
}
