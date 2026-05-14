"use client"

import { useState, useRef, useEffect } from "react"
import { useAuth } from "@/hooks/useAuth"
import { useChat } from "@/hooks/useChat"
import { useRouter } from "next/navigation"
import { Paperclip, FileText, Send, Download, Users } from "lucide-react"
import { API_URL } from "@/lib/api-config"

export default function ChatPage() {
  const { user, isLoading } = useAuth()
  const router = useRouter()
  const [token, setToken] = useState<string | null>(null)
  const [messageInput, setMessageInput] = useState("")
  const [showNoteModal, setShowNoteModal] = useState(false)
  const [noteContent, setNoteContent] = useState("")
  const [uploading, setUploading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Get token from localStorage (not cookies since HttpOnly prevents JS access)
  useEffect(() => {
    const retrievedToken = localStorage.getItem('auth_token')
    console.log('[CHAT] Token retrieved from localStorage:', retrievedToken ? 'YES (length: ' + retrievedToken.length + ')' : 'NO')
    console.log('[CHAT] API_URL:', API_URL)
    
    // Test if backend is reachable
    fetch(`${API_URL}/api/chat/test`)
      .then(r => r.json())
      .then(data => console.log('[CHAT] Backend test:', data))
      .catch(e => console.error('[CHAT] Backend NOT reachable:', e))
    
    setToken(retrievedToken)
  }, [user])

  const { messages, onlineCount, isConnected, sendMessage, sendFile } = useChat(token)

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !user) {
      router.push("/login")
    }
  }, [user, isLoading, router])

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  const handleSendMessage = () => {
    if (!messageInput.trim()) return
    sendMessage(messageInput, "text")
    setMessageInput("")
  }

  const handleSendNote = () => {
    if (!noteContent.trim()) return
    sendMessage(noteContent, "note")
    setNoteContent("")
    setShowNoteModal(false)
  }

  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate file type
    const allowedTypes = ["image/jpeg", "image/jpg", "image/png", "image/gif", "application/pdf"]
    if (!allowedTypes.includes(file.type)) {
      alert("Only images (JPEG, PNG, GIF) and PDFs are allowed")
      return
    }

    // Validate file size (10MB)
    if (file.size > 10 * 1024 * 1024) {
      alert("File size must be less than 10MB")
      return
    }

    setUploading(true)
    try {
      await sendFile(file)
    } catch (error: any) {
      alert(error.message || "Failed to upload file")
    } finally {
      setUploading(false)
      if (fileInputRef.current) {
        fileInputRef.current.value = ""
      }
    }
  }

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" })
  }

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-deep-bg">
        <div className="h-12 w-12 animate-pulse border border-neon-cyan/50 bg-neon-cyan/10" />
      </div>
    )
  }

  if (!user) return null

  return (
    <div className="flex h-screen flex-col bg-deep-bg pt-16">
      {/* Header */}
      <div className="border-b border-panel-border bg-panel-bg/60 px-6 py-4 backdrop-blur-xl">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-mono text-xl font-bold tracking-wider text-neon-cyan">
              💬 SKILLSPRINT COMMUNITY CHAT
            </h1>
            <p className="mt-1 font-mono text-xs text-muted-foreground">
              Share notes, images, and PDFs with the community
            </p>
          </div>
          <div className="flex items-center gap-2 border border-neon-cyan/30 bg-neon-cyan/5 px-4 py-2">
            <Users className="h-4 w-4 text-neon-cyan" />
            <span className="font-mono text-sm font-bold text-neon-cyan">
              {onlineCount} ONLINE
            </span>
            <div
              className={`h-2 w-2 rounded-full ${isConnected ? "bg-green-500 animate-pulse" : "bg-red-500"}`}
            />
          </div>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto px-6 py-4">
        <div className="mx-auto max-w-4xl space-y-4">
          {messages.length === 0 && (
            <div className="text-center py-8">
              <p className="font-mono text-sm text-muted-foreground">
                No messages yet. Start the conversation!
              </p>
            </div>
          )}

          {messages.map((msg, idx) => {
            const isOwnMessage = msg.user_id === user.id
            const isSystemMessage = msg.type !== "message"

            // Skip system messages (user_joined, user_left, etc.)
            if (isSystemMessage) return null

            return (
              <div
                key={idx}
                className={`flex ${isOwnMessage ? "justify-end" : "justify-start"}`}
              >
                <div className={`max-w-[70%] ${isOwnMessage ? "items-end" : "items-start"} flex flex-col gap-1`}>
                  {/* Username and time */}
                  {!isOwnMessage && (
                    <div className="flex items-center gap-2 px-2">
                      {msg.avatar && (
                        <img
                          src={msg.avatar}
                          alt={msg.username}
                          className="h-5 w-5 rounded-full border border-neon-cyan/30"
                        />
                      )}
                      <span className="font-mono text-xs text-muted-foreground">
                        {msg.username}
                      </span>
                      <span className="font-mono text-[10px] text-muted-foreground/50">
                        {formatTime(msg.timestamp)}
                      </span>
                    </div>
                  )}

                  {/* Message bubble */}
                  <div
                    className={`border px-4 py-3 ${
                      msg.message_type === "note"
                        ? "border-neon-yellow/30 bg-neon-yellow/10"
                        : isOwnMessage
                        ? "border-neon-pink/30 bg-neon-pink/10"
                        : "border-panel-border bg-panel-bg"
                    }`}
                  >
                    {msg.message_type === "text" && (
                      <p className="whitespace-pre-wrap font-mono text-sm text-foreground">
                        {msg.content}
                      </p>
                    )}

                    {msg.message_type === "note" && (
                      <div>
                        <div className="mb-2 flex items-center gap-2">
                          <FileText className="h-4 w-4 text-neon-yellow" />
                          <span className="font-mono text-xs font-bold text-neon-yellow">
                            NOTE
                          </span>
                        </div>
                        <p className="whitespace-pre-wrap font-mono text-sm text-foreground">
                          {msg.content}
                        </p>
                      </div>
                    )}

                    {msg.message_type === "image" && (
                      <div>
                        <img
                          src={`${API_URL}${msg.content}`}
                          alt="Shared image"
                          className="max-w-[300px] border border-neon-cyan/20"
                        />
                      </div>
                    )}

                    {msg.message_type === "pdf" && (
                      <div className="flex items-center gap-3">
                        <div className="flex h-12 w-12 items-center justify-center border border-neon-cyan/30 bg-neon-cyan/5">
                          <FileText className="h-6 w-6 text-neon-cyan" />
                        </div>
                        <div className="flex-1">
                          <p className="font-mono text-sm font-bold text-foreground">
                            {msg.file_name || "Document.pdf"}
                          </p>
                          <a
                            href={`${API_URL}${msg.content}`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-1 font-mono text-xs text-neon-cyan hover:underline"
                          >
                            <Download className="h-3 w-3" />
                            Download
                          </a>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Time for own messages */}
                  {isOwnMessage && (
                    <div className="px-2">
                      <span className="font-mono text-[10px] text-muted-foreground/50">
                        {formatTime(msg.timestamp)} ✓
                      </span>
                    </div>
                  )}
                </div>
              </div>
            )
          })}
          <div ref={messagesEndRef} />
        </div>
      </div>

      {/* Input Bar */}
      <div className="border-t border-panel-border bg-panel-bg/60 px-6 py-4 backdrop-blur-xl">
        <div className="mx-auto flex max-w-4xl items-center gap-3">
          {/* File upload button */}
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            className="flex h-10 w-10 items-center justify-center border border-neon-cyan/30 bg-neon-cyan/5 transition-colors hover:bg-neon-cyan/10 disabled:opacity-50"
            title="Upload image or PDF"
          >
            <Paperclip className="h-4 w-4 text-neon-cyan" />
          </button>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/jpg,image/png,image/gif,application/pdf"
            onChange={handleFileSelect}
            className="hidden"
          />

          {/* Note button */}
          <button
            onClick={() => setShowNoteModal(true)}
            className="flex h-10 w-10 items-center justify-center border border-neon-yellow/30 bg-neon-yellow/5 transition-colors hover:bg-neon-yellow/10"
            title="Share a note"
          >
            <FileText className="h-4 w-4 text-neon-yellow" />
          </button>

          {/* Text input */}
          <input
            type="text"
            value={messageInput}
            onChange={(e) => setMessageInput(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
            placeholder="Type a message..."
            className="flex-1 border border-panel-border bg-deep-bg px-4 py-2 font-mono text-sm text-foreground placeholder:text-muted-foreground focus:border-neon-cyan/50 focus:outline-none"
          />

          {/* Send button */}
          <button
            onClick={handleSendMessage}
            disabled={!messageInput.trim()}
            className="flex h-10 w-10 items-center justify-center border border-neon-cyan/50 bg-neon-cyan/10 transition-colors hover:bg-neon-cyan/20 disabled:opacity-50"
          >
            <Send className="h-4 w-4 text-neon-cyan" />
          </button>
        </div>
      </div>

      {/* Note Modal */}
      {showNoteModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
          <div className="w-full max-w-2xl border border-neon-yellow/30 bg-panel-bg p-6">
            <h2 className="mb-4 font-mono text-lg font-bold text-neon-yellow">
              SHARE A NOTE
            </h2>
            <textarea
              value={noteContent}
              onChange={(e) => setNoteContent(e.target.value)}
              placeholder="Write your note here..."
              rows={8}
              className="w-full border border-panel-border bg-deep-bg px-4 py-3 font-mono text-sm text-foreground placeholder:text-muted-foreground focus:border-neon-yellow/50 focus:outline-none"
            />
            <div className="mt-4 flex justify-end gap-3">
              <button
                onClick={() => {
                  setShowNoteModal(false)
                  setNoteContent("")
                }}
                className="border border-panel-border bg-deep-bg px-6 py-2 font-mono text-xs tracking-widest text-muted-foreground transition-colors hover:text-foreground"
              >
                CANCEL
              </button>
              <button
                onClick={handleSendNote}
                disabled={!noteContent.trim()}
                className="border border-neon-yellow/50 bg-neon-yellow/10 px-6 py-2 font-mono text-xs tracking-widest text-neon-yellow transition-colors hover:bg-neon-yellow/20 disabled:opacity-50"
              >
                SEND NOTE
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Uploading indicator */}
      {uploading && (
        <div className="fixed bottom-24 right-6 border border-neon-cyan/30 bg-panel-bg px-4 py-3">
          <p className="font-mono text-sm text-neon-cyan">Uploading...</p>
        </div>
      )}
    </div>
  )
}
