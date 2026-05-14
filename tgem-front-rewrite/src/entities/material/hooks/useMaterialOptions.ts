import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import getAllMaterials from "../api/getAll"
import Material from "../types"

export default function useMaterialOptions(): IReactSelectOptions<number>[] {
  const { data } = useQuery<Material[], Error, Material[]>({
    queryKey: ["all-materials"],
    queryFn: getAllMaterials,
  })
  return useMemo(
    () => data?.map((m) => ({ label: m.name, value: m.id })) ?? [],
    [data],
  )
}
