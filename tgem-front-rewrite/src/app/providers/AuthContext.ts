import { createContext } from "react"
import { EffectivePermission } from "@entities/auth/permissions"

export interface IAuthContext {
  username: string
  effectivePermissions: EffectivePermission[]
  permissionsLoaded: boolean
  setUsername: (username: string) => void
  setEffectivePermissions: (permissions: EffectivePermission[]) => void
  // can answers "is this user allowed to do action on resource (in projectId, if scoped)?"
  // Pass projectId = undefined for non-project-scoped routes (admin, auctions);
  // only global grants then apply.
  can: (action: string, resource: string, projectId?: number) => boolean
  hasAnyPermission: () => boolean
  clearContext: () => void
}

export const AuthContext = createContext<IAuthContext>({
  effectivePermissions: [],
  permissionsLoaded: false,
  setEffectivePermissions: () => {},
  username: "",
  setUsername: () => {},
  can: () => false,
  hasAnyPermission: () => false,
  clearContext: () => {},
})
