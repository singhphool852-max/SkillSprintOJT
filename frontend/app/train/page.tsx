"use client"

import { ProtectedRoute } from "@/components/ProtectedRoute"
import { Activity } from "lucide-react"
import { TopicCard } from "@/components/train/TopicCard"
import { ModeCard } from "@/components/train/ModeCard"
import { NotesUpload } from "@/components/train/NotesUpload"
import { SessionSetupPanel } from "@/components/train/SessionSetupPanel"
import {
  Brain,
  Code2,
  Database,
  Layout,
  MessageSquare,
  Cpu,
  Timer,
  Target,
  UserCheck,
  Zap,
  Swords
} from "lucide-react"
import { useState, useRef, useEffect } from "react"
import { useRouter } from "next/navigation"
import { getTrainingHistory, recommendDifficulty } from "@/lib/training-history"
import { API_URL } from "@/lib/api-config"

const topTopics = [
  {
    title: "DSA",
    description: "Data Structures & Algorithms. Trees, Graphs, and DP combat prep.",
    icon: Code2,
    count: 42,
    mastery: 65,
    difficulty: "Hard" as const,
    lastScore: 78,
    color: "pink" as const
  },
  {
    title: "DBMS",
    description: "Database Management. SQL, Normalization, and ACID architecture.",
    icon: Database,
    count: 24,
    mastery: 42,
    difficulty: "Medium" as const,
    lastScore: 85,
    color: "cyan" as const
  },
  {
    title: "OS",
    description: "Operating Systems. Process Sync, Memory, and Kernel logic.",
    icon: Cpu,
    count: 18,
    mastery: 15,
    difficulty: "Medium" as const,
    lastScore: 40,
    color: "yellow" as const
  },
  {
    title: "JAVASCRIPT",
    description: "JS Engines, Closures, Event Loop, and Web API mastery.",
    icon: Layout,
    count: 35,
    mastery: 88,
    difficulty: "Easy" as const,
    lastScore: 92,
    color: "cyan" as const
  },
  {
    title: "APTITUDE",
    description: "Quantitative and Logical reasoning drills for speed.",
    icon: Target,
    count: 50,
    mastery: 30,
    difficulty: "Easy" as const,
    lastScore: 65,
    color: "yellow" as const
  },
]

const trainingModes = [
  {
    title: "PRACTICE MODE",
    description: "Lower pressure environment focused on learning and accuracy. Helpful hints enabled.",
    icon: Brain,
    duration: "Unlimited",
    difficulty: "Beginner" as const,
    color: "cyan" as const
  },
  {
    title: "SPEED MODE",
    description: "High-intensity technical drills with strict per-question timers. Focus on instinct.",
    icon: Timer,
    duration: "5-min",
    difficulty: "Advanced" as const,
    color: "yellow" as const
  },
  {
    title: "TARGET MODE",
    description: "Fully customizable session. Select specific sub-topics, difficulty, and question count.",
    icon: Target,
    duration: "Flexible",
    difficulty: "Intermediate" as const,
    color: "cyan" as const
  },
  {
    title: "MOCK INTERVIEW",
    description: "End-to-end simulated technical interview. Mixed domains, live pressure, and AI review.",
    icon: UserCheck,
    duration: "45-min",
    difficulty: "Advanced" as const,
    color: "pink" as const,
    isPremium: true
  },
]

