import { ReactNode, useCallback, useEffect, useMemo, useState } from "react"
import { AuthContext } from "./AuthContext"
import { EffectivePermission, getEffectivePermissions } from "@entities/auth/permissions"

interface Props {
  children: ReactNode
}

export default function AuthProvider({ children }: Props) {
  const [username, setUsername] = useState<string>(() => localStorage.getItem("username") ?? "")
  const [effectivePermissions, setEffectivePermissionsState] = useState<EffectivePermission[]>([])
  const [permissionsLoaded, setPermissionsLoaded] = useState<boolean>(false)

  useEffect(() => {
    const token = localStorage.getItem("token")
    if (!token) {
      setPermissionsLoaded(true)
      return
    }
    getEffectivePermissions()
      .then(setEffectivePermissionsState)
      .catch(() => setEffectivePermissionsState([]))
      .finally(() => setPermissionsLoaded(true))
  }, [])

  // All callbacks are wrapped in useCallback so the AuthContext value is stable
  // across re-renders. Without this, the value object is recreated each render
  // and any consumer that puts authContext in a useEffect dep array would fire
  // its effect repeatedly — that's how clearContext kept wiping perms in the
  // login flow.
  const setEffectivePermissions = useCallback((p: EffectivePermission[]) => {
    setEffectivePermissionsState(p)
  }, [])

  // can() — the v2 permission check. A grant matches when:
  //   - resourceType + action match exactly, AND
  //   - the grant is global (projectId === null) OR matches the requested project.
  // Pass projectId = undefined for non-project-scoped routes; only globals apply.
  const can = useCallback((action: string, resource: string, projectId?: number): boolean => {
    return effectivePermissions.some((p) => {
      if (p.resourceType !== resource || p.action !== action) return false
      if (p.projectId === null) return true
      return projectId !== undefined && p.projectId === projectId
    })
  }, [effectivePermissions])

  const hasAnyPermission = useCallback(() => effectivePermissions.length > 0, [effectivePermissions])

  const clearContext = useCallback(() => {
    setEffectivePermissionsState([])
    setUsername("")
    setPermissionsLoaded(true)
  }, [])

  const value = useMemo(() => ({
    username,
    effectivePermissions,
    permissionsLoaded,
    setUsername,
    setEffectivePermissions,
    can,
    hasAnyPermission,
    clearContext,
  }), [username, effectivePermissions, permissionsLoaded, setEffectivePermissions, can, hasAnyPermission, clearContext])

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}
