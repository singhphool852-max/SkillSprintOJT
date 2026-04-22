"use client"
import { useState, useEffect } from "react"
import { BarChart3, Plus, Trash2, Clock, Zap } from "lucide-react"

const API = "http://localhost:8080/api/admin"

interface Category { id: string; name: string; slug: string }
interface Arena { id: string; title: string; slug: string; difficulty: string; status: string; durationSeconds: number; description: string; category: Category }

export function ArenasTab() {
  const [arenas, setArenas] = useState<Arena[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [title, setTitle] = useState("")
  const [difficulty, setDifficulty] = useState("Medium")
  const [categoryId, setCategoryId] = useState("")
  const [duration, setDuration] = useState(10)
  const [description, setDescription] = useState("")
  const [error, setError] = useState("")
  const [success, setSuccess] = useState("")
  const [loading, setLoading] = useState(false)

  const fetchArenas = async () => {
    const res = await fetch(`${API}/arenas`, { credentials: "include" })
    if (res.ok) setArenas(await res.json())
  }
  const fetchCategories = async () => {
    const res = await fetch(`${API}/categories`, { credentials: "include" })
    if (res.ok) {
      const cats = await res.json()
      setCategories(cats)
      if (cats.length > 0 && !categoryId) setCategoryId(cats[0].id)
    }
  }

  useEffect(() => { fetchArenas(); fetchCategories() }, [])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(""); setSuccess(""); setLoading(true)
    try {
      const res = await fetch(`${API}/arenas`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title, difficulty, categoryId, durationMinutes: duration, description }),
        credentials: "include",
      })
      const data = await res.json()
      if (!res.ok) { setError(data.error); setLoading(false); return }
      setSuccess(`Arena "${title}" created`)
      setTitle(""); setDescription(""); setDuration(10)
      fetchArenas()
    } catch { setError("Failed to create arena") }
    setLoading(false)
  }

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this arena and all its questions?")) return
    await fetch(`${API}/arenas/${id}`, { method: "DELETE", credentials: "include" })
    fetchArenas()
  }

  return (
    <div className="space-y-6">
      {/* Create Arena Form */}
      <div className="border border-panel-border bg-deep-bg/80 p-6">
        <div className="flex items-center gap-2 mb-4">
          <Plus className="h-4 w-4 text-neon-amber" />
          <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-amber">CREATE ARENA</span>
        </div>
        <form onSubmit={handleCreate} className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="flex flex-col gap-1">
            <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">TITLE</label>
            <input type="text" value={title} onChange={e => setTitle(e.target.value)} required placeholder="DSA Speed Relay"
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">DIFFICULTY</label>
            <select value={difficulty} onChange={e => setDifficulty(e.target.value)}
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none">
              <option value="Easy">Easy</option>
              <option value="Medium">Medium</option>
              <option value="Hard">Hard</option>
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">CATEGORY</label>
            <select value={categoryId} onChange={e => setCategoryId(e.target.value)}
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none">
              {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.15em] text-muted-foreground">
              <Clock className="h-3 w-3 text-neon-cyan/60" /> TIME LIMIT (MINUTES)
            </label>
            <input type="number" value={duration} onChange={e => setDuration(parseInt(e.target.value) || 1)} min={1} required
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none" />
          </div>
          <div className="md:col-span-2 flex flex-col gap-1">
            <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">DESCRIPTION (OPTIONAL)</label>
            <input type="text" value={description} onChange={e => setDescription(e.target.value)} placeholder="Arena description..."
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
          </div>
          <div className="md:col-span-2 flex items-center gap-4">
            <button type="submit" disabled={loading}
              className="bg-neon-amber/90 hover:bg-neon-amber px-6 py-2.5 font-mono text-xs font-bold tracking-widest text-deep-bg transition-all disabled:opacity-50">
              {loading ? "CREATING..." : "CREATE ARENA"}
            </button>
            {error && <span className="font-mono text-[11px] text-neon-pink">{error.toUpperCase()}</span>}
            {success && <span className="font-mono text-[11px] text-neon-cyan">{success.toUpperCase()}</span>}
          </div>
        </form>
      </div>

      {/* Arena List */}
      <div className="border border-panel-border bg-deep-bg/80">
        <div className="flex items-center gap-2 border-b border-panel-border px-6 py-3 bg-neon-pink/5">
          <BarChart3 className="h-4 w-4 text-neon-pink" />
          <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-pink">ARENAS ({arenas.length})</span>
        </div>
        <div className="divide-y divide-panel-border/50">
          {arenas.length === 0 ? (
            <div className="px-6 py-8 text-center font-mono text-xs text-muted-foreground">NO ARENAS FOUND</div>
          ) : arenas.map(a => (
            <div key={a.id} className="flex items-center justify-between px-6 py-4 hover:bg-neon-pink/5 transition-colors">
              <div className="flex items-center gap-4">
                <div className="flex h-10 w-10 items-center justify-center border border-neon-pink/30 bg-neon-pink/10">
                  <Zap className="h-5 w-5 text-neon-pink" />
                </div>
                <div>
                  <span className="block font-mono text-sm font-bold text-foreground">{a.title}</span>
                  <div className="flex items-center gap-3 mt-0.5">
                    <span className="font-mono text-[10px] text-muted-foreground">{a.category?.name || "—"}</span>
                    <span className="font-mono text-[10px] text-neon-amber">{a.difficulty}</span>
                    <span className="flex items-center gap-1 font-mono text-[10px] text-neon-cyan">
                      <Clock className="h-3 w-3" /> {Math.floor(a.durationSeconds / 60)}m
                    </span>
                    <span className={`font-mono text-[10px] ${a.status === "live" ? "text-green-400" : "text-muted-foreground"}`}>
                      {a.status?.toUpperCase()}
                    </span>
                  </div>
                </div>
              </div>
              <button onClick={() => handleDelete(a.id)} className="p-2 text-muted-foreground hover:text-neon-pink transition-colors">
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
