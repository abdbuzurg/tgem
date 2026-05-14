import { IObject, ObjectDataForSelect } from "./types"
import { TeamDataForSelect } from "@entities/team/types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"

const URL = "/object"

export interface ObjectCreateShape {
  id: number,
  name: string,
  objectDetailedID: number,
  status: string,
  type: string, 
  model: string,
  amountStores: number,
  amountEntrances: number,
  hasBasement: boolean,
  voltageClass: string,
  nourashes: string,
  ttCoefficient: string,
  amountFeeders: number,
  length: number,
  supervisors: number[]
}

export async function createObject(data: ObjectCreateShape): Promise<IObject> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<IObject>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function deleteObject(key: number) {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`${URL}/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.success
  } else {
    throw new Error(response.error)
  }
}

export async function getAllObjects(): Promise<IObject[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IObject[]>>(`${URL}/all`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface IObjectPaginated {
  id: number
  name: string
  type: string
  status: string
  supervisors: string[]
}

export interface IObjectGetAllResponse {
  data: IObjectPaginated[]
  count: number
  page: number
}

export async function getPaginatedObjects({ pageParam = 1 }): Promise<IObjectGetAllResponse> {
  console.log(pageParam)
  const responseRaw = await axiosClient.get<IApiResponseFormat<IObjectGetAllResponse>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return {...response.data, page: pageParam}
  } else {
    throw new Error(response.error)
  }
}

export async function updateObject(data: IObject): Promise<IObject> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<IObject>>("{URL}/", data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getTeamsByObjectID(objectID: number): Promise<TeamDataForSelect[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<TeamDataForSelect[]>>(`${URL}/teams/${objectID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllUniqueObjects(): Promise<ObjectDataForSelect[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<ObjectDataForSelect[]>>(`/material-location/unique/object`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}
