import { Navigate, Outlet } from "react-router-dom"
import useAuth from "@app/hooks/useAuth"
import LoadingDots from "@shared/ui/LoadingDots"
import { LOGIN } from "@routes/paths"

// RequireAuth — gate for menu pages, dashboards, and other routes that have
// no specific resource permission. The user must be logged in (have a token).
// Per the permissions spec §6, /home, /reference-book menu, /admin/home,
// /report use this. Inner items are individually gated with <Require>.
export default function RequireAuth() {
  const { permissionsLoaded } = useAuth()

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

  return <Outlet />
}
