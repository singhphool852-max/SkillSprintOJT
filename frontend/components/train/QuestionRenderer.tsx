"use client"
import { useEffect } from "react"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

import { AIDebrief } from "./ai-debrief"

interface Question {
  id: string
  prompt: string
  type: string
  options?: { id: string, text: string }[]
  template?: string
  hint?: string
  starterCode?: string
  constraints?: string
  testCases?: any[]
}

interface RendererProps {
  question: Question
  answer: string
  onChange: (value: string) => void
  isLocked?: boolean
  result?: { isCorrect: boolean, feedback: string, explanation: string, correctOptionId?: string }
}

export function QuestionRenderer({ question, answer, onChange, isLocked, result }: RendererProps) {
  // Normalize types safely
  const type = (question?.type || "mcq").toLowerCase()

  // Use starter code if answer is empty and we have it
  useEffect(() => {
    if (!answer && question?.starterCode && !isLocked) {
      onChange(question.starterCode)
    }
  }, [question?.id, question?.starterCode, answer, isLocked, onChange])

  // 1. MCQ Renderer
  if (type === "mcq") {
    const options = question?.options || []
    if (options.length === 0) {
      return (
        <div className="p-6 border border-neon-pink/30 bg-neon-pink/5 font-mono text-[10px] text-neon-pink uppercase tracking-widest animate-pulse">
           Error: No options discovered for this query.
        </div>
      )
    }

    return (
      <div className="space-y-6">
        <div className="grid gap-3 sm:grid-cols-2">
          {options.map((opt) => {
            const isSelected = answer === opt.id
            const isCorrect = result?.correctOptionId === opt.id
            const isWrong = isSelected && result && !isCorrect
            
            let borderClass = "border-panel-border bg-panel-bg/60 hover:border-neon-cyan/40"
            if (isSelected) borderClass = "border-neon-cyan bg-neon-cyan/10 shadow-[0_0_15px_rgba(0,240,255,0.2)]"
            if (isLocked && isCorrect) borderClass = "border-neon-green bg-neon-green/10 shadow-[0_0_20px_rgba(0,255,150,0.15)]"
            if (isLocked && isWrong) borderClass = "border-neon-pink bg-neon-pink/10 shadow-[0_0_20px_rgba(255,45,111,0.15)]"

            return (
              <button
                key={opt.id}
                disabled={isLocked}
                onClick={() => onChange(opt.id)}
                className={`p-5 text-left border transition-all ${borderClass} ${isLocked ? 'cursor-default' : ''}`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-mono text-sm text-foreground">{opt.text || String(opt.id)}</span>
                  {isLocked && isCorrect && <div className="h-1.5 w-1.5 rounded-full bg-neon-green animate-pulse" />}
                </div>
              </button>
            )
          })}
        </div>
        {result && (
          <AIDebrief 
            isCorrect={result.isCorrect} 
            feedback={result.feedback} 
            explanation={result.explanation} 
          />
        )}
      </div>
    )
  }

  // 2. Code Renderer (Debug, Fix, Write)
  if (type.includes("code") || type.includes("debug") || type.includes("fix")) {
    return (
      <div className="flex flex-col gap-6">
        {question.constraints && (
          <div className="p-4 border border-panel-border bg-white/[0.02] flex flex-col gap-2">
            <span className="font-mono text-[10px] text-neon-cyan uppercase tracking-widest font-bold">Constraints</span>
            <p className="font-mono text-[11px] text-muted-foreground whitespace-pre-wrap">{question.constraints}</p>
          </div>
        )}
        <div className="relative group">
           <div className="absolute -inset-0.5 bg-gradient-to-r from-neon-cyan/10 to-transparent opacity-0 group-focus-within:opacity-100 transition-opacity" />
           <textarea
            disabled={isLocked}
            value={answer || question.template || ""}
            onChange={(e) => onChange(e.target.value)}
            onKeyDown={(e) => {
              if (isLocked) return;
              const target = e.currentTarget
              const start = target.selectionStart
              const end = target.selectionEnd
              const value = target.value

              const pairs: Record<string, string> = {
                "{": "}",
                "(": ")",
                "[": "]",
                '"': '"',
                "'": "'",
                "`": "`",
              }

              // 1. AUTO-CLOSE: insert pair and place cursor between them
              if (pairs[e.key] && !e.ctrlKey && !e.metaKey && !e.altKey) {
                e.preventDefault()
                const open = e.key
                const close = pairs[open]
                const before = value.slice(0, start)
                const selected = value.slice(start, end)
                const after = value.slice(end)

                const newValue = before + open + selected + close + after
                const newCursor = start + 1 + selected.length

                onChange(newValue)
                requestAnimationFrame(() => {
                  target.selectionStart = target.selectionEnd = newCursor
                })
                return
              }

              // 2. SKIP OVER: if next char is closing bracket and user types it, skip
              const closers = new Set([")", "]", "}", '"', "'", "`"])
              if (
                closers.has(e.key) &&
                value[start] === e.key &&
                start === end &&
                !e.ctrlKey && !e.metaKey
              ) {
                e.preventDefault()
                target.selectionStart = target.selectionEnd = start + 1
                return
              }

              // 3. Tab key: insert 4 spaces
              if (e.key === "Tab") {
                e.preventDefault()
                if (e.shiftKey) {
                  // Shift+Tab: dedent
                  const before = value.substring(0, start)
                  const lineStart = before.lastIndexOf("\n") + 1
                  const linePrefix = value.substring(lineStart, start)
                  const leadingMatch = linePrefix.match(/^( {1,4})/)
                  if (leadingMatch) {
                    const removeCount = leadingMatch[1].length
                    const newValue = value.substring(0, lineStart) + value.substring(lineStart + removeCount)
                    onChange(newValue)
                    const newPos = Math.max(lineStart, start - removeCount)
                    requestAnimationFrame(() => {
                      target.selectionStart = target.selectionEnd = newPos
                    })
                  }
                } else {
                  // Tab: insert 4 spaces
                  const newValue = value.substring(0, start) + "    " + value.substring(end)
                  onChange(newValue)
                  requestAnimationFrame(() => {
                    target.selectionStart = target.selectionEnd = start + 4
                  })
                }
                return
              }

              // 4. Enter key: auto-indent
              if (e.key === "Enter" && !e.ctrlKey && !e.metaKey) {
                e.preventDefault()
                const before = value.slice(0, start)
                const after = value.slice(end)
                const lastNewLine = before.lastIndexOf("\n")
                const currentLine = before.slice(lastNewLine + 1)
                const indentMatch = currentLine.match(/^\s*/)
                const indent = indentMatch ? indentMatch[0] : ""

                // Extra indent if opening a block
                let extraIndent = ""
                const charBefore = before.trim().slice(-1)
                if (charBefore === "{" || charBefore === ":" || charBefore === "(") {
                   extraIndent = "    "
                }

                // If pressing enter between { and }, expand the block
                if (charBefore === "{" && after.trim().startsWith("}")) {
                  const newValue = before + "\n" + indent + extraIndent + "\n" + indent + after
                  onChange(newValue)
                  requestAnimationFrame(() => {
                    target.selectionStart = target.selectionEnd = start + 1 + indent.length + extraIndent.length
                  })
                  return
                }

                const newValue = before + "\n" + indent + extraIndent + after
                onChange(newValue)
                requestAnimationFrame(() => {
                  target.selectionStart = target.selectionEnd = start + 1 + indent.length + extraIndent.length
                })
                return
              }
            }}
            placeholder="// Enter your code solution here..."
            className={`relative w-full h-[400px] bg-deep-bg/60 border ${result ? (result.isCorrect ? 'border-neon-green/40' : 'border-neon-pink/40') : 'border-panel-border'} p-6 font-mono text-xs leading-relaxed text-foreground focus:border-neon-cyan/50 focus:outline-none resize-none transition-colors`}
            style={{ tabSize: 4, MozTabSize: 4 }}
            spellCheck={false}
          />
          {result && (
            <div className={`absolute top-4 right-4 font-mono text-[9px] px-2 py-0.5 border ${result.isCorrect ? 'text-neon-green border-neon-green/30 bg-neon-green/5' : 'text-neon-pink border-neon-pink/30 bg-neon-pink/5'} uppercase tracking-widest`}>
              {result.isCorrect ? 'CORRECT' : 'INCORRECT'}
            </div>
          )}
        </div>
        {result && (
          <AIDebrief 
            isCorrect={result.isCorrect} 
            feedback={result.feedback} 
            explanation={result.explanation} 
          />
        )}
      </div>
    )
  }

  // 3. Logic / Subjective Renderer (Explanation, Scenario, Brainstorm, or Unknown)
  const isGeneric = !type.includes("logic") && !type.includes("scenario") && !type.includes("brainstorm") && !type.includes("explanation")
  
  return (
    <div className="flex flex-col gap-6">
      <div className="relative group">
        {isGeneric && (
           <div className="absolute top-0 right-0 -mt-3 mr-4 z-10 px-2 py-0.5 bg-panel-bg border border-neon-cyan/50">
             <span className="font-mono text-[8px] text-neon-cyan uppercase font-bold tracking-widest leading-none block">Universal Input Fallback</span>
           </div>
        )}
        <textarea
          disabled={isLocked}
          value={answer}
          onChange={(e) => onChange(e.target.value)}
          placeholder={isGeneric ? "Awaiting neural input..." : "Deconstruct your logic here..."}
          className={`w-full h-60 bg-deep-bg/60 border ${result ? (result.isCorrect ? 'border-neon-green/40' : 'border-neon-pink/40') : (isGeneric ? 'border-neon-cyan/30' : 'border-panel-border')} p-6 font-mono text-sm leading-relaxed text-foreground focus:border-neon-cyan/50 focus:outline-none resize-none transition-colors`}
        />
        {result && (
            <div className={`absolute top-4 right-4 font-mono text-[9px] px-2 py-0.5 border ${result.isCorrect ? 'text-neon-green border-neon-green/30 bg-neon-green/5' : 'text-neon-pink border-neon-pink/30 bg-neon-pink/5'} uppercase tracking-widest`}>
              {result.isCorrect ? 'CORRECT' : 'INCORRECT'}
            </div>
        )}
      </div>
      {result && (
          <AIDebrief 
            isCorrect={result.isCorrect} 
            feedback={result.feedback} 
            explanation={result.explanation} 
          />
      )}
    </div>
  )
}
