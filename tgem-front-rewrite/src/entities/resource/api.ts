import Resource from "./types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"

const URL = '/resource'

export async function getAllResources(): Promise<Resource[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Resource[]>>(`${URL}/`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else{
    throw new Error(response.error)
  }
}
