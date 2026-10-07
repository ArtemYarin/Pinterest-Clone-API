import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { getAccessToken, setAccessToken } from '../localStorage/tokenStore'
import client from '../api/client'
import { refresh } from '../api/refresh'
import { logout as logoutRequest } from '../api/logout'

const AuthContext = createContext(null)

// The backend rotates the refresh token on every call, so concurrent refreshes
// (StrictMode double effects, several 401s at once) must share one request:
// a second request with the old cookie would be rejected and clear the session.
let refreshPromise = null
function refreshOnce() {
  if (!refreshPromise) {
    refreshPromise = refresh().finally(() => {
      refreshPromise = null
    })
  }
  return refreshPromise
}

// Reads the user_id claim from the jwt payload. Only used for building links;
// the backend still verifies the token on every request.
function userIdFromToken(token) {
  if (!token) return null
  try {
    const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    return JSON.parse(atob(payload)).user_id ?? null
  } catch {
    return null
  }
}

// Keeps the jwt access token in memory and the refresh token in an httpOnly
// cookie. Restores the session on load and attaches/renews the token for
// every request made through the api client.
export function AuthProvider({ children }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [isInitializing, setIsInitializing] = useState(true)
  const [userId, setUserId] = useState(null)

  const setAccessTokenC = useCallback((token) => {
    setAccessToken(token)
    setIsAuthenticated(!!token)
    setUserId(userIdFromToken(token))
  }, [])

  // Accepts the { token } payload returned by login, signup and refresh.
  const signIn = useCallback(
    (data) => setAccessTokenC(data.token),
    [setAccessTokenC],
  )

  const logout = useCallback(async () => {
    setAccessTokenC(null)
    try {
      await logoutRequest()
    } catch {
      // The local session is already gone; a failed revoke is not actionable here.
    }
  }, [setAccessTokenC])

  // Restore the session from the refresh cookie on reload.
  useEffect(() => {
    refreshOnce()
      .then(signIn)
      .catch(() => {})
      .finally(() => setIsInitializing(false))
  }, [signIn])

  useEffect(() => {
    const requestId = client.interceptors.request.use((config) => {
      const token = getAccessToken()
      if (token) config.headers.Authorization = `Bearer ${token}`
      return config
    })

    // On 401 refresh the access token once and retry the original request.
    const responseId = client.interceptors.response.use(
      undefined,
      async (error) => {
        const original = error.config
        const isAuthCall = original?.url?.startsWith('/auth/')
        if (
          error.response?.status !== 401 ||
          !original ||
          original._retried ||
          isAuthCall
        ) {
          throw error
        }
        original._retried = true
        try {
          const data = await refreshOnce()
          signIn(data)
        } catch {
          setAccessTokenC(null)
          throw error
        }
        return client(original)
      },
    )

    return () => {
      client.interceptors.request.eject(requestId)
      client.interceptors.response.eject(responseId)
    }
  }, [signIn, setAccessTokenC])

  const value = useMemo(
    () => ({ isAuthenticated, isInitializing, userId, signIn, logout }),
    [isAuthenticated, isInitializing, userId, signIn, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used inside <AuthProvider>')
  return context
}
