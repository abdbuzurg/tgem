import { useMemo } from "react"
import { useQuery } from "@tanstack/react-query"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { getAllDistricts } from "../api"
import { IDistrict } from "../types"

export default function useDistrictOptions(): IReactSelectOptions<number>[] {
  const { data } = useQuery<IDistrict[], Error, IDistrict[]>({
    queryKey: ["all-districts"],
    queryFn: getAllDistricts,
  })
  return useMemo(
    () => data?.map((d) => ({ label: d.name, value: d.id })) ?? [],
    [data],
  )
}
