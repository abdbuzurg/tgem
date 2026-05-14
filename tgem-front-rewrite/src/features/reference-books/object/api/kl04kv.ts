import fileDownload from "js-file-download"
import { IObject } from "@entities/object/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/kl04kv"

export interface KL04KVSearchParameters {
  objectName: string
  teamID: number
  supervisorWorkerID: number
  tpObjectID: number
}

export interface IKL04KVObjectPaginated {
  objectID: number
  objectDetailedID: number
  name: string
  status: string
  length: number
  nourashes: string
  supervisors: string[]
  teams: string[]
  tpNames: string[]
}

export interface IKL04KVObjectGetAllResponse {
  data: IKL04KVObjectPaginated[]
  count: number
  page: number
}

export async function getPaginatedKL04KVObjects({ pageParam = 1 }, searchParameters: KL04KVSearchParameters): Promise<IKL04KVObjectGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IKL04KVObjectGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&teamID=${searchParameters.teamID}&supervisorWorkerID=${searchParameters.supervisorWorkerID}&tpObjectID=${searchParameters.tpObjectID}&objectName=${searchParameters.objectName}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export interface IKL04KVObjectCreate {
  baseInfo: IObject
  detailedInfo: {
    length: number
    nourashes: string
  }
  supervisors: number[]
  teams: number[]
  nourashedByTP: number[]
}

export async function createKL04KVObject(data: IKL04KVObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function updateKL04KVObject(data: IKL04KVObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function deleteKL04KVObject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getKL04KVTemplateDocument(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Шаблон для импорта объектов - КЛ 04 КВ.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function importKL04KV(data: File): Promise<boolean> {
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

export async function exportKL04KV(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Экспорт КЛ 04 КВ.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function getKL04KVObjectNames(): Promise<IReactSelectOptions<string>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<string>[]>>(`${URL}/search/object-names`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }

}
