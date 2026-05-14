import fileDownload from "js-file-download"
import { IObject } from "@entities/object/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/stvt"

export interface ISTVTObjectPaginated {
  objectID: number
  objectDetailedID: number
  name: string
  status: string
  amountFeeders: number
  voltageClass: string
  ttCoefficient: string
  supervisors: string[]
  teams: string[]
}

export interface ISTVTObjectGetAllResponse {
  data: ISTVTObjectPaginated[]
  count: number
  page: number
}

export interface STVTSearchParameters {
  objectName: string
  teamID: number
  supervisorWorkerID: number
}

export async function getPaginatedSTVTObjects({ pageParam = 1 }, searchParameters: STVTSearchParameters): Promise<ISTVTObjectGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<ISTVTObjectGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&teamID=${searchParameters.teamID}&supervisorWorkerID=${searchParameters.supervisorWorkerID}&objectName=${searchParameters.objectName}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export interface ISTVTObjectCreate {
  baseInfo: IObject
  detailedInfo: {
    voltageClass: string
    ttCoefficient: string
  }
  supervisors: number[]
  teams: number[]
}

export async function createSTVTObject(data: ISTVTObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function updateSTVTObject(data: ISTVTObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function deleteSTVTObject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getSTVTTemplateDocument(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Шаблон для импорта объектов - СТВТ.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function importSTVT(data: File): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data)
  const responseRaw = await axiosClient.post(`${URL}/document/import`, formData)
  if (responseRaw.status == 200) {
    if (typeof responseRaw.data == 'object') {
      const response: IApiResponseFormat<string> = responseRaw.data
      if (!response.success) {
        throw new Error(response.error)
      } else {
        return true
      }
    } else {
      return true
    }
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function getSTVTObjectNames(): Promise<IReactSelectOptions<string>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<string>[]>>(`${URL}/search/object-names`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function exportSTVT(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Экспорт МЖД.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

