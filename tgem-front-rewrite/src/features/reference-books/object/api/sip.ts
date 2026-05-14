import fileDownload from "js-file-download"
import { IObject } from "@entities/object/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/sip"

export interface ISIPObjectPaginated {
  objectID: number
  objectDetailedID: number
  name: string
  status: string
  amountFeeders: number
  supervisors: string[]
  teams: string[]
  tpNames: string[]
}

export interface ISIPObjectGetAllResponse {
  data: ISIPObjectPaginated[]
  count: number
  page: number
}

export interface SIPObjectSearchParameters {
  objectName: string
  teamID: number
  supervisorWorkerID: number
}

export async function getPaginatedSIPObjects({ pageParam = 1 }, searchParameters: SIPObjectSearchParameters): Promise<ISIPObjectGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<ISIPObjectGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&objectName=${searchParameters.objectName}&teamID=${searchParameters.teamID}&supervisorWorkerID=${searchParameters.supervisorWorkerID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export interface ISIPObjectCreate {
  baseInfo: IObject
  detailedInfo: {
    amountFeeders: number
  }
  supervisors: number[]
  teams: number[]
  nourashedByTP: number[]
}

export async function createSIPObject(data: ISIPObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function updateSIPObject(data: ISIPObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function deleteSIPObject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getSIPTemplateDocument(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Шаблон для импорта объектов - СИП.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function importSIP(data: File): Promise<boolean> {
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

export async function getSIPObjectNames(): Promise<IReactSelectOptions<string>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<string>[]>>(`${URL}/search/object-names`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }

}

export async function getTPNamesForSIP(): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/tp-names`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function exportSIP(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Экспорт МЖД.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}
