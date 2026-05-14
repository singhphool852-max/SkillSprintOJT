"use client"

import { useState, useEffect, useRef, useCallback } from "react"
import { API_URL } from "@/lib/api-config"

export interface ChatEvent {
  type: string // "message", "user_joined", "user_left", "online_count"
  user_id: string
  username: string
  avatar: string
  message_type: string // "text", "note", "image", "pdf"
  content: string
  file_name?: string
  timestamp: string
  online_count?: number
}

export function useChat(token: string | null) {
  const [messages, setMessages] = useState<ChatEvent[]>([])
  const [onlineCount, setOnlineCount] = useState(0)
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>()
  const reconnectAttemptsRef = useRef(0)

  // Load chat history on mount
  useEffect(() => {
    if (!token) return

    const loadHistory = async () => {
      try {
        const res = await fetch(`${API_URL}/api/chat/history`, {
          credentials: "include",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        })
        if (res.ok) {
          const history = await res.json()
          console.log('[CHAT] Loaded history:', history.length, 'messages')
          setMessages(history)
        }
      } catch (error) {
        console.error("[CHAT] Failed to load history:", error)
      }
    }

    loadHistory()
  }, [token])

  // Connect to WebSocket
  const connect = useCallback(() => {
    console.log('[CHAT] connect() called, token:', token ? 'EXISTS' : 'NULL')
    
    if (!token) {
      console.log('[CHAT] No token available, cannot connect')
      return
    }

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('[CHAT] Already connected')
      return
    }

    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }

    // Build WebSocket URL with token as query parameter
    const wsBase = API_URL.replace('https://', 'wss://').replace('http://', 'ws://')
    const wsUrl = `${wsBase}/ws/chat?token=${encodeURIComponent(token)}`
    
    console.log('[CHAT] Connecting to:', wsUrl.replace(token, 'TOKEN_HIDDEN'))
    
    try {
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('[CHAT] ✅ WebSocket connected successfully')
        setIsConnected(true)
        reconnectAttemptsRef.current = 0
      }

      ws.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data)
          console.log('[CHAT] 📨 Message received:', msg.type, msg)
          
          if (msg.type === 'message') {
            setMessages((prev) => {
              // Prevent duplicate messages
              const isDuplicate = prev.some(
                m => m.timestamp === msg.timestamp &&
                     m.user_id === msg.user_id &&
                     m.content === msg.content
              )
              if (isDuplicate) {
                console.log('[CHAT] Duplicate message detected, skipping')
                return prev
              }
              console.log('[CHAT] Adding message to state')
              return [...prev, msg]
            })
          }
          
          if (msg.type === 'online_count' || msg.online_count !== undefined) {
            console.log('[CHAT] Updating online count:', msg.online_count)
            setOnlineCount(msg.online_count)
          }
          
          if (msg.type === 'user_joined' || msg.type === 'user_left') {
            console.log('[CHAT] 👤 User event:', msg.type, msg.username)
            if (msg.online_count !== undefined) {
              setOnlineCount(msg.online_count)
            }
          }
        } catch (e) {
          console.error('[CHAT] ❌ Failed to parse message:', e, event.data)
        }
      }

      ws.onerror = (err) => {
        console.error('[CHAT] ❌ WebSocket error:', err)
      }

      ws.onclose = (e) => {
        console.log('[CHAT] 🔌 WebSocket closed:', e.code, e.reason)
        setIsConnected(false)
        wsRef.current = null

        // Attempt reconnect with exponential backoff
        reconnectAttemptsRef.current++
        const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current - 1), 10000)
        console.log(`[CHAT] 🔄 Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current})`)
        
        reconnectTimeoutRef.current = setTimeout(() => {
          connect()
        }, delay)
      }

      wsRef.current = ws
    } catch (error) {
      console.error('[CHAT] ❌ Failed to create WebSocket:', error)
    }
  }, [token])

  // Connect on mount and when token changes
  useEffect(() => {
    console.log('[CHAT] useEffect triggered, token:', token ? 'EXISTS' : 'NULL')
    connect()

    return () => {
      console.log('[CHAT] 🧹 Cleaning up WebSocket connection')
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [connect])

  // Send a text message
  const sendMessage = useCallback(
    (content: string, messageType: string = "text") => {
      if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
        console.error("[CHAT] ❌ Cannot send message - WebSocket not connected, readyState:", wsRef.current?.readyState)
        return
      }

      const event: Partial<ChatEvent> = {
        type: "message",
        message_type: messageType,
        content,
        timestamp: new Date().toISOString(),
      }

      console.log('[CHAT] 📤 Sending message:', event)
      wsRef.current.send(JSON.stringify(event))
    },
    []
  )

  // Upload and send a file
  const sendFile = useCallback(
    async (file: File): Promise<void> => {
      if (!token) {
        throw new Error("Not authenticated")
      }

      console.log('[CHAT] 📎 Uploading file:', file.name, file.type, file.size)

      const formData = new FormData()
      formData.append("file", file)

      const res = await fetch(`${API_URL}/api/chat/upload`, {
        method: "POST",
        credentials: "include",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      })

      if (!res.ok) {
        const error = await res.json()
        throw new Error(error.error || "Upload failed")
      }

      const data = await res.json()
      console.log('[CHAT] ✅ File uploaded:', data.url)

      // Determine message type based on file type
      const messageType = file.type.startsWith("image/") ? "image" : "pdf"

      // Send message with file URL
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        const event: Partial<ChatEvent> = {
          type: "message",
          message_type: messageType,
          content: data.url,
          file_name: data.filename,
          timestamp: new Date().toISOString(),
        }
        console.log('[CHAT] 📤 Sending file message:', event)
        wsRef.current.send(JSON.stringify(event))
      }
    },
    [token]
  )

  return {
    messages,
    onlineCount,
    isConnected,
    sendMessage,
    sendFile,
  }
}
