import { IMaterialCost } from "@entities/material-cost/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function createMaterialCost(data: IMaterialCost): Promise<IMaterialCost> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<IMaterialCost>>(`/material-cost/`, data)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}