"use client"

import { useEffect, useState, use } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { ProtectedRoute } from "@/components/ProtectedRoute"
import { TrainingSolver } from "@/components/train/training-solver"
import { Loader2, ShieldAlert, ArrowLeft, RefreshCcw, Info, ShieldCheck } from "lucide-react"
import { API_URL } from "@/lib/api-config"

export default function TrainingPlayPage({ params: paramsPromise }: { params: Promise<{ id: string }> }) {
  const params = use(paramsPromise)
  const router = useRouter()
  const searchParams = useSearchParams()
  
  const { id } = params
  const topic = searchParams.get("topic") || "General"
  const mode = searchParams.get("mode") || "Standard"
  const difficulty = searchParams.get("difficulty") || "Medium"
  const count = parseInt(searchParams.get("count") || "10")
  
  const [questions, setQuestions] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [offlineStatus, setOfflineStatus] = useState<string | null>(null)

  const fetchQuestions = async () => {
    if (!id) return;
    setLoading(true);
    setOfflineStatus(null);
    
    try {
      const isUUID = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/.test(id);
      let apiData: any;

      if (isUUID) {
        // PATH A: Retrieve existing session by UUID
        const res = await fetch(`${API_URL}/api/train/session/${id}`, {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include"
        });

        if (res.status === 404) {
          setOfflineStatus("Session not found. Please regenerate.");
          setQuestions([]);
          setLoading(false);
          return;
        }

        if (!res.ok) {
          const errData = await res.json().catch(() => ({}));
          setOfflineStatus(errData.error || `Failed to fetch session (${res.status})`);
          setQuestions([]);
          setLoading(false);
          return;
        }

        apiData = await res.json();
      } else {
        // PATH B: Create new session from topic name
        const res = await fetch(`${API_URL}/api/train/session`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            topic: id.toLowerCase(),
            count,
            difficulty: difficulty.toLowerCase()
          }),
          credentials: "include"
        });

        if (!res.ok) {
          const errData = await res.json().catch(() => ({}));
          setOfflineStatus(errData.error || `Failed to create session (${res.status})`);
          setQuestions([]);
          setLoading(false);
          return;
        }

        apiData = await res.json();
      }

      console.log("Fetched session:", apiData);

      const sessionQuestions = apiData.questions || [];
      const sessionId = apiData.session_id || apiData.sessionId;

      if (sessionQuestions.length === 0) {
        setOfflineStatus("No questions available");
        setQuestions([]);
        setLoading(false);
        return;
      }

      const mapped = sessionQuestions.map((q: any) => {
        let normalizedOptions = q.options;
        if (typeof q.options === 'string') {
          try { normalizedOptions = JSON.parse(q.options); } catch { normalizedOptions = []; }
        }

        if (q.type === "mcq" && Array.isArray(normalizedOptions)) {
           normalizedOptions = normalizedOptions.map((optText: string, idx: number) => {
             if (typeof optText === 'object' && optText !== null) return optText;
             return { id: `OPT_${q.id}_${idx}`, text: optText };
           });
        }

        return {
          id: String(q.id),
          prompt: q.prompt || "No prompt provided",
          type: q.type,
          options: normalizedOptions,
          explanation: q.explanation,
          maxScore: 10,
          _answer: q.answer,
          _source: q.source,
          _sessionId: sessionId
        };
      });

      setQuestions(mapped);
      setLoading(false);

    } catch (err: any) {
      console.error("[TrainPlay] Error:", err?.message || err);
      setOfflineStatus(err?.message || "Failed to load session");
      setQuestions([]);
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchQuestions()
  }, [id])

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen gap-4 bg-deep-bg">
        <div className="relative">
          <Loader2 className="h-12 w-12 text-neon-cyan animate-spin" />
          <div className="absolute inset-0 h-12 w-12 border-b-2 border-neon-cyan/30 rounded-full animate-pulse" />
        </div>
        <div className="flex flex-col items-center gap-1">
          <span className="font-mono text-[10px] tracking-[0.4em] text-neon-cyan uppercase">Initializing Neural Session</span>
          <span className="font-mono text-[8px] text-muted-foreground uppercase opacity-40 italic">Decrypting logical parameters...</span>
        </div>
      </div>
    )
  }

  if (questions.length === 0 && !loading) {
     return (
       <div className="flex flex-col items-center justify-center min-h-screen gap-6 bg-deep-bg p-4 relative overflow-hidden">
         <ShieldAlert className="h-12 w-12 text-neon-pink" />
         <div className="text-center">
            <h2 className="text-xl font-bold font-mono text-foreground uppercase tracking-widest">CRITICAL ANOMALY</h2>
            <p className="text-xs text-muted-foreground font-mono mt-2 uppercase">No data found in vault and fallback datasets are compromised.</p>
         </div>
         <button onClick={() => router.push("/train")} className="px-6 py-3 border border-panel-border bg-panel-bg/50 text-xs font-mono font-bold tracking-widest text-muted-foreground uppercase hover:bg-white/5 transition-all flex items-center justify-center gap-2">
           <ArrowLeft className="h-3 w-3" /> Return to Hub
         </button>
       </div>
     )
  }

  return (
    <ProtectedRoute>
      <div className="relative">
         {offlineStatus && (
           <div className="fixed top-20 right-8 z-50 animate-pulse">
             <div className="flex items-center gap-2 px-3 py-2 bg-neon-yellow/10 border border-neon-yellow/40 backdrop-blur">
               <ShieldCheck className="h-4 w-4 text-neon-yellow" />
               <span className="font-mono text-[9px] text-neon-yellow uppercase tracking-widest font-black max-w-[200px] truncate">{offlineStatus}</span>
             </div>
           </div>
         )}
         <TrainingSolver 
           initialQuestions={questions}
           topic={topic + (offlineStatus ? " (Offline)" : "")}
           mode={mode}
           difficulty={difficulty}
           count={offlineStatus ? questions.length : count}
           arenaId={id}
         />
      </div>
    </ProtectedRoute>
  )
}
