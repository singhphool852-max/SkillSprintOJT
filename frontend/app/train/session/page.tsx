"use client"

import { useEffect, useState, Suspense } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import { Loader2, Zap, Brain, Target, Timer, UserCheck } from "lucide-react"
import { API_URL } from "@/lib/api-config"

export default function SessionInitPage() {
  return (
    <Suspense fallback={<div className="flex items-center justify-center min-h-screen bg-deep-bg"><span className="font-mono text-xs tracking-widest text-neon-cyan animate-pulse">LOADING...</span></div>}>
      <SessionInitContent />
    </Suspense>
  )
}

function SessionInitContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const topic = searchParams.get("topic")
  const mode = searchParams.get("mode")
  const difficulty = searchParams.get("difficulty") || "Medium"
  const count = searchParams.get("count") || "10"
  
  const [status, setStatus] = useState("INITIALIZING NEURAL LINK...")
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    async function initializeSession() {
      if (!topic || !mode) {
        setError("MISSING SESSION PARAMETERS")
        return
      }

      try {
        // 1. Fetch available arenas
        setStatus(`SEARCHING FOR ${topic.toUpperCase()} BATTLEFIELD...`)
        const res = await fetch(`${API_URL}/api/arenas`)
        if (!res.ok) throw new Error("COULD NOT CONNECT TO ARENA SERVER")
        
        const arenas = await res.json()
        
        // 2. Find best match for topic
        // Fallback to first arena if no exact match (as discussed in plan)
        const match = arenas.find((a: any) => 
          a.title.toLowerCase().includes(topic.toLowerCase()) || 
          (a.category?.name && a.category.name.toLowerCase().includes(topic.toLowerCase()))
        ) || arenas[0]

        if (!match) {
          throw new Error("NO ACTIVE ARENAS AVAILABLE")
        }

        setStatus(`SYNCHRONIZING ${mode.toUpperCase()} CONFIGURATION...`)
        
        // Simulate high-tech delay for vibe
        await new Promise(r => setTimeout(r, 2000))

        // 3. Redirect to the specialized training play page with the config
        router.push(`/train/play/${match.id}?topic=${encodeURIComponent(topic)}&mode=${encodeURIComponent(mode)}&difficulty=${difficulty}&count=${count}`)

      } catch (err: any) {
        setError(err.message || "UNABLE TO INITIALIZE SESSION")
        setTimeout(() => router.push("/train"), 3000)
      }
    }

    initializeSession()
  }, [topic, mode, router])

  const getModeIcon = () => {
    switch (mode?.toUpperCase()) {
      case "PRACTICE MODE": return Brain
      case "SPEED MODE": return Timer
      case "TARGET MODE": return Target
      case "MOCK INTERVIEW": return UserCheck
      default: return Zap
    }
  }

  const ModeIcon = getModeIcon()

  return (
    <ProtectedRoute>
      <div className="relative min-h-screen flex flex-col items-center justify-center bg-deep-bg overflow-hidden">
        <div className="absolute inset-0 grid-bg opacity-20 pointer-events-none" />
        
        {/* Decorative scanline */}
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-neon-cyan/5 to-transparent h-20 w-full animate-scanline pointer-events-none" />

        <div className="relative z-10 flex flex-col items-center text-center px-4 max-w-xl">
          {!error ? (
            <>
              <div className="relative mb-12">
                <div className="absolute inset-0 bg-neon-cyan/20 blur-3xl rounded-full animate-pulse" />
                <div className="relative flex h-24 w-24 items-center justify-center border border-neon-cyan/40 bg-white/5">
                  <ModeIcon className="h-10 w-10 text-neon-cyan animate-pulse-glow" />
                </div>
                
                {/* Orbiting dots */}
                <div className="absolute -inset-4 border border-dashed border-neon-cyan/20 rounded-full animate-spin-slow" />
              </div>

              <div className="space-y-4">
                <div className="flex items-center justify-center gap-3">
                  <Loader2 className="h-4 w-4 text-neon-cyan animate-spin" />
                  <span className="font-mono text-xs tracking-[0.4em] text-neon-cyan uppercase">
                    {status}
                  </span>
                </div>
                
                <h1 className="text-2xl font-black tracking-tight text-foreground uppercase italic">
                  PREPARING {topic || 'SESSION'}
                </h1>
                
                <p className="font-mono text-[10px] text-muted-foreground uppercase tracking-widest leading-relaxed opacity-60">
                  Topic: {topic} // Mode: {mode} <br />
                  Intensity: {difficulty} // Stack: {count} Qs
                </p>
              </div>
            </>
          ) : (
            <div className="space-y-6">
               <div className="relative flex h-20 w-20 items-center justify-center border border-neon-pink/40 bg-neon-pink/5 mx-auto">
                  <Zap className="h-10 w-10 text-neon-pink" />
               </div>
               <div className="space-y-2">
                  <h2 className="font-mono text-sm font-bold text-neon-pink uppercase tracking-widest">SESSION_ABORTED</h2>
                  <p className="font-mono text-[10px] text-muted-foreground uppercase">{error}</p>
               </div>
               <p className="font-mono text-[9px] text-muted-foreground uppercase animate-pulse">REVERTING TO HUB IN 3s...</p>
            </div>
          )}
        </div>

        {/* Binary rain background mockup */}
        <div className="absolute bottom-10 left-10 flex flex-col font-mono text-[8px] text-neon-cyan/20 pointer-events-none hidden lg:flex">
          <span>01101000 01101111 01101110 01100101</span>
          <span>01111001 01101111 01110101 01110010</span>
          <span>01110011 01101011 01101001 01101100</span>
          <span>01101100 01110011 00101110 00101110</span>
        </div>
      </div>
    </ProtectedRoute>
  )
}

