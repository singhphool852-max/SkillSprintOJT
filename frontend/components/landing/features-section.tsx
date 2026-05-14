"use client"

import { Activity, Brain, Database, Swords, Target, Trophy, Zap } from "lucide-react"

const features = [
  {
    icon: Brain,
    title: "ADAPTIVE TRAINING",
    description: "AI-driven sessions that evolve with your performance. Targeting weak topics to ensure technical mastery.",
    color: "cyan" as const,
    tag: "CORE",
  },
  {
    icon: Zap,
    title: "RECOVERY MODE",
    description: "Systematically re-master previously failed questions. Turn technical blind spots into tactical advantages.",
    color: "pink" as const,
    tag: "RECOVERY",
  },
  {
    icon: Trophy,
    title: "RANKED LADDER",
    description: "Climb the global leaderboard from Rookie to Apex tier. Gain prestige as you conquer technical challenges.",
    color: "yellow" as const,
    tag: "COMPETE",
  },
  {
    icon: Activity,
    title: "NEURAL SYNTHESIS",
    description: "Sync your personal notes or PDFs. Our AI extracts core concepts and generates custom training drills.",
    color: "cyan" as const,
    tag: "SYNC",
  },
  {
    icon: Target,
    title: "FULL ANALYTICS",
    description: "Track accuracy, reaction speed, and mastery trends. Data-driven insights to optimize your learning path.",
    color: "pink" as const,
    tag: "DATA",
  },
  {
    icon: Database,
    title: "TECHNICAL VAULT",
    description: "A secure repository of diverse coding and MCQ questions covering DSA, DBMS, OS, and JavaScript.",
    color: "yellow" as const,
    tag: "VAULT",
  },
]

const colorMap = {
  cyan: {
    text: "text-neon-cyan",
    border: "border-neon-cyan/20",
    hoverBorder: "hover:border-neon-cyan/40",
    glow: "group-hover:shadow-[0_0_30px_rgba(0,240,255,0.06)]",
    tag: "text-neon-cyan border-neon-cyan/30",
    accent: "bg-neon-cyan/40",
    iconBg: "bg-neon-cyan/5",
  },
  pink: {
    text: "text-neon-pink",
    border: "border-neon-pink/20",
    hoverBorder: "hover:border-neon-pink/40",
    glow: "group-hover:shadow-[0_0_30px_rgba(255,45,111,0.06)]",
    tag: "text-neon-pink border-neon-pink/30",
    accent: "bg-neon-pink/40",
    iconBg: "bg-neon-pink/5",
  },
  yellow: {
    text: "text-neon-yellow",
    border: "border-neon-yellow",
    hoverBorder: "hover:border-neon-yellow shadow-[0_0_15px_rgba(255,184,0,0.1)]",
    glow: "group-hover:shadow-[0_0_30px_rgba(255,184,0,0.1)]",
    tag: "text-neon-yellow border-neon-yellow/50",
    accent: "bg-neon-yellow/40",
    iconBg: "bg-neon-yellow/5",
  },
}

export function FeaturesSection() {
  return (
    <section className="relative bg-background py-24 lg:py-32">
      <div className="absolute inset-0 grid-bg opacity-40" />

      <div className="relative z-10 mx-auto max-w-7xl px-4 lg:px-8">
        {/* Section header */}
        <div className="flex items-center gap-4 mb-4">
          <Activity className="h-4 w-4 text-neon-cyan" />
          <span className="font-mono text-[11px] tracking-[0.3em] text-neon-cyan">
            COMBAT SYSTEMS
          </span>
          <div className="h-px flex-1 bg-panel-border" />
        </div>

        <h2 className="text-3xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl text-balance">
          BUILT FOR <span className="text-neon-cyan text-glow-cyan">BATTLE</span>
        </h2>
        <p className="mt-4 max-w-xl text-sm leading-relaxed text-muted-foreground lg:text-base">
          Every feature is engineered to push your cognitive limits. No lectures. No videos.
          Just you versus the clock.
        </p>

        {/* Feature grid */}
        <div className="mt-16 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => {
            const colors = colorMap[feature.color as keyof typeof colorMap]
            return (
              <div
                key={feature.title}
                className={`group relative border ${colors.border} hover:border-${feature.color === 'yellow' ? 'neon-yellow/60' : feature.color === 'cyan' ? 'neon-cyan/60' : 'neon-pink/60'} bg-panel-bg/40 p-6 transition-all duration-300 hover:shadow-[0_0_20px_rgba(251,191,36,0.08)] backdrop-blur-sm overflow-hidden`}
              >
                {/* Tag */}
                <div className="flex items-center justify-between mb-5">
                  <span
                    className={`font-mono text-[9px] tracking-[0.2em] border px-2 py-0.5 ${colors.tag} opacity-60 group-hover:opacity-100 transition-opacity`}
                  >
                    {feature.tag}
                  </span>
                  <Zap className="h-3 w-3 text-panel-border group-hover:text-muted-foreground transition-colors" />
                </div>

                {/* Icon */}
                <div className={`mb-4 inline-flex h-10 w-10 items-center justify-center border ${colors.border} group-hover:border-${feature.color === 'yellow' ? 'neon-yellow' : feature.color === 'cyan' ? 'neon-cyan' : 'neon-pink'} ${colors.iconBg} ${colors.text} opacity-70 group-hover:opacity-100 transition-all`}>
                  <feature.icon className="h-5 w-5" strokeWidth={1.5} />
                </div>

                {/* Content */}
                <h3 className="font-mono text-xs font-bold tracking-widest text-foreground/80 group-hover:text-foreground mb-2 transition-colors">
                  {feature.title}
                </h3>
                <p className="text-sm leading-relaxed text-muted-foreground/80 group-hover:text-muted-foreground transition-colors">
                  {feature.description}
                </p>

                {/* Bottom accent line */}
                <div
                  className={`absolute bottom-0 left-0 h-px w-0 group-hover:w-full transition-all duration-500 ${colors.accent}`}
                />
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
