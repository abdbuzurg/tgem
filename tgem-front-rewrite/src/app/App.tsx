import { Navigate, Route, Routes } from "react-router-dom"
import { ADMIN_PAGES, PAGES_WITHOUT_LAYOUT, PAGES_WITH_LAYOUT, AppRoute } from "@routes/index"
import AppLayout from "@app/layouts/AppLayout"
import AdminLayout from "@app/layouts/AdminLayout"
import RequireAuth from "@routes/guards/RequireAuth"
import Require from "@routes/guards/Require"

// renderGated wraps a route element in the guard prescribed by its `gate`
// field. Phase 3+ uses per-route gating; the previous "one RequirePermission
// wraps every PAGES_WITH_LAYOUT route" pattern is gone.
function renderGated(route: AppRoute, key: number) {
  if (route.gate.kind === "auth") {
    return (
      <Route key={key} element={<RequireAuth />}>
        <Route path={route.path} element={route.element} />
      </Route>
    )
  }
  return (
    <Route key={key} element={<Require action={route.gate.action} resource={route.gate.resource} />}>
      <Route path={route.path} element={route.element} />
    </Route>
  )
}

export default function App() {
  return (
    <Routes>
      {/* public routes */}
      {PAGES_WITHOUT_LAYOUT.map((route, index) => <Route key={index} path={route.path} element={route.element} />)}

      {/* logged-in routes, gated per-route */}
      <Route element={<AppLayout />}>
        {PAGES_WITH_LAYOUT.map(renderGated)}
      </Route>

      {/* admin layout, gated per-route */}
      <Route element={<AdminLayout />}>
        {ADMIN_PAGES.map(renderGated)}
      </Route>

      {/* 404 */}
      <Route path="*" element={<Navigate to="/404" replace />} />
    </Routes>
  )
}
