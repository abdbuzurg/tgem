import fileDownload from "js-file-download"
import { IOperation } from "./types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"

const URL = "/operation"

export interface OperationMutation {
  id: number
  projectID: number
  name: string
  code: string
  costPrime: number
  costWithCustomer: number
  materialID: number
  plannedAmountForProject: number
  showPlannedAmountInReport: boolean
}

export async function createOperation(data: OperationMutation): Promise<IOperation> {
  console.log(data)
  const responseRaw = await axiosClient.post<IApiResponseFormat<IOperation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function deleteOperation(key: number) {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.success
  } else {
    throw new Error(response.error)
  }
}

export async function getAllOperations(): Promise<OperationPaginated[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<OperationPaginated[]>>(`${URL}/all`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface OperationGetAllResponse {
  data: OperationPaginated[]
  count: number
  page: number
}

export interface OperationPaginated {
  id: number
  projectID: number
  name: string
  code: string
  costPrime: number
  costWithCustomer: number
  materialID: number
  materialName: string
  showPlannedAmountInReport: boolean
  plannedAmountForProject: number
}

export interface OperationSearchParameters {
  name: string
  code: string
  materialID: number
}

export async function getPaginatedOperations({ pageParam = 1 }, searchParameters: OperationSearchParameters): Promise<OperationGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<OperationGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&name=${searchParameters.name}&code=${searchParameters.code}&materialID=${searchParameters.materialID}`)
  const responseData = responseRaw.data
  if (responseData.success) {
    return { ...responseData.data, page: pageParam }
  } else {
    throw new Error(responseData.error)
  }
}

export async function getOperationByKey(key: number): Promise<IOperation> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IOperation>>(`${URL}/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export default async function updateOperation(data: OperationMutation): Promise<IOperation> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<IOperation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getOperationTemplateDocument(): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, "Шаблон для импорта Услуг.xlsx")
    return true
  } else {
    throw new Error(responseRaw.statusText)
  }
}

export async function importOperations(data: File): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data)
  const responseRaw = await axiosClient.post<IApiResponseFormat<null>>(`${URL}/document/import`, formData)
  const response = responseRaw.data
  if (response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}
