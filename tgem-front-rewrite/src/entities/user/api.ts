import IUser from "./types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";
import { ENTRY_LIMIT } from "@shared/config/pagination"

const URL = "/user"

export interface UserView {
  id: number
  username: string
  workerName: string
  workerMobileNumber: string
  workerJobTitle: string
  roleName: string
  accessToProjects: string[]
}

export interface UserPaginated {
  data: UserView[]
  count: number
  page: number
}

export async function getPaginatedUser({ pageParam = 1 }): Promise<UserPaginated> {
  const responseraw = await axiosClient.get<IApiResponseFormat<UserPaginated>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}`)
  const response = responseraw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export interface NewUserData {
  userData: IUser,
  projects: number[]
}

export async function createUser(data: NewUserData): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function updateUser(data: NewUserData): Promise<boolean> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<boolean>>(`${URL}/${data.userData.id}`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}
