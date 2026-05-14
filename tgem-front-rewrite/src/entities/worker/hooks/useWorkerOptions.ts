import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { getWorkerByJobTitle } from "../api"
import IWorker from "../types"

export default function useWorkerOptions(jobTitle: string): IReactSelectOptions<number>[] {
  const { data } = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["worker", jobTitle],
    queryFn: () => getWorkerByJobTitle(jobTitle),
  })
  return useMemo(
    () => data?.map((w) => ({ label: w.name, value: w.id })) ?? [],
    [data],
  )
}
