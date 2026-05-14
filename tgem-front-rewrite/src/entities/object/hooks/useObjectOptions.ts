import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { getAllUniqueObjects } from "../api"
import { ObjectDataForSelect } from "../types"

export default function useObjectOptions(): IReactSelectOptions<number>[] {
  const { data } = useQuery<ObjectDataForSelect[], Error, ObjectDataForSelect[]>({
    queryKey: ["all-unique-objects"],
    queryFn: getAllUniqueObjects,
  })
  return useMemo(
    () => data?.map((o) => ({ label: `${o.objectName} (${o.objectType})`, value: o.id })) ?? [],
    [data],
  )
}
