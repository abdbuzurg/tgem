import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"

const URL = "/worker-attendance"

export async function createWorkerAttendance(data: File): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data)
  const responseRaw = await axiosClient.post<IApiResponseFormat<null>>(`${URL}/`, formData)
  const response = responseRaw.data
  if (response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export interface WorkerAttendancePaginated {
  id: number
  workerName: string
  companyWorkerID: string
  start: string
  end: string
}

export interface WorkerAttendancePaginatedResponse {
  data: WorkerAttendancePaginated[]
  count: number
  page: number
}

export default async function getPaginatedWorkerAttendance({pageParam = 1}): Promise<WorkerAttendancePaginatedResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<WorkerAttendancePaginatedResponse>>(`${URL}/paginated`)
  const responseData = responseRaw.data
  if (responseData.success) {
    return {...responseData.data, page: pageParam}
  } else {
    throw new Error(responseData.error)
  }
}
