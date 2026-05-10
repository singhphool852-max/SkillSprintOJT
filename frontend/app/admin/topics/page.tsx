"use client"

import { useEffect, useState } from "react"
import {
  Edit3,
  FolderOpen,
  Loader2,
  Plus,
  Shield,
  Trash2,
  X,
} from "lucide-react"
import { API_URL } from "@/lib/api-config"

const API = API_URL

interface Topic {
  id: string
  name: string
  slug: string
  description: string
  createdAt: string
}

export default function AdminTopicsPage() {
  const [topics, setTopics] = useState<Topic[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [creating, setCreating] = useState(false)

  // Form state
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")

  // Edit state
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editName, setEditName] = useState("")
  const [editDesc, setEditDesc] = useState("")
  const [editLoading, setEditLoading] = useState(false)

  // Delete
  const [deleteLoading, setDeleteLoading] = useState<string | null>(null)

  // Error
  const [error, setError] = useState("")

  async function fetchTopics() {
    try {
      const res = await fetch(`${API}/api/admin/topics`, { credentials: "include" })
      if (res.ok) setTopics(await res.json())
    } catch (e) {
      console.error("Failed to fetch topics:", e)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTopics()
  }, [])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!name.trim()) return
    setCreating(true)
    setError("")
    try {
      const res = await fetch(`${API}/api/admin/topics`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ name: name.trim(), description: description.trim() }),
      })
      if (res.ok) {
        setName("")
        setDescription("")
        setShowForm(false)
        await fetchTopics()
      } else {
        const data = await res.json()
        setError(data.error || "Failed to create topic")
      }
    } catch (e) {
      console.error("Create failed:", e)
      setError("Connection failed")
    } finally {
      setCreating(false)
    }
  }

  async function handleUpdate(topicId: string) {
    setEditLoading(true)
    try {
      const res = await fetch(`${API}/api/admin/topics/${topicId}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ name: editName.trim(), description: editDesc.trim() }),
      })
      if (res.ok) {
        setEditingId(null)
        await fetchTopics()
      }
    } catch (e) {
      console.error("Update failed:", e)
    } finally {
      setEditLoading(false)
    }
  }

  async function handleDelete(topicId: string) {
    if (!confirm("Delete this topic? Tests must be removed first.")) return
    setDeleteLoading(topicId)
    try {
      const res = await fetch(`${API}/api/admin/topics/${topicId}`, {
        method: "DELETE",
        credentials: "include",
      })
      if (res.ok) {
        await fetchTopics()
      } else {
        const data = await res.json()
        alert(data.error || "Failed to delete")
      }
    } catch (e) {
      console.error("Delete failed:", e)
    } finally {
      setDeleteLoading(null)
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
                TOPIC MANAGEMENT
              </span>
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
              ALL <span className="text-neon-pink text-glow-pink">TOPICS</span>
            </h1>
          </div>

          <button
            onClick={() => { setShowForm(!showForm); setError("") }}
            className="flex items-center gap-2 bg-neon-pink/90 hover:bg-neon-pink px-5 py-2.5 font-mono text-[11px] font-bold tracking-widest text-white transition-all"
          >
            <Plus className="h-3.5 w-3.5" />
            {showForm ? "CANCEL" : "CREATE TOPIC"}
          </button>
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
                NEW TOPIC
              </span>
            </div>

            <div className="grid gap-5 sm:grid-cols-2">
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  TOPIC NAME
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                  placeholder="e.g. DSA, DBMS, OS, CN"
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none placeholder:text-muted-foreground/40"
                />
              </div>
              <div className="flex flex-col gap-2">
                <label className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">
                  DESCRIPTION (OPTIONAL)
                </label>
                <input
                  type="text"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder="Brief description of this topic"
                  className="bg-deep-bg/80 border border-panel-border px-4 py-2.5 font-mono text-sm text-foreground focus:border-neon-pink/50 focus:outline-none placeholder:text-muted-foreground/40"
                />
              </div>
            </div>

            {error && (
              <div className="mt-4 flex items-center gap-2 border border-neon-pink/30 bg-neon-pink/5 px-4 py-2">
                <div className="h-1.5 w-1.5 rounded-full bg-neon-pink animate-pulse" />
                <span className="font-mono text-[11px] tracking-wider text-neon-pink">
                  {error.toUpperCase()}
                </span>
              </div>
            )}

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

        {/* Loading */}
        {loading && (
          <div className="flex flex-col gap-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-16 w-full animate-pulse border border-panel-border bg-panel-bg/20" />
            ))}
          </div>
        )}

        {/* Empty */}
        {!loading && topics.length === 0 && (
          <div className="flex flex-col items-center justify-center py-20 gap-4">
            <FolderOpen className="h-10 w-10 text-muted-foreground" />
            <span className="font-mono text-xs tracking-widest text-muted-foreground">
              NO TOPICS CREATED YET
            </span>
            <span className="font-mono text-[10px] text-muted-foreground/60">
              Create topics like DSA, DBMS, OS, CN to organize your tests
            </span>
          </div>
        )}

        {/* Topics list */}
        {!loading && topics.length > 0 && (
          <div className="border border-panel-border bg-panel-bg/40">
            {/* Header row */}
            <div className="grid grid-cols-[1fr_1fr_120px_100px] gap-4 px-5 py-3 border-b border-panel-border">
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">NAME</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">DESCRIPTION</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase">CREATED</span>
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground uppercase text-right">ACTIONS</span>
            </div>

            {topics.map((topic) => (
              <div
                key={topic.id}
                className="grid grid-cols-[1fr_1fr_120px_100px] gap-4 items-center px-5 py-3 border-b border-panel-border/50 last:border-0 hover:bg-neon-pink/5 transition-colors"
              >
                {editingId === topic.id ? (
                  // Inline edit mode
                  <>
                    <input
                      type="text"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                      className="bg-deep-bg/80 border border-neon-pink/40 px-3 py-1.5 font-mono text-sm text-foreground focus:outline-none"
                    />
                    <input
                      type="text"
                      value={editDesc}
                      onChange={(e) => setEditDesc(e.target.value)}
                      className="bg-deep-bg/80 border border-panel-border px-3 py-1.5 font-mono text-sm text-foreground focus:outline-none"
                    />
                    <span className="font-mono text-[10px] text-muted-foreground">
                      {new Date(topic.createdAt).toLocaleDateString()}
                    </span>
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={() => handleUpdate(topic.id)}
                        disabled={editLoading}
                        className="px-2 py-1 border border-neon-cyan/40 text-neon-cyan font-mono text-[9px] tracking-widest hover:bg-neon-cyan/10 transition-colors disabled:opacity-50"
                      >
                        {editLoading ? <Loader2 className="h-3 w-3 animate-spin" /> : "SAVE"}
                      </button>
                      <button
                        onClick={() => setEditingId(null)}
                        className="px-2 py-1 text-muted-foreground hover:text-foreground transition-colors"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </div>
                  </>
                ) : (
                  // Display mode
                  <>
                    <div className="flex items-center gap-3 min-w-0">
                      <FolderOpen className="h-4 w-4 text-neon-pink shrink-0" />
                      <span className="font-mono text-sm font-bold tracking-wider text-foreground truncate">
                        {topic.name.toUpperCase()}
                      </span>
                      <span className="font-mono text-[9px] text-muted-foreground/60 shrink-0">
                        /{topic.slug}
                      </span>
                    </div>
                    <span className="font-mono text-[10px] text-muted-foreground truncate">
                      {topic.description || "—"}
                    </span>
                    <span className="font-mono text-[10px] text-muted-foreground">
                      {new Date(topic.createdAt).toLocaleDateString()}
                    </span>
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={() => { setEditingId(topic.id); setEditName(topic.name); setEditDesc(topic.description) }}
                        className="p-1.5 text-muted-foreground hover:text-neon-cyan transition-colors"
                        title="Edit"
                      >
                        <Edit3 className="h-3.5 w-3.5" />
                      </button>
                      <button
                        onClick={() => handleDelete(topic.id)}
                        disabled={deleteLoading === topic.id}
                        className="p-1.5 text-muted-foreground hover:text-neon-pink transition-colors disabled:opacity-50"
                        title="Delete"
                      >
                        {deleteLoading === topic.id ? (
                          <Loader2 className="h-3.5 w-3.5 animate-spin" />
                        ) : (
                          <Trash2 className="h-3.5 w-3.5" />
                        )}
                      </button>
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Footer */}
        {topics.length > 0 && (
          <div className="mt-6 flex items-center gap-4">
            <div className="h-px flex-1 bg-panel-border" />
            <div className="flex items-center gap-2">
              <FolderOpen className="h-3 w-3 text-muted-foreground" />
              <span className="font-mono text-[10px] tracking-widest text-muted-foreground uppercase">
                {topics.length} TOPIC{topics.length !== 1 ? "S" : ""}
              </span>
            </div>
            <div className="h-px flex-1 bg-panel-border" />
          </div>
        )}
      </div>
    </div>
  )
}

