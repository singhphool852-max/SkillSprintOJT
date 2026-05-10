import { FileUp, FileText, Zap, Activity, AlertCircle } from "lucide-react"
import { useState, useRef } from "react"
import { useRouter } from "next/navigation"
import { API_URL } from "@/lib/api-config"

export function NotesUpload() {
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)
  
  const [topic, setTopic] = useState("DBMS")
  const [difficulty, setDifficulty] = useState("Medium")
  const [count, setCount] = useState(5)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files[0]) {
      setSelectedFile(e.target.files[0])
      setError(null)
    }
  }

  const handleUpload = async () => {
    if (loading) return // Prevent double-clicks
    if (!selectedFile) {
      setError("SELECT A DATA FILE FIRST")
      return
    }

    setLoading(true)
    setError(null)

    const formData = new FormData()
    formData.append("file", selectedFile)
    formData.append("topic", topic)
    formData.append("difficulty", difficulty)
    formData.append("count", count.toString())

    const UPLOAD_URL = `${API_URL}/api/train/upload-notes`

    console.log("[NOTES_UPLOAD] Init Sync Request:")
    console.log(" - API_URL:", API_URL)
    console.log(" - URL:", UPLOAD_URL)
    console.log(" - File:", selectedFile.name, `(${selectedFile.size} bytes)`)
    console.log(" - Metadata:", { topic, difficulty, count })

    try {
      const res = await fetch(UPLOAD_URL, {
        method: "POST",
        body: formData,
        credentials: "include"
      })

      const raw = await res.text()
      let parsed: any = null
      try {
        parsed = raw ? JSON.parse(raw) : null
      } catch (parseErr) {
        console.error("[NOTES_UPLOAD] Response JSON parse failed:", parseErr)
        console.error("[NOTES_UPLOAD] Raw response text (non-JSON):", raw)
      }
      console.log("[NOTES_UPLOAD] Response status:", res.status)
      console.log("[NOTES_UPLOAD] Raw response text:", raw)

      if (!res.ok) {
        console.error("[NOTES_UPLOAD] Backend Error:", parsed ?? raw)
        const backendMessage =
          parsed?.error ||
          (raw && raw.trim().length > 0 ? raw : "Backend returned an empty error response")
        throw new Error(backendMessage || `HTTP_ERROR: ${res.status}`)
      }

      const data = parsed ?? {}
      console.log("[NOTES_UPLOAD] Success Response:", data)
      router.push(`/train/play/${data.session_id}?topic=${encodeURIComponent(topic)}&mode=AI_SYNTH_MODE&difficulty=${encodeURIComponent(difficulty)}&count=${data.count}`)
    } catch (err: any) {
      console.error("[NOTES_UPLOAD] Fatal Error:", err)
      
      // Distinguish between network errors (TypeError) and response errors
      if (err instanceof TypeError && err.message === "Failed to fetch") {
        setError("Backend unreachable. Please check your network and server status.")
      } else {
        setError(err.message || "SYNC_FAILURE: UNKNOWN_ERROR")
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative border border-panel-border bg-panel-bg/60 p-8 lg:p-12 overflow-hidden group">
      <div className="absolute top-0 right-0 p-4 opacity-5 group-hover:opacity-10 transition-opacity">
        <FileText className="h-40 w-40 -mr-16 -mt-16" />
      </div>

      <div className="relative z-10 grid gap-12 lg:grid-cols-2 lg:items-center">
        <div>
          <div className="flex items-center gap-3 mb-6">
            <div className="h-8 w-8 flex items-center justify-center border border-neon-pink text-neon-pink bg-neon-pink/5">
              <Zap className="h-4 w-4" />
            </div>
            <span className="font-mono text-[10px] tracking-[0.3em] text-neon-pink uppercase">
              COGNITIVE SYNC
            </span>
          </div>

          <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl pr-10">
            TRAIN FROM YOUR <span className="text-neon-pink text-glow-pink">NOTES</span>
          </h2>
          <p className="mt-4 text-sm leading-relaxed text-muted-foreground lg:text-base max-w-lg">
            Upload technical docs or PDFs. Our AI will extract context, summarize key concepts, and generate a contextual training session.
          </p>

          <div className="mt-8 grid grid-cols-2 gap-4">
             <div>
                <label className="block font-mono text-[9px] text-muted-foreground tracking-widest uppercase mb-2">DOMAIN CONTEXT</label>
                <select 
                  className="w-full bg-deep-bg/50 border border-panel-border px-3 py-2 font-mono text-[10px] text-foreground focus:border-neon-pink focus:outline-none appearance-none"
                  value={topic}
                  onChange={(e) => setTopic(e.target.value)}
                >
                  <option value="DBMS">DBMS</option>
                  <option value="DSA">DSA</option>
                  <option value="OS">OPERATING SYSTEMS</option>
                  <option value="JS">JAVASCRIPT</option>
                  <option value="SYSTEM DESIGN">SYSTEM DESIGN</option>
                </select>
             </div>
             <div>
                <label className="block font-mono text-[9px] text-muted-foreground tracking-widest uppercase mb-2">DIFFICULTY</label>
                <select 
                  className="w-full bg-deep-bg/50 border border-panel-border px-3 py-2 font-mono text-[10px] text-foreground focus:border-neon-pink focus:outline-none appearance-none"
                  value={difficulty}
                  onChange={(e) => setDifficulty(e.target.value)}
                >
                  <option value="Easy">EASY</option>
                  <option value="Medium">MEDIUM</option>
                  <option value="Hard">HARD</option>
                </select>
             </div>
          </div>
        </div>

        <div className="relative">
          <input 
            type="file" 
            ref={fileInputRef}
            className="hidden" 
            accept=".txt,.pdf"
            onChange={handleFileChange}
          />
          
          <div 
            onClick={() => !loading && fileInputRef.current?.click()}
            className={`flex flex-col items-center justify-center border-2 border-dashed p-10 text-center transition-all cursor-pointer ${
              selectedFile ? 'border-neon-pink bg-neon-pink/10 shadow-[0_0_15px_rgba(255,0,255,0.1)]' : 'border-panel-border bg-panel-bg/40 hover:bg-panel-bg/60 hover:border-neon-pink/40'
            } ${loading ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            <div className={`mb-4 h-16 w-16 flex items-center justify-center rounded-full bg-neon-pink/5 text-neon-pink transition-transform ${!loading && 'group-hover/upload:-translate-y-1'}`}>
              {loading ? <Activity className="h-8 w-8 animate-spin" /> : <FileUp className="h-8 w-8" />}
            </div>
            
            <p className="font-mono text-xs font-bold tracking-widest text-foreground uppercase mb-1">
              {loading ? "SYNCHRONIZING..." : selectedFile ? selectedFile.name : "DROP YOUR DATA HERE"}
            </p>
            <p className="text-[10px] text-muted-foreground font-mono uppercase">
              {loading ? "CALIBRATING NEURAL WEIGHTS" : "PDF or TXT ONLY // 10MB LIMIT"}
            </p>

            {!loading && selectedFile && (
              <button 
                disabled={loading}
                onClick={(e) => { e.stopPropagation(); handleUpload(); }}
                className="mt-6 px-6 py-2 border border-neon-pink text-neon-pink font-mono text-[10px] tracking-widest hover:bg-neon-pink hover:text-white transition-all animate-pulse-glow disabled:opacity-50 disabled:cursor-not-allowed"
              >
                INITIALIZE SYNC
              </button>
            )}
          </div>

          {error && (
            <div className="mt-4 flex items-center gap-2 border border-neon-pink/20 bg-neon-pink/5 p-3 font-mono text-[9px] text-neon-pink uppercase tracking-wider">
              <AlertCircle className="h-3 w-3 shrink-0" />
              {error}
            </div>
          )}
          
          <div className="absolute -inset-1 border border-neon-pink/0 group-hover:border-neon-pink/10 pointer-events-none transition-all duration-500" />
        </div>
      </div>
    </div>
  )
}