export default function TrainPage() {
  const router = useRouter()
  const operationsRef = useRef<HTMLDivElement>(null)

  const [selectedTopic, setSelectedTopic] = useState<string | null>(null)
  const [selectedMode, setSelectedMode] = useState<string | null>(null)
  const [showSetup, setShowSetup] = useState(false)
  const [adaptiveDifficulty, setAdaptiveDifficulty] = useState<string>("Medium")

  // AI Generator State
  const [aiTopic, setAiTopic] = useState("")
  const [aiDifficulty, setAiDifficulty] = useState("Medium")
  const [aiCount, setAiCount] = useState<number>(5)
  const [aiLoading, setAiLoading] = useState(false)
  const [aiError, setAiError] = useState<string | null>(null)
  const [aiDebugInfo, setAiDebugInfo] = useState<{ isRealAI: boolean; provider: string } | null>(null)

  const handleGenerateAI = async () => {
    if (!aiTopic.trim()) {
      setAiError("DOMAIN SPECIFICATION REQUIRED");
      return;
    }
    setAiError(null);
    setAiLoading(true);

    const url = `${API_URL}/api/training/generate`;

    const payload = { 
      topic: aiTopic.toLowerCase(),
      difficulty: aiDifficulty.toLowerCase(), 
      count: aiCount 
    };

    console.log("[AI_GEN] URL:", url);
    console.log("[AI_GEN] Payload:", payload);

    try {
      const res = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
        credentials: "include"
      });

      console.log("[AI_GEN] Status:", res.status);

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        const errorMsg = errData.error || `Generation failed (${res.status})`;
        console.error("[AI_GEN] API error:", errorMsg);
        
        if (res.status === 401) {
          setAiError("UNAUTHORIZED: PLEASE SIGN IN TO GENERATE AI SESSIONS");
        } else {
          setAiError(errorMsg);
        }
        setAiLoading(false);
        return;
      }

      const data = await res.json();
      console.log("[AI_GEN] Data Received:", data);

      if (!data.sessionId && !data.session_id) {
        console.error("[AI_GEN] No session ID in response");
        setAiError("AI generation failed - no session created");
        setAiLoading(false);
        return;
      }

      if (!data.questions || data.questions.length === 0) {
        console.error("[AI_GEN] No questions in response");
        setAiError("AI generation failed - no questions generated");
        setAiLoading(false);
        return;
      }

      const sessionId = data.sessionId || data.session_id;
      console.log("[AI_GEN] Redirecting to play:", sessionId);
      router.push(`/train/play/${sessionId}?topic=${encodeURIComponent(aiTopic)}&mode=AI_SYNTH_MODE&difficulty=${encodeURIComponent(aiDifficulty)}&count=${aiCount}`);
    } catch (err: any) {
      console.error("[AI_GEN] Fetch Exception:", err?.message);
      // DISTINGUISH: Network Error vs App Error
      if (err?.message?.includes("fetch") || err?.message?.includes("NetworkError")) {
        setAiError("CONNECTIVITY FAILURE: BACKEND UNREACHABLE.");
      } else {
        setAiError(err?.message || "AI SERVICES OFFLINE.");
      }
    } finally {
      setAiLoading(false);
    }
  };

  // Auto-scroll logic and adaptive difficulty calculation
  useEffect(() => {
    if (selectedTopic) {
      operationsRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' })

      // Calculate Adaptive Difficulty
      const history = getTrainingHistory().filter(h => h.topic === selectedTopic)
      if (history.length > 0) {
        const avgAcc = history.map(h => h.accuracy).reduce((a, b) => a + b, 0) / history.length
        const rec = recommendDifficulty(avgAcc, "Medium")
        setAdaptiveDifficulty(rec.newDifficulty)
      } else {
        setAdaptiveDifficulty("Medium") // Default baseline
      }
    }
  }, [selectedTopic])

  const handleStart = (modeOverride?: string) => {
    const topic = selectedTopic
    const mode = modeOverride || selectedMode

    if (!topic || !mode) return

    if (mode === "TARGET MODE") {
      setShowSetup(true)
      return
    }

    // Standardized Mapping: Ensure frontend IDs match backend combat arenas
    const topicIDMap: Record<string, string> = {
      "DSA": "dsa",
      "DBMS": "dbms",
      "OS": "os",
      "JAVASCRIPT": "javascript",
      "APTITUDE": "aptitude"
    }

    const arenaId = topicIDMap[topic] || topic.toLowerCase().replace(/\s+/g, '_')
    console.log("[TrainHub] Session Starting:", { topic, arenaId, mode })
    router.push(`/train/play/${arenaId}?topic=${encodeURIComponent(topic)}&mode=${encodeURIComponent(mode)}&difficulty=${encodeURIComponent(adaptiveDifficulty)}`)
  }

  const handleInitialize = (quizId: string) => {
    if (!selectedTopic || !selectedMode) return
    router.push(`/train/play/${quizId}?topic=${encodeURIComponent(selectedTopic)}&mode=${encodeURIComponent(selectedMode)}&difficulty=${encodeURIComponent(adaptiveDifficulty)}`)
  }

  return (
    <ProtectedRoute>
      <div className="relative min-h-screen">
        <div className="absolute inset-0 grid-bg opacity-40 pointer-events-none" />

        <div className="relative z-10 mx-auto max-w-7xl px-4 py-8 lg:px-8">
          {/* Hero Section */}
          <div className="flex flex-col gap-4 mb-12 lg:mb-16">
            <div className="flex items-center gap-3 mb-2">
              <div className="h-2 w-2 rounded-full bg-neon-cyan animate-pulse-glow" />
              <span className="font-mono text-[10px] tracking-[0.3em] text-neon-cyan uppercase">
                TRAINING DIVISION // ACTIVE
              </span>
            </div>

            <div className="flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6">
              <div className="max-w-3xl">
                <h1 className="text-4xl font-black uppercase tracking-tight text-foreground sm:text-5xl lg:text-6xl mb-6">
                  DEVELOP YOUR <span className="text-neon-cyan text-glow-cyan">EDGE</span>
                </h1>
                <p className="font-mono text-sm leading-relaxed text-muted-foreground lg:text-base max-w-2xl">
                  Welcome to the Training Hub. Hone your technical skills, master advanced architecture, and simulate real-world interview conditions in a high-intensity environment.
                </p>
              </div>

              <div className="flex gap-4 p-4 border border-panel-border bg-panel-bg/40 backdrop-blur-sm self-start lg:self-auto">
                <div className="flex flex-col border-r border-panel-border pr-6">
                  <span className="font-mono text-[9px] text-muted-foreground uppercase mb-1">Weekly Intensity</span>
                  <span className="font-mono text-xl font-bold text-neon-cyan tracking-wider">12.4H</span>
                </div>
                <div className="flex flex-col pl-2">
                  <span className="font-mono text-[9px] text-muted-foreground uppercase mb-1">Global Tier</span>
                  <span className="font-mono text-xl font-bold text-neon-pink tracking-wider">Lvl 14</span>
                </div>
              </div>
            </div>
          </div>

          {/* Status Telemetry */}
          <div className="mb-12 flex items-center gap-4 bg-white/[0.02] border border-panel-border p-4">
            <div className={`h-2 w-2 rounded-full ${selectedTopic ? 'bg-neon-cyan animate-pulse' : 'bg-muted-foreground'}`} />
            <div className="flex-1 font-mono text-[10px] tracking-[0.2em] uppercase">
              {selectedTopic
                ? `[ STATUS: DOMAIN ${selectedTopic} SYNCHRONIZED // RECOMMENDED INTENSITY: ${adaptiveDifficulty} ]`
                : "[ STATUS: WAITING FOR NEURAL INGESTION // SELECT DOMAIN ]"
              }
            </div>
            {selectedTopic && (
              <button
                onClick={() => { setSelectedTopic(null); setSelectedMode(null); }}
                className="font-mono text-[9px] text-neon-pink hover:underline uppercase tracking-widest"
              >
                Change Domain
              </button>
            )}
          </div>

          {/* Section: Core Topics */}
          <div className="mb-16">
            <div className="flex items-center gap-4 mb-8">
              <span className="font-mono text-[10px] font-bold text-neon-cyan border border-neon-cyan px-2 py-0.5">PHASE 01</span>
              <Brain className="h-4 w-4 text-neon-cyan" />
              <span className="font-mono text-[11px] tracking-[0.3em] text-foreground uppercase">
                CORE TRAINING DOMAINS
              </span>
              <div className="h-px flex-1 bg-panel-border" />
            </div>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5">
              {topTopics.map((topic) => (
                <TopicCard
                  key={topic.title}
                  {...topic}
                  isActive={selectedTopic === topic.title}
                  onClick={() => setSelectedTopic(topic.title)}
                />
              ))}
            </div>
          </div>

          {/* Section: Advanced Modes */}
          <div ref={operationsRef} className={`mb-16 transition-all duration-700 ${!selectedTopic ? 'opacity-30 grayscale pointer-events-none' : 'opacity-100 grayscale-0'}`}>
            <div className="flex items-center gap-4 mb-8">
              <span className={`font-mono text-[10px] font-bold border px-2 py-0.5 ${selectedTopic ? 'text-neon-yellow border-neon-yellow' : 'text-muted-foreground border-panel-border'}`}>PHASE 02</span>
              <Swords className={`h-4 w-4 ${selectedTopic ? 'text-neon-yellow animate-pulse-glow' : 'text-muted-foreground'}`} />
              <span className="font-mono text-[11px] tracking-[0.3em] text-foreground uppercase">
                SPECIAL OPERATIONS
              </span>
              <div className="h-px flex-1 bg-panel-border" />
            </div>
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {trainingModes.map((mode) => (
                <ModeCard
                  key={mode.title}
                  {...mode}
                  isActive={selectedMode === mode.title}
                  onClick={() => {
                    setSelectedMode(mode.title)
                    if (selectedTopic) handleStart(mode.title)
                  }}
                />
              ))}
            </div>
          </div>

          {/* Section: AI Custom Generation */}
          <div className="mb-16">
            <div className="flex items-center gap-4 mb-8">
              <span className="font-mono text-[10px] font-bold border border-neon-pink text-neon-pink px-2 py-0.5">PHASE 03</span>
              <Activity className="h-4 w-4 text-neon-pink animate-pulse" />
              <span className="font-mono text-[11px] tracking-[0.3em] text-foreground uppercase">
                GENERATE AI TRAINING SESSION
              </span>
              <div className="h-px flex-1 bg-panel-border" />
            </div>

            <div className="border border-neon-pink/20 bg-neon-pink/5 backdrop-blur-sm p-6 lg:p-8 max-w-4xl mx-auto rounded-none relative overflow-hidden">
              {/* Cyberpunk aesthetics */}
              <div className="absolute top-0 right-0 w-32 h-32 bg-neon-pink opacity-[0.03] blur-3xl rounded-full" />
              <div className="absolute top-0 right-0 p-4 font-mono text-[83px] font-black text-neon-pink opacity-[0.02] tracking-tighter mix-blend-overlay">AI</div>

              <div className="grid gap-6 md:grid-cols-3 mb-8 relative z-10">
                <div>
                  <label className="block font-mono text-[10px] text-muted-foreground tracking-widest uppercase mb-2">DOMAIN TOPIC</label>
                  <select
                    className="w-full bg-deep-bg/50 border border-panel-border px-4 py-3 font-mono text-xs text-foreground focus:border-neon-pink focus:outline-none appearance-none"
                    value={aiTopic}
                    onChange={(e) => setAiTopic(e.target.value)}
                  >
                    <option value="" disabled>SELECT DOMAIN...</option>
                    <option value="DSA">DSA</option>
                    <option value="DBMS">DBMS</option>
                    <option value="OS">OPERATING SYSTEMS</option>
                    <option value="JAVASCRIPT">JAVASCRIPT</option>
                    <option value="APTITUDE">APTITUDE</option>
                    <option value="SYSTEM DESIGN">SYSTEM DESIGN</option>
                  </select>
                </div>
                <div>
                  <label className="block font-mono text-[10px] text-muted-foreground tracking-widest uppercase mb-2">INTENSITY MULTIPLIER</label>
                  <select
                    className="w-full bg-deep-bg/50 border border-panel-border px-4 py-3 font-mono text-xs text-foreground focus:border-neon-pink focus:outline-none appearance-none"
                    value={aiDifficulty}
                    onChange={(e) => setAiDifficulty(e.target.value)}
                  >
                    <option value="Easy">LEVEL 01: EASY</option>
                    <option value="Medium">LEVEL 02: MODERATE</option>
                    <option value="Hard">LEVEL 03: EXTREME</option>
                  </select>
                </div>
                <div>
                  <label className="block font-mono text-[10px] text-muted-foreground tracking-widest uppercase mb-2">OPERATION COUNT ({aiCount})</label>
                  <input
                    type="range"
                    min="1"
                    max="15"
                    className="w-full h-1 mt-4 bg-panel-border appearance-none cursor-pointer accent-neon-pink"
                    value={aiCount}
                    onChange={(e) => setAiCount(Number(e.target.value))}
                  />
                  <div className="flex justify-between font-mono text-[9px] text-muted-foreground mt-2">
                    <span>1</span>
                    <span>15</span>
                  </div>
                </div>
              </div>

              <div className="flex flex-col items-center gap-3 relative z-10">
                <button
                  onClick={handleGenerateAI}
                  disabled={aiLoading}
                  className="group relative flex items-center gap-3 border border-neon-pink bg-neon-pink/10 px-10 py-5 font-mono text-xs font-bold tracking-[0.3em] text-neon-pink transition-all hover:bg-neon-pink/20 hover:shadow-[0_0_20px_rgba(255,0,255,0.2)] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <div className="absolute inset-0 w-full h-full bg-gradient-to-r from-neon-pink/0 via-neon-pink/10 to-neon-pink/0 -translate-x-[100%] group-hover:translate-x-[100%] transition-transform duration-1000" />
                  {aiLoading ? (
                    <div className="flex items-center gap-3">
                      <Activity className="h-4 w-4 animate-spin" />
                      COMPILING MATRICES...
                    </div>
                  ) : (
                    "GENERATE AI SESSION"
                  )}
                </button>
                {aiError && (
                  <div className="font-mono text-[10px] text-neon-pink uppercase tracking-widest mt-2 px-3 py-1 bg-neon-pink/10 border border-neon-pink/20">
                    {aiError}
                  </div>
                )}
                {aiDebugInfo && (
                  <div className={`mt-2 px-3 py-1 font-mono text-[10px] font-bold border tracking-widest ${aiDebugInfo.isRealAI ? 'bg-green-500/10 text-green-400 border-green-500/20' : 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20'}`}>
                    {aiDebugInfo.isRealAI ? '[ REAL AI MODE ]' : '[ FALLBACK MODE ]'} - {aiDebugInfo.provider}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Section: Notes / Upload */}
          <div className="mb-16">
            <NotesUpload />
          </div>

          {/* Ready status */}
          <div className="mt-20 flex flex-col items-center gap-6 pb-12">
            <div className="flex items-center gap-4">
              <div className="h-px w-16 bg-panel-border" />
              <div className="flex h-10 w-10 items-center justify-center border border-panel-border bg-panel-bg/40 animate-pulse">
                <Target className="h-5 w-5 text-neon-cyan" />
              </div>
              <div className="h-px w-16 bg-panel-border" />
            </div>
            <div className="text-center">
              <h3 className="font-mono text-xs font-bold tracking-[0.4em] text-foreground uppercase mb-2">READY TO SYNC?</h3>
              <p className="text-[10px] font-mono text-muted-foreground uppercase">SYSTEM STATUS: OPTIMAL // ALL MODULES LOADED</p>
            </div>
          </div>
        </div>

        {/* Setup Overlay */}
        {showSetup && selectedTopic && selectedMode && (
          <SessionSetupPanel
            topic={selectedTopic}
            mode={selectedMode}
            onClose={() => setShowSetup(false)}
            onStart={handleInitialize}
          />
        )}
      </div>
    </ProtectedRoute>
  )
}

