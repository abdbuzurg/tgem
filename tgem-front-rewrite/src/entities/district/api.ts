import { IDistrict } from "./types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";
import { ENTRY_LIMIT } from "@shared/config/pagination";

export async function getAllDistricts():Promise<IDistrict[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IDistrict[]>>("/district/all")
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface DistrictsPaginatedData {
  data: IDistrict[]
  count: number
  page: number
}

export async function getDistrictsPaginated({pageParam = 1}): Promise<DistrictsPaginatedData> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<DistrictsPaginatedData>>(`/district/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return {...response.data, page: pageParam}
  } else {
    throw new Error(response.error)
  }
}

export async function createDistrict(data: IDistrict): Promise<IDistrict>{
  const responseRaw = await axiosClient.post<IApiResponseFormat<IDistrict>>(`/district/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function updateDistrict(data: IDistrict): Promise<IDistrict>{
  const responseRaw = await axiosClient.patch<IApiResponseFormat<IDistrict>>(`/district/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function deleteDistrict(id: number): Promise<boolean>{
  const responseRaw = await axiosClient.delete<IApiResponseFormat<string>>(`/district/${id}`, )
  const response = responseRaw.data
  if (response.permission && response.success && response.data == "deleted") {
    return true
  } else {
    throw new Error(response.error)
  }
}



