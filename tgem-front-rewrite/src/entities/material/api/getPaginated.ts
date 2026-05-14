import Material from "@entities/material/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { MaterialSearchParameters } from "@entities/material/api"

export interface MaterialsGetAllResponse {
  data: Material[]
  count: number
  page: number
}


export default async function getPaginatedMaterials({pageParam = 1}, searchParameters: MaterialSearchParameters): Promise<MaterialsGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<MaterialsGetAllResponse>>(`/material/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&name=${searchParameters.name}&category=${searchParameters.category}&code=${searchParameters.code}&unit=${searchParameters.unit}`)
  const responseData = responseRaw.data
  if (responseData.success) {
    return {...responseData.data, page: pageParam}
  } else {
    throw new Error(responseData.error)
  }
}
