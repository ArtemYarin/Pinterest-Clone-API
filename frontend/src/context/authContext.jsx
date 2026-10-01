import { createContext, useRef, useState, useCallback, useContext } from 'react'

const AuthContext = createContext(null)

// Saves jwt token in ref. Gives access to methods: getAccessToken,
// setAccessToken, isAuthenticated, logout.
export function AuthProvider({ children }) {
  const tokenRef = useRef(null)
  const [isAuthenticated, setIsAuthenticated] = useState(false)

  const setAccessToken = useCallback((token) => {
    tokenRef.current = token
    setIsAuthenticated(!!token)
  }, [])

  const getAccessToken = useCallback(() => tokenRef.current, [])

  const logout = useCallback(() => {
    tokenRef.current = null
    setIsAuthenticated(false)
  }, [])

  return (
    <AuthContext.Provider
      value={{ getAccessToken, setAccessToken, isAuthenticated, logout }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
