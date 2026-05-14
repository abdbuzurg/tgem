import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { getAllTeamsForSelect } from "../api"
import { TeamDataForSelect } from "../types"

// Returns teams formatted as "<number> (<leader>)" — the canonical way teams
// are presented in selects across the app.
export default function useTeamOptions(): IReactSelectOptions<number>[] {
  const { data } = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["all-teams-for-select"],
    queryFn: getAllTeamsForSelect,
  })
  return useMemo(
    () => data?.map((t) => ({ label: `${t.teamNumber} (${t.teamLeaderName})`, value: t.id })) ?? [],
    [data],
  )
}
