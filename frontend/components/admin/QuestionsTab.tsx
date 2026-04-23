"use client"
import { useState, useEffect } from "react"
import { HelpCircle, Plus, Trash2, CheckCircle2 } from "lucide-react"

const API = "http://localhost:8080/api/admin"

interface Arena { id: string; title: string }
interface Option { id: string; text: string; isCorrect: boolean }
interface Question { id: string; prompt: string; maxScore: number; options: Option[] }

export function QuestionsTab() {
  const [arenas, setArenas] = useState<Arena[]>([])
  const [selectedArena, setSelectedArena] = useState("")
  const [questions, setQuestions] = useState<Question[]>([])
  const [prompt, setPrompt] = useState("")
  const [maxScore, setMaxScore] = useState(10)
  const [explanation, setExplanation] = useState("")
  const [options, setOptions] = useState([
    { text: "", isCorrect: true },
    { text: "", isCorrect: false },
    { text: "", isCorrect: false },
    { text: "", isCorrect: false },
  ])
  const [error, setError] = useState("")
  const [success, setSuccess] = useState("")
  const [loading, setLoading] = useState(false)

  const fetchArenas = async () => {
    const res = await fetch(`${API}/arenas`, { credentials: "include" })
    if (res.ok) {
      const data = await res.json()
      setArenas(data)
      if (data.length > 0 && !selectedArena) setSelectedArena(data[0].id)
    }
  }

  const fetchQuestions = async (arenaId: string) => {
    if (!arenaId) return
    const res = await fetch(`${API}/arenas/${arenaId}/questions`, { credentials: "include" })
    if (res.ok) setQuestions(await res.json())
  }

  useEffect(() => { fetchArenas() }, [])
  useEffect(() => { if (selectedArena) fetchQuestions(selectedArena) }, [selectedArena])

  const setOptionText = (idx: number, text: string) => {
    const copy = [...options]; copy[idx].text = text; setOptions(copy)
  }
  const setCorrectOption = (idx: number) => {
    setOptions(options.map((o, i) => ({ ...o, isCorrect: i === idx })))
  }

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(""); setSuccess(""); setLoading(true)
    const filledOptions = options.filter(o => o.text.trim())
    if (filledOptions.length < 2) { setError("At least 2 options required"); setLoading(false); return }
    if (!filledOptions.some(o => o.isCorrect)) { setError("Mark one option as correct"); setLoading(false); return }
    try {
      const res = await fetch(`${API}/arenas/${selectedArena}/questions`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt, maxScore, explanation, options: filledOptions }),
        credentials: "include",
      })
      const data = await res.json()
      if (!res.ok) { setError(data.error); setLoading(false); return }
      setSuccess("Question added")
      setPrompt(""); setExplanation(""); setMaxScore(10)
      setOptions([{ text: "", isCorrect: true }, { text: "", isCorrect: false }, { text: "", isCorrect: false }, { text: "", isCorrect: false }])
      fetchQuestions(selectedArena)
    } catch { setError("Failed to add question") }
    setLoading(false)
  }

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this question?")) return
    await fetch(`${API}/questions/${id}`, { method: "DELETE", credentials: "include" })
    fetchQuestions(selectedArena)
  }

  return (
    <div className="space-y-6">
      {/* Arena Selector */}
      <div className="border border-panel-border bg-deep-bg/80 p-4">
        <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground mb-2 block">SELECT ARENA</label>
        <select value={selectedArena} onChange={e => setSelectedArena(e.target.value)}
          className="w-full border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none">
          {arenas.map(a => <option key={a.id} value={a.id}>{a.title}</option>)}
        </select>
      </div>

      {/* Add Question Form */}
      {selectedArena && (
        <div className="border border-panel-border bg-deep-bg/80 p-6">
          <div className="flex items-center gap-2 mb-4">
            <Plus className="h-4 w-4 text-neon-amber" />
            <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-amber">ADD QUESTION</span>
          </div>
          <form onSubmit={handleAdd} className="space-y-4">
            <div className="flex flex-col gap-1">
              <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">QUESTION PROMPT</label>
              <textarea value={prompt} onChange={e => setPrompt(e.target.value)} required rows={2} placeholder="What is the time complexity of..."
                className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none resize-none" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              {options.map((opt, i) => (
                <div key={i} className="flex items-center gap-2">
                  <button type="button" onClick={() => setCorrectOption(i)}
                    className={`flex-shrink-0 h-6 w-6 flex items-center justify-center border transition-colors ${opt.isCorrect ? "border-green-400 bg-green-400/20 text-green-400" : "border-panel-border text-muted-foreground/40 hover:border-green-400/50"}`}>
                    <CheckCircle2 className="h-3.5 w-3.5" />
                  </button>
                  <input type="text" value={opt.text} onChange={e => setOptionText(i, e.target.value)} placeholder={`Option ${String.fromCharCode(65 + i)}`}
                    className="flex-1 border border-panel-border bg-deep-bg/60 px-3 py-2 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
                </div>
              ))}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1">
                <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">MAX SCORE</label>
                <input type="number" value={maxScore} onChange={e => setMaxScore(parseInt(e.target.value) || 10)} min={1}
                  className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground focus:border-neon-cyan/50 focus:outline-none" />
              </div>
              <div className="flex flex-col gap-1">
                <label className="font-mono text-[10px] tracking-[0.15em] text-muted-foreground">EXPLANATION (OPTIONAL)</label>
                <input type="text" value={explanation} onChange={e => setExplanation(e.target.value)} placeholder="Why this answer is correct..."
                  className="border border-panel-border bg-deep-bg/60 px-3 py-2.5 font-mono text-sm text-foreground placeholder:text-muted-foreground/40 focus:border-neon-cyan/50 focus:outline-none" />
              </div>
            </div>
            <div className="flex items-center gap-4">
              <button type="submit" disabled={loading}
                className="bg-neon-amber/90 hover:bg-neon-amber px-6 py-2.5 font-mono text-xs font-bold tracking-widest text-deep-bg transition-all disabled:opacity-50">
                {loading ? "ADDING..." : "ADD QUESTION"}
              </button>
              {error && <span className="font-mono text-[11px] text-neon-pink">{error.toUpperCase()}</span>}
              {success && <span className="font-mono text-[11px] text-neon-cyan">{success.toUpperCase()}</span>}
            </div>
          </form>
        </div>
      )}

      {/* Question List */}
      <div className="border border-panel-border bg-deep-bg/80">
        <div className="flex items-center gap-2 border-b border-panel-border px-6 py-3 bg-neon-amber/5">
          <HelpCircle className="h-4 w-4 text-neon-amber" />
          <span className="font-mono text-xs font-bold tracking-[0.2em] text-neon-amber">QUESTIONS ({questions.length})</span>
        </div>
        <div className="divide-y divide-panel-border/50">
          {questions.length === 0 ? (
            <div className="px-6 py-8 text-center font-mono text-xs text-muted-foreground">
              {selectedArena ? "NO QUESTIONS YET — ADD ONE ABOVE" : "SELECT AN ARENA FIRST"}
            </div>
          ) : questions.map((q, idx) => (
            <div key={q.id} className="px-6 py-4 hover:bg-neon-amber/5 transition-colors">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <span className="font-mono text-xs text-neon-cyan mr-2">Q{idx + 1}.</span>
                  <span className="font-mono text-sm text-foreground">{q.prompt}</span>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {q.options?.map(opt => (
                      <span key={opt.id} className={`font-mono text-[11px] px-2 py-1 border ${opt.isCorrect ? "border-green-400/50 bg-green-400/10 text-green-400" : "border-panel-border text-muted-foreground"}`}>
                        {opt.text}
                      </span>
                    ))}
                  </div>
                  <span className="font-mono text-[10px] text-muted-foreground mt-1 block">Score: {q.maxScore}</span>
                </div>
                <button onClick={() => handleDelete(q.id)} className="p-2 text-muted-foreground hover:text-neon-pink transition-colors ml-4">
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
