"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import {
  ChevronRight,
  Clock,
  Eye,
  EyeOff,
  FileText,
  Loader2,
  Plus,
  Shield,
  Target,
  Trash2,
  Zap,
} from "lucide-react"
import { API_URL } from "@/lib/api-config"

const API = API_URL

interface Topic {
  id: string
  name: string
  slug: string
}

interface Test {
  id: string
  title: string
  description: string
  topicId: string
  startTime: string
  durationSeconds: number
  isPublished: boolean
  isActive: boolean
  createdBy: string
  topic?: Topic
}

export default function AdminTestsPage() {
  const [tests, setTests] = useState<Test[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [creating, setCreating] = useState(false)

  // Form state
  const [title, setTitle] = useState("")
  const [description, setDescription] = useState("")
  const [topicId, setTopicId] = useState("")
  const [difficulty, setDifficulty] = useState("easy")
  const [startTime, setStartTime] = useState("")
  const [duration, setDuration] = useState(3600)

  // Topics
  const [topics, setTopics] = useState<Topic[]>([])

  const [publishLoading, setPublishLoading] = useState<string | null>(null)
  const [activateLoading, setActivateLoading] = useState<string | null>(null)

  // AI Build Test state
  const [showAIModal, setShowAIModal] = useState(false)
  const [aiFile, setAiFile] = useState<File | null>(null)
  const [aiGenerating, setAiGenerating] = useState(false)
  const [aiError, setAiError] = useState<string | null>(null)

  async function fetchTests() {
    try {
      const res = await fetch(`${API}/api/admin/tests`, { credentials: "include" })
      if (res.ok) setTests(await res.json())
    } catch (e) {
      console.error("Failed to fetch tests:", e)
    } finally {
      setLoading(false)
    }
  }

  async function fetchTopics() {
    try {
      const res = await fetch(`${API}/api/admin/topics`, { credentials: "include" })
      if (res.ok) setTopics(await res.json())
    } catch (e) {
      console.error("Failed to fetch topics:", e)
    }
  }

  useEffect(() => {
    fetchTests()
    fetchTopics()
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim() || !startTime) return
    setCreating(true)
    try {
      const res = await fetch(`${API}/api/admin/tests`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          title: title.trim(),
          description: description.trim(),
          topicId: topicId || undefined,
          difficulty: difficulty,
          startTime: new Date(startTime).toISOString(),
          durationSeconds: duration,
        }),
      })
      if (res.ok) {
        setTitle("")
        setDescription("")
        setTopicId("")
        setStartTime("")
        setDuration(3600)
        setShowForm(false)
        await fetchTests()
      }
    } catch (e) {
      console.error("Create failed:", e)
    } finally {
      setCreating(false)
    }
  }

  async function handleTogglePublish(testId: string, currentPublished: boolean) {
    setPublishLoading(testId)
    try {
      await fetch(`${API}/api/admin/tests/${testId}/publish`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ isPublished: !currentPublished }),
      })
      await fetchTests()
    } catch (e) {
      console.error("Publish toggle failed:", e)
    } finally {
      setPublishLoading(null)
    }
  }

  async function handleToggleActivate(testId: string) {
    setActivateLoading(testId)
    try {
      await fetch(`${API}/api/admin/tests/${testId}/activate`, {
        method: "PATCH",
        credentials: "include",
      })
      await fetchTests()
    } catch (e) {
      console.error("Activate failed:", e)
    } finally {
      setActivateLoading(null)
    }
  }

  async function handleDelete(testId: string, title: string) {
    if (!confirm(`Are you sure you want to permanently delete "${title}"? This will remove all questions and student attempts.`)) {
      return
    }

    try {
      const res = await fetch(`${API}/api/admin/tests/${testId}`, {
        method: "DELETE",
        credentials: "include",
      })
      if (res.ok) {
        await fetchTests()
      } else {
        alert("Failed to delete test. Check logs.")
      }
    } catch (e) {
      console.error("Delete failed:", e)
    }
  }

  async function handleAIBuildTest(e: React.FormEvent) {
    e.preventDefault()
    if (!aiFile) return

    setAiGenerating(true)
    setAiError(null)

    try {
      const formData = new FormData()
      formData.append("file", aiFile)

      const token = localStorage.getItem("token")
      const res = await fetch(`${API}/api/admin/ai/build-test`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        credentials: "include",
        body: formData,
      })

      if (res.ok) {
        const data = await res.json()
        setShowAIModal(false)
        setAiFile(null)
        await fetchTests()
        
        // Redirect to edit page
        window.location.href = `/admin/tests/${data.testId}`
      } else {
        const error = await res.json()
        setAiError(error.error || "Failed to generate test")
      }
    } catch (e) {
      console.error("AI generation failed:", e)
      setAiError("Network error. Please try again.")
    } finally {
      setAiGenerating(false)
    }
  }

  return (
    <div className="relative min-h-screen">
      <div className="absolute inset-0 grid-bg opacity-20" />

      <div className="relative z-10 px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <Shield className="h-4 w-4 text-neon-pink animate-pulse-glow" />
              <span className="font-mono text-[10px] tracking-[0.3em] text-neon-pink">
                TEST MANAGEMENT
              </span>
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
              ALL <span className="text-neon-pink text-glow-pink">TESTS</span>
            </h1>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => setShowAIModal(true)}
              className="flex items-center gap-2 border border-neon-cyan/50 bg-neon-cyan/10 hover:bg-neon-cyan/20 px-5 py-2.5 font-mono text-[11px] font-bold tracking-widest text-neon-cyan transition-all"
            >
              <Zap className="h-3.5 w-3.5" />
              AI BUILD TEST
            </button>
            <button
              onClick={() => setShowForm(!showForm)}
              className="flex items-center gap-2 bg-neon-pink/90 hover:bg-neon-pink px-5 py-2.5 font-mono text-[11px] font-bold tracking-widest text-white transition-all"
            >
              <Plus className="h-3.5 w-3.5" />
              {showForm ? "CANCEL" : "CREATE TEST"}
            </button>
          </div>
        </div>

        {/* Create form */}
        {showForm && (
          <form
            onSubmit={handleCreate}
            className="mb-8 border border-neon-pink/30 bg-panel-bg/60 backdrop-blur-sm p-6"
          >
            <div className="flex items-center gap-2 mb-5">
              <Plus className="h-3.5 w-3.5 text-neon-pink" />
              <span className="font-mono text-[10px] tracking-[0.2em] text-neon-pink uppercase">
                NEW TEST
              </span>
            </div>

            <div className="grid gap-5 sm:grid-cols-3">
              {/* Title */}
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  TITLE *
                </label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  required
                  placeholder="e.g. Data Structures Round 1"
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none placeholder:text-muted-foreground/40"
                />
              </div>

              {/* Topic */}
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  TOPIC
                </label>
                <select
                  value={topicId}
                  onChange={(e) => setTopicId(e.target.value)}
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                >
                  <option value="">— No topic —</option>
                  {topics.map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>

              {/* Difficulty */}
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  DIFFICULTY
                </label>
                <select
                  value={difficulty}
                  onChange={(e) => setDifficulty(e.target.value)}
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                >
                  <option value="easy">EASY</option>
                  <option value="medium">MEDIUM</option>
                  <option value="hard">HARD</option>
                </select>
              </div>
            </div>

            {/* Description */}
            <div className="mt-5 flex flex-col gap-2">
              <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                DESCRIPTION
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={2}
                placeholder="Brief description shown to users in the Arena"
                className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none resize-none placeholder:text-muted-foreground/40"
              />
            </div>

            <div className="mt-5 grid gap-5 sm:grid-cols-2">
              {/* Start time */}
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  START TIME *
                </label>
                <input
                  type="datetime-local"
                  value={startTime}
                  onChange={(e) => setStartTime(e.target.value)}
                  required
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none [color-scheme:dark]"
                />
              </div>

              {/* Duration */}
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  DURATION (MINUTES)
                </label>
                <input
                  type="number"
                  value={Math.floor(duration / 60)}
                  onChange={(e) => setDuration((parseInt(e.target.value) || 0) * 60)}
                  min={1}
                  required
                  placeholder="60"
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none"
                />
              </div>
            </div>

            <div className="mt-5 flex justify-end">
              <button
                type="submit"
                disabled={creating}
                className="flex items-center gap-2 bg-neon-pink/90 hover:bg-neon-pink px-6 py-2 font-mono text-[11px] font-bold tracking-widest text-white transition-all disabled:opacity-50"
              >
                {creating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                CREATE
              </button>
            </div>
          </form>
        )}

        {/* Loading skeleton */}
        {loading && (
          <div className="flex flex-col gap-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 w-full animate-pulse border border-panel-border bg-panel-bg/20" />
            ))}
          </div>
        )}

        {/* Empty */}
        {!loading && tests.length === 0 && (
          <div className="flex flex-col items-center justify-center py-20 gap-4">
            <FileText className="h-10 w-10 text-muted-foreground" />
            <span className="font-mono text-xs tracking-widest text-muted-foreground">
              NO TESTS CREATED YET
            </span>
          </div>
        )}

        {/* Tests table */}
        {!loading && tests.length > 0 && (
          <div className="border border-panel-border bg-panel-bg/40">
            {/* Header */}
            <div className="grid grid-cols-[1fr_120px_160px_80px_80px_160px] gap-3 px-5 py-3 border-b border-panel-border">
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">TITLE</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">TOPIC</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">START TIME</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">DURATION</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">STATUS</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase text-right">ACTIONS</span>
            </div>

            {/* Rows */}
            {tests.map((test) => (
              <div
                key={test.id}
                className="grid grid-cols-[1fr_120px_160px_80px_80px_160px] gap-3 items-center px-5 py-3 border-b border-panel-border/50 last:border-0 hover:bg-neon-pink/5 transition-colors"
              >
                <Link
                  href={`/admin/tests/${test.id}`}
                  className="flex items-center gap-3 group min-w-0"
                >
                  <FileText className="h-4 w-4 text-neon-pink shrink-0" />
                  <span className="font-mono text-sm font-bold tracking-wider text-foreground group-hover:text-neon-pink transition-colors truncate">
                    {test.title.toUpperCase()}
                  </span>
                  <ChevronRight className="h-3 w-3 text-muted-foreground group-hover:text-neon-pink transition-colors shrink-0" />
                </Link>

                <span className="font-mono text-[10px] text-neon-amber truncate">
                  {test.topic?.name || "—"}
                </span>

                <span className="font-mono text-[10px] text-muted-foreground">
                  {new Date(test.startTime).toLocaleString()}
                </span>

                <span className="flex items-center gap-1.5">
                  <Clock className="h-3 w-3 text-muted-foreground" />
                  <span className="font-mono text-[10px] text-muted-foreground">
                    {Math.floor(test.durationSeconds / 60)}m
                  </span>
                </span>

                <div className="flex items-center gap-2">
                  <div className={`h-1.5 w-1.5 rounded-full ${test.isActive ? "bg-neon-green animate-pulse-glow" : test.isPublished ? "bg-neon-cyan animate-pulse-glow" : "bg-muted-foreground"}`} />
                  <span className={`font-mono text-[10px] tracking-wider ${test.isActive ? "text-neon-green" : test.isPublished ? "text-neon-cyan" : "text-muted-foreground"}`}>
                    {test.isActive ? "ACTIVE" : test.isPublished ? "LIVE" : "DRAFT"}
                  </span>
                </div>

                <div className="flex justify-end gap-2">
                  {test.isPublished && (
                    <button
                      onClick={() => handleToggleActivate(test.id)}
                      disabled={activateLoading === test.id || test.isActive}
                      className={`flex items-center gap-1 px-2 py-1 font-mono text-[9px] tracking-widest border transition-all ${
                        test.isActive
                          ? "border-neon-green/30 text-neon-green/60 cursor-default"
                          : "border-neon-green/40 text-neon-green hover:bg-neon-green/10"
                      } disabled:opacity-50`}
                      title={test.isActive ? "Already active" : "Set as active test"}
                    >
                      {activateLoading === test.id ? (
                        <Loader2 className="h-3 w-3 animate-spin" />
                      ) : (
                        <Zap className="h-3 w-3" />
                      )}
                      {test.isActive ? "ACTIVE" : "ACTIVATE"}
                    </button>
                  )}
                  <button
                    onClick={() => handleTogglePublish(test.id, test.isPublished)}
                    disabled={publishLoading === test.id}
                    className={`flex items-center gap-1 px-2 py-1 font-mono text-[9px] tracking-widest border transition-all ${
                      test.isPublished
                        ? "border-muted-foreground/30 text-muted-foreground hover:border-neon-pink/40 hover:text-neon-pink"
                        : "border-neon-cyan/40 text-neon-cyan hover:bg-neon-cyan/10"
                    } disabled:opacity-50`}
                  >
                    {publishLoading === test.id ? (
                      <Loader2 className="h-3 w-3 animate-spin" />
                    ) : test.isPublished ? (
                      <EyeOff className="h-3 w-3" />
                    ) : (
                      <Eye className="h-3 w-3" />
                    )}
                    {test.isPublished ? "UNPUBLISH" : "PUBLISH"}
                  </button>

                  <button
                    onClick={() => handleDelete(test.id, test.title)}
                    className="flex items-center justify-center p-1.5 border border-panel-border text-muted-foreground hover:border-neon-pink/40 hover:text-neon-pink transition-all"
                    title="Delete Test Permanently"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Footer */}
        {tests.length > 0 && (
          <div className="mt-6 flex items-center gap-4">
            <div className="h-px flex-1 bg-panel-border" />
            <div className="flex items-center gap-2">
              <Target className="h-3 w-3 text-muted-foreground" />
              <span className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">
                {tests.length} TEST{tests.length !== 1 ? "S" : ""}
              </span>
            </div>
            <div className="h-px flex-1 bg-panel-border" />
          </div>
        )}

        {/* AI Build Test Modal */}
        {showAIModal && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
            <div className="relative w-full max-w-lg border border-neon-cyan/30 bg-panel-bg p-8">
              <button
                onClick={() => {
                  setShowAIModal(false)
                  setAiFile(null)
                  setAiError(null)
                }}
                className="absolute top-4 right-4 text-muted-foreground hover:text-neon-pink transition-colors"
              >
                <Plus className="h-4 w-4 rotate-45" />
              </button>

              <div className="flex items-center gap-3 mb-6">
                <Zap className="h-5 w-5 text-neon-cyan animate-pulse-glow" />
                <h2 className="text-xl font-bold tracking-tight text-foreground">
                  AI BUILD TEST
                </h2>
              </div>

              <p className="text-sm text-muted-foreground mb-6">
                Upload a PDF or CSV file containing notes, MCQs, coding questions, or question banks. 
                AI will automatically generate a complete draft test with questions, options, testcases, and more.
              </p>

              <form onSubmit={handleAIBuildTest} className="space-y-6">
                <div className="flex flex-col gap-3">
                  <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                    UPLOAD FILE (PDF OR CSV)
                  </label>
                  <div className="relative">
                    <input
                      type="file"
                      accept=".pdf,.csv"
                      onChange={(e) => {
                        setAiFile(e.target.files?.[0] || null)
                        setAiError(null)
                      }}
                      required
                      className="w-full bg-deep-bg/80 border border-panel-border px-4 py-3 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none file:mr-4 file:py-1 file:px-3 file:border-0 file:bg-neon-cyan/20 file:text-neon-cyan file:font-mono file:text-xs file:tracking-wider"
                    />
                  </div>
                  {aiFile && (
                    <div className="flex items-center gap-2 text-xs text-neon-cyan">
                      <FileText className="h-3 w-3" />
                      <span className="font-mono">{aiFile.name}</span>
                      <span className="text-muted-foreground">
                        ({(aiFile.size / 1024).toFixed(1)} KB)
                      </span>
                    </div>
                  )}
                </div>

                {aiError && (
                  <div className="border border-red-500/30 bg-red-500/10 px-4 py-3">
                    <p className="font-mono text-xs text-red-400">{aiError}</p>
                  </div>
                )}

                <div className="border-t border-panel-border pt-4">
                  <div className="flex items-start gap-2 mb-4">
                    <Shield className="h-4 w-4 text-neon-yellow shrink-0 mt-0.5" />
                    <div className="text-xs text-muted-foreground space-y-1">
                      <p>• Test will be created as <span className="text-neon-yellow">DRAFT</span> (not published)</p>
                      <p>• You can review and edit before publishing</p>
                      <p>• AI will generate questions from notes if needed</p>
                      <p>• Testcases will be auto-generated for coding problems</p>
                    </div>
                  </div>
                </div>

                <div className="flex justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => {
                      setShowAIModal(false)
                      setAiFile(null)
                      setAiError(null)
                    }}
                    className="px-5 py-2 font-mono text-[10px] tracking-widest text-muted-foreground hover:text-foreground transition-colors"
                  >
                    CANCEL
                  </button>
                  <button
                    type="submit"
                    disabled={aiGenerating || !aiFile}
                    className="flex items-center gap-2 bg-neon-cyan/90 hover:bg-neon-cyan px-6 py-2 font-mono text-[11px] font-bold tracking-widest text-deep-bg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {aiGenerating ? (
                      <>
                        <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        GENERATING...
                      </>
                    ) : (
                      <>
                        <Zap className="h-3.5 w-3.5" />
                        GENERATE TEST
                      </>
                    )}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

