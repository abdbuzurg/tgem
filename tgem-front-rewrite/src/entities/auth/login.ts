import IApiResponseFormat from "@shared/api/envelope"
import { USER_PATH } from "@shared/api/paths"
import axiosClient from "@shared/api/client"

export interface LoginRequestData{
  username: string
  password: string
  projectID: number
}

export interface LoginResponseData {
  token: string
  admin: boolean
}

export default async function loginUser(data: LoginRequestData): Promise<LoginResponseData> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<LoginResponseData>>(`/${USER_PATH}/login`, data)
  const responseData = responseRaw.data
  if (responseData.success) {
    return responseData.data
  } else {
    throw new Error(responseData.error)
  }
}
