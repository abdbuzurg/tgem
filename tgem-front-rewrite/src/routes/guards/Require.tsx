import { Navigate, Outlet } from "react-router-dom"
import useAuth from "@app/hooks/useAuth"
import LoadingDots from "@shared/ui/LoadingDots"
import { LOGIN, PERMISSION_DENIED } from "@routes/paths"

interface Props {
  action: string
  resource: string
}

// Require — declarative v2 permission gate. Pass the action and resource type
// the route needs; the user must have a matching grant either globally or in
// the active project. The backend resolver applies the same check on the
// server side (middleware.RequirePermission), so this is a UX hint mirror —
// the source of truth is the backend.
export default function Require({ action, resource }: Props) {
  const { permissionsLoaded, can } = useAuth()

  if (!permissionsLoaded) {
    return (
      <div className="w-screen h-screen text-center">
        <LoadingDots height={120} width={120} />
      </div>
    )
  }

  if (!localStorage.getItem("token")) {
    return <Navigate to={LOGIN} replace />
  }

  if (!can(action, resource)) {
    return <Navigate to={PERMISSION_DENIED} replace />
  }

  return <Outlet />
}
