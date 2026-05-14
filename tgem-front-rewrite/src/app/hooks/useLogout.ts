import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "react-router-dom"
import toast from "react-hot-toast"
import { LOGIN } from "@routes/paths"

// Logging out must wipe the react-query cache. Query keys don't include
// the projectID, so without clear() the next login (often into a different
// project) renders the previous project's cached rows for the first paint
// of every list. See MIGRATION/06-project-scoping-investigation.md §5.
export default function useLogout() {
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  return () => {
    const loadingToast = toast.loading("Выход.....")
    queryClient.clear()
    localStorage.removeItem("token")
    localStorage.removeItem("username")
    toast.dismiss(loadingToast)
    toast.success("Операция успешна")
    navigate(LOGIN)
  }
}
