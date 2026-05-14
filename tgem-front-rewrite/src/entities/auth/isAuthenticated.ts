import IApiResponseFormat from "@shared/api/envelope"
import { USER_PATH } from "@shared/api/paths"
import axiosClient from "@shared/api/client"

export default async function isAuthenticated(): Promise<IApiResponseFormat<string>> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string>>(`/${USER_PATH}/is-authenticated`)
  const responseData = responseRaw.data
  return responseData
}