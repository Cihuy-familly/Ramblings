import React, { createContext, useContext, useState, useEffect, useCallback } from 'react'
import type { Creator } from '../types'
import { getMe as apiGetMe, login as apiLogin } from '../api/client'

interface AuthContextType {
  creator: Creator | null
  token: string | null
  loading: boolean
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  setToken: (token: string) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [creator, setCreator] = useState<Creator | null>(null)
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('token'))
  const [loading, setLoading] = useState(true)

  const isAuthenticated = creator !== null

  useEffect(() => {
    async function validateToken() {
      const storedToken = localStorage.getItem('token')
      if (!storedToken) {
        setLoading(false)
        return
      }

      try {
        const { creator: me } = await apiGetMe()
        setCreator(me)
      } catch {
        localStorage.removeItem('token')
        setToken(null)
        setCreator(null)
      } finally {
        setLoading(false)
      }
    }

    validateToken()
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const { token: newToken, creator: newCreator } = await apiLogin(email, password)
    localStorage.setItem('token', newToken)
    setToken(newToken)
    setCreator(newCreator)
  }, [])

  const handleSetToken = useCallback(async (newToken: string) => {
    localStorage.setItem('token', newToken)
    setToken(newToken)
    try {
      const { creator: me } = await apiGetMe()
      setCreator(me)
    } catch {
      // Token invalid, clean up
      localStorage.removeItem('token')
      setToken(null)
      setCreator(null)
    }
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    setToken(null)
    setCreator(null)
  }, [])

  return (
    <AuthContext.Provider
      value={{
        creator,
        token,
        loading,
        isAuthenticated,
        login,
        logout,
        setToken: handleSetToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}