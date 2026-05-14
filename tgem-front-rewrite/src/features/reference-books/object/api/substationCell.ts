import fileDownload from "js-file-download"
import { IObject } from "@entities/object/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/cell-substation"

export interface ISubstationCellObjectPaginated {
  objectID: number
  objectDetailedID: number
  name: string
  status: string
  supervisors: string[]
  teams: string[]
  substationName: string
}

export interface ISubstationCellObjectGetAllResponse {
  data: ISubstationCellObjectPaginated[]
  count: number
  page: number
}

export interface SubstationCellObjectSearchParameters {
  objectName: string
  teamID: number
  supervisorWorkerID: number
  substationObjectID: number
}

export async function getPaginatedSubstationCellObjects({ pageParam = 1 }, searchParameters: SubstationCellObjectSearchParameters): Promise<ISubstationCellObjectGetAllResponse> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<ISubstationCellObjectGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&teamID=${searchParameters.teamID}&supervisorWorkerID=${searchParameters.supervisorWorkerID}&objectName=${searchParameters.objectName}&substationObjectID=${searchParameters.substationObjectID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export interface ISubstationCellObjectCreate {
  baseInfo: IObject
  detailedInfo: {
    amountFeeders: number
  }
  supervisors: number[]
  teams: number[]
  substationObjectID: number
}

export async function createSubstationCellObject(data: ISubstationCellObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function updateSubstationCellObject(data: ISubstationCellObjectCreate): Promise<boolean> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function deleteSubstationCellObject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getSubstationCellTemplateDocument(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Шаблон для импорта объектов - Ячейки Подстанции.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function importSubstationCell(data: File): Promise<boolean> {
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


export async function getSubstationCellObjectNames(): Promise<IReactSelectOptions<string>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<string>[]>>(`${URL}/search/object-names`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}


export async function exportSubstationCell(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Экспорт Подстанции.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}
