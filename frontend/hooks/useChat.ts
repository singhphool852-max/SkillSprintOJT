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
    if (!token || wsRef.current?.readyState === WebSocket.OPEN) return

    const wsUrl = API_URL.replace("http", "ws")
    const ws = new WebSocket(`${wsUrl}/ws/chat?token=${token}`)

    ws.onopen = () => {
      console.log("[CHAT] WebSocket connected")
      setIsConnected(true)
    }

    ws.onmessage = (event) => {
      try {
        const data: ChatEvent = JSON.parse(event.data)

        if (data.type === "message") {
          setMessages((prev) => [...prev, data])
        } else if (data.type === "user_joined" || data.type === "user_left") {
          if (data.online_count !== undefined) {
            setOnlineCount(data.online_count)
          }
        } else if (data.type === "online_count") {
          if (data.online_count !== undefined) {
            setOnlineCount(data.online_count)
          }
        }
      } catch (error) {
        console.error("[CHAT] Failed to parse message:", error)
      }
    }

    ws.onerror = (error) => {
      console.error("[CHAT] WebSocket error:", error)
    }

    ws.onclose = () => {
      console.log("[CHAT] WebSocket closed, reconnecting in 3s...")
      setIsConnected(false)
      wsRef.current = null

      // Attempt reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, 3000)
    }

    wsRef.current = ws
  }, [token])

  // Connect on mount
  useEffect(() => {
    connect()

    return () => {
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
        console.error("[CHAT] WebSocket not connected")
        return
      }

      const event: Partial<ChatEvent> = {
        type: "message",
        message_type: messageType,
        content,
        timestamp: new Date().toISOString(),
      }

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
