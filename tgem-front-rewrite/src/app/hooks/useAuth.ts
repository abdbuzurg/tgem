import { useContext } from "react"
import { AuthContext } from "@app/providers/AuthContext"

export default function useAuth() {
  return useContext(AuthContext)
}
