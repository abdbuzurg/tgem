import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { IRole } from "./types"

const URL = "/role"

export async function getAllRoles(): Promise<IRole[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IRole[]>>(`${URL}/all`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function createRole(data: IRole):Promise<IRole> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<IRole>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}
