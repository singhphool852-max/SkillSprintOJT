"use client"

import { useState } from "react"
import { ArenaLobby } from "@/components/arena/arena-lobby"
import { TestArena } from "@/components/arena/test-arena"

export default function ArenaPage() {
  const [isTestActive, setIsTestActive] = useState(false)

  // When a test is active, hide the lobby and render the test
  // in a fullscreen-style container that takes over the page.
  if (isTestActive) {
    return (
      <main className="fixed inset-0 z-50 bg-deep-bg overflow-auto">
        <TestArena onActiveChange={setIsTestActive} />
      </main>
    )
  }

  return (
    <main className="min-h-screen bg-deep-bg">
      <div className="pt-20">
        <ArenaLobby />
      </div>

      {/* Coding Tests Section */}
      <div className="relative mx-auto max-w-7xl px-4 lg:px-8 py-4">
        <div className="h-px bg-panel-border" />
      </div>
      <TestArena onActiveChange={setIsTestActive} />
    </main>
  )
}
