import fileDownload from "js-file-download"
import { IObject } from "@entities/object/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/tp"

export interface ITPObjectPaginated {
  objectID: number
  objectDetailedID: number
  name: string
  status: string
  model: string
  voltageClass: string
  nourashes: string
  supervisors: string[]
  teams: string[]
}

export interface ITPObjectGetAllResponse {
  data: ITPObjectPaginated[]
  count: number
  page: number
}

export interface TPObjectSearchParameters {
  objectName: string
  teamID: number
  supervisorWorkerID: number
}

export async function getPaginatedTPObjects({ pageParam = 1 }, searchParameters: TPObjectSearchParameters): Promise<ITPObjectGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<ITPObjectGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&teamID=${searchParameters.teamID}&supervisorWorkerID=${searchParameters.supervisorWorkerID}&objectName=${searchParameters.objectName}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export interface ITPObjectCreate {
  baseInfo: IObject
  detailedInfo: {
    model: string
    voltageClass: string
    nourashes: string
  }
  supervisors: number[]
  teams: number[]
}

export async function createTPObject(data: ITPObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function updateTPObject(data: ITPObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function deleteTPObject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getTPTemplateDocument(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Шаблон для импорта объектов - ТП.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function importTP(data: File): Promise<boolean> {
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

export async function getAllTPs(): Promise<IObject[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IObject[]>>(`${URL}/all`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getTPObjectNames(): Promise<IReactSelectOptions<string>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<string>[]>>(`${URL}/search/object-names`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function exportTP(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Экспорт МЖД.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

