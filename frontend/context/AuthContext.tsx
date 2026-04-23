"use client"

import React, { createContext, useState, useEffect, ReactNode } from "react"
import { useRouter } from "next/navigation"

export interface User {
  id: string
  username: string
  email: string
}

export interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  isLoading: boolean
  setUser: (user: User | null) => void
  setIsLoading: (isLoading: boolean) => void
  logout: () => Promise<void>
  checkAuth: () => Promise<void>
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  const checkAuth = async () => {
    try {
      const res = await fetch("http://localhost:8080/api/auth/me", {
        credentials: "include",
      })
      if (res.ok) {
        const data = await res.json()
        setUser({ id: data.id, email: data.email, username: data.username })
      } else {
        setUser(null)
      }
    } catch (error) {
      console.error("Auth check failed", error)
      setUser(null)
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    checkAuth()
  }, [])

  const logout = async () => {
    try {
      await fetch("http://localhost:8080/api/auth/logout", {
        method: "POST",
        credentials: "include",
      })
    } catch (error) {
      console.error("Logout failed", error)
    } finally {
      setUser(null)
      router.push("/login")
    }
  }

  const value = {
    user,
    isAuthenticated: !!user,
    isLoading,
    setUser,
    setIsLoading,
    logout,
    checkAuth,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
