"use client"
import { useState, useEffect } from "react"
import { Users, UserPlus, Trash2, Mail, Lock, Crosshair } from "lucide-react"

const API = "http://localhost:8080/api/admin"

interface Student {
  id: string; email: string; username: string; createdAt: string
}

export function StudentsTab() {
  const [students, setStudents] = useState<Student[]>([])
  const [email, setEmail] = useState("")
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [success, setSuccess] = useState("")
  const [loading, setLoading] = useState(false)

  const fetchStudents = async () => {
    const res = await fetch(`${API}/students`, { credentials: "include" })
    if (res.ok) setStudents(await res.json())
  }

  useEffect(() => { fetchStudents() }, [])

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(""); setSuccess(""); setLoading(true)
    try {
      const res = await fetch(`${API}/students`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, username, password }),
        credentials: "include",
      })
      const data = await res.json()
      if (!res.ok) { setError(data.error); setLoading(false); return }
      setSuccess(`Student "${username}" created`)
      setEmail(""); setUsername(""); setPassword("")
      fetchStudents()
    } catch { setError("Failed to create student") }
    setLoading(false)
  }

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this student?")) return
    await fetch(`${API}/students/${id}`, { method: "DELETE", credentials: "include" })
    fetchStudents()
  }

  return (
    <div className="space-y-6">
      {/* Add Student Form */}
      <div className="border border-panel-border bg-deep-bg/80 p-6">
        <div className="flex items-center gap-2 mb-4">
          <UserPlus className="h-4 w-4 text-neon-amber" />
          <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-amber">ADD STUDENT</span>
        </div>
        <form onSubmit={handleAdd} className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="flex flex-col gap-1">
            <label className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.15em] text-muted-foreground">
              <Mail className="h-3 w-3 text-neon-cyan/60" /> EMAIL
            </label>
            <input type="email" value={email} onChange={e => setEmail(e.target.value)} required placeholder="student@example.com"
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.15em] text-muted-foreground">
              <Crosshair className="h-3 w-3 text-neon-cyan/60" /> USERNAME
            </label>
            <input type="text" value={username} onChange={e => setUsername(e.target.value)} required placeholder="john_doe"
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
          </div>
          <div className="flex flex-col gap-1">
            <label className="flex items-center gap-1.5 font-mono text-[10px] tracking-[0.15em] text-muted-foreground">
              <Lock className="h-3 w-3 text-neon-cyan/60" /> PASSWORD
            </label>
            <input type="password" value={password} onChange={e => setPassword(e.target.value)} required placeholder="••••••" minLength={6}
              className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
          </div>
          <div className="md:col-span-3 flex items-center gap-4">
            <button type="submit" disabled={loading}
              className="bg-neon-amber/90 hover:bg-neon-amber px-6 py-2.5 font-mono text-xs font-bold tracking-widest text-deep-bg transition-all disabled:opacity-50">
              {loading ? "CREATING..." : "ADD STUDENT"}
            </button>
            {error && <span className="font-mono text-[11px] text-neon-pink">{error.toUpperCase()}</span>}
            {success && <span className="font-mono text-[11px] text-neon-cyan">{success.toUpperCase()}</span>}
          </div>
        </form>
      </div>

      {/* Student List */}
      <div className="border border-panel-border bg-deep-bg/80">
        <div className="flex items-center gap-2 border-b border-panel-border px-6 py-3 bg-neon-cyan/5">
          <Users className="h-4 w-4 text-neon-cyan" />
          <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-cyan">ENROLLED STUDENTS ({students.length})</span>
        </div>
        <div className="divide-y divide-panel-border/50">
          {students.length === 0 ? (
            <div className="px-6 py-8 text-center font-mono text-xs text-muted-foreground">NO STUDENTS FOUND</div>
          ) : students.map(s => (
            <div key={s.id} className="flex items-center justify-between px-6 py-3 hover:bg-neon-cyan/5 transition-colors">
              <div className="flex items-center gap-4">
                <div className="flex h-8 w-8 items-center justify-center border border-neon-cyan/30 bg-neon-cyan/10">
                  <span className="font-mono text-xs font-bold text-neon-cyan">{s.username[0]?.toUpperCase()}</span>
                </div>
                <div>
                  <span className="block font-mono text-sm font-bold text-foreground">{s.username}</span>
                  <span className="font-mono text-[10px] text-muted-foreground">{s.email}</span>
                </div>
              </div>
              <button onClick={() => handleDelete(s.id)} className="p-2 text-muted-foreground hover:text-neon-pink transition-colors">
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
