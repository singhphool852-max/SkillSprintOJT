"use client"

import { useEffect } from "react"
import { useRouter, usePathname } from "next/navigation"
import { useAuth } from "../hooks/useAuth"
import { Loader2 } from "lucide-react"

interface ProtectedRouteProps {
  children: React.ReactNode
  requiredRole?: "student" | "admin"
}

export function ProtectedRoute({ children, requiredRole }: ProtectedRouteProps) {
  const { user, isAuthenticated, isLoading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push(`/login?redirect=${encodeURIComponent(pathname)}`)
    }
    // If authenticated but role doesn't match, redirect to login
    if (!isLoading && isAuthenticated && requiredRole && user?.role !== requiredRole) {
      router.push("/login")
    }
  }, [isLoading, isAuthenticated, router, pathname, requiredRole, user])

  if (isLoading) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen bg-deep-bg gap-4">
        <Loader2 className="h-8 w-8 text-neon-cyan animate-spin" />
        <span className="font-mono text-xs tracking-widest text-neon-cyan uppercase">
          VERIFYING CLEARANCE...
        </span>
      </div>
    )
  }

  if (!isAuthenticated) {
    return null // Will redirect in useEffect
  }

  if (requiredRole && user?.role !== requiredRole) {
    return null // Will redirect in useEffect
  }

  return <>{children}</>
}
