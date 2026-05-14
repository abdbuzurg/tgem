import { IMaterialCostView } from "@entities/material-cost/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { MaterialCostSearchParameteres } from "@entities/material-cost/api"

export interface MaterialCostGetAllResponse {
  data: IMaterialCostView[]
  count: number
  page: number
}

export default async function getPaginatedMaterialCost({pageParam = 1}, searchParameters: MaterialCostSearchParameteres): Promise<MaterialCostGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<MaterialCostGetAllResponse>>(`/material-cost/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&materialName=${searchParameters.materialName}`)
  const responseData = responseRaw.data
  if (responseData.success) {
    return {...responseData.data, page: pageParam}
  } else {
    throw new Error(responseData.error)
  }
}
