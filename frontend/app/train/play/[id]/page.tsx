// Server component — no "use client" here.
// Only TrainPlayContent (below) is a client component.
import { Suspense } from "react"
import { Loader2 } from "lucide-react"
import TrainPlayContent from "./TrainPlayContent"

// Shown while TrainPlayContent (which calls useSearchParams) is suspending.
function LoadingScreen() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4 bg-deep-bg">
      <div className="relative">
        <Loader2 className="h-12 w-12 text-neon-cyan animate-spin" />
        <div className="absolute inset-0 h-12 w-12 border-b-2 border-neon-cyan/30 rounded-full animate-pulse" />
      </div>
      <div className="flex flex-col items-center gap-1">
        <span className="font-mono text-[10px] tracking-[0.4em] text-neon-cyan uppercase">
          Initializing Neural Session
        </span>
        <span className="font-mono text-[8px] text-muted-foreground uppercase opacity-40 italic">
          Decrypting logical parameters...
        </span>
      </div>
    </div>
  )
}

export default async function TrainingPlayPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  return (
    <Suspense fallback={<LoadingScreen />}>
      <TrainPlayContent id={id} />
    </Suspense>
  )
}
