import { Permission } from "./types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

const URL = "/permission"

export async function getPermissionsByRole(roleID: number):Promise<Permission[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Permission[]>>(`${URL}/role/${roleID}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface UserPermission{
  resourceName: string
  resourceUrl: string
  r: boolean
  w: boolean
  u: boolean
  d: boolean
}
export async function getPermissionsByRoleName(roleName: string):Promise<UserPermission[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<UserPermission[]>>(`${URL}/role/name/${roleName}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function createPermissions(data: Permission[]): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/batch`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getPermissionByResourceURL(resourceURL: string): Promise<boolean> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<boolean>>(`${URL}/role/url/${resourceURL}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}
