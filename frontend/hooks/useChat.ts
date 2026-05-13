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
    
    if (!token || wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('[CHAT] Skipping connection - token:', !!token, 'wsState:', wsRef.current?.readyState)
      return
    }

    // Use the same API_URL as the rest of the app
    const wsBase = API_URL.replace('https://', 'wss://').replace('http://', 'ws://')
    const wsUrl = `${wsBase}/ws/chat?token=${token}`
    
    console.log('[CHAT] Connecting to:', wsUrl)
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('[CHAT] WebSocket connected to:', wsUrl)
      setIsConnected(true)
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        console.log('[CHAT] Message received:', msg)
        console.log('[CHAT] Message type:', msg.type)
        console.log('[CHAT] Current messages array length:', messages.length)
        
        if (msg.type === 'message') {
          console.log('[CHAT] Adding message to state:', msg)
          setMessages((prev) => {
            const newMessages = [...prev, msg]
            console.log('[CHAT] New messages array length:', newMessages.length)
            return newMessages
          })
        }
        if (msg.type === 'online_count') {
          console.log('[CHAT] Updating online count:', msg.online_count)
          setOnlineCount(msg.online_count)
        }
        if (msg.type === 'user_joined' || msg.type === 'user_left') {
          console.log('[CHAT] User event:', msg.type)
          if (msg.online_count !== undefined) {
            setOnlineCount(msg.online_count)
          }
        }
      } catch (e) {
        console.error('[CHAT] Failed to parse message:', e)
      }
    }

    ws.onerror = (err) => {
      console.error('[CHAT] WebSocket error:', err)
    }

    ws.onclose = (e) => {
      console.log('[CHAT] WebSocket closed:', e.code, e.reason)
      setIsConnected(false)
      wsRef.current = null

      // Attempt reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, 3000)
    }

    wsRef.current = ws
  }, [token])

  // Connect on mount and when token changes
  useEffect(() => {
    console.log('[CHAT] useEffect triggered, token:', token ? 'EXISTS' : 'NULL')
    connect()

    return () => {
      console.log('[CHAT] Cleaning up WebSocket connection')
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connect])

  // Send a text message
  const sendMessage = useCallback(
    (content: string, messageType: string = "text") => {
      if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
        console.error("[CHAT] WebSocket not connected, readyState:", wsRef.current?.readyState)
        return
      }

      const event: Partial<ChatEvent> = {
        type: "message",
        message_type: messageType,
        content,
        timestamp: new Date().toISOString(),
      }

      console.log('[CHAT] Sending message:', event)
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
