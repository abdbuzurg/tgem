import Tbl_Supervisor from "@entities/supervisor/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"

export interface SupervisorsGetAllResponse {
  data: Tbl_Supervisor[]
  count: number
  page: number
}


export default async function getAllSupervisorPaginated({pageParam = 1}): Promise<SupervisorsGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<SupervisorsGetAllResponse>>(`/tbl_supervisor/all?page=${pageParam}&limit=${ENTRY_LIMIT}`)
  const responseData = responseRaw.data
  if (responseData.success) {
    return {...responseData.data, page: pageParam}
  } else {
    throw new Error(responseData.error)
  }
}