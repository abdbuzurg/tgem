import Material from "@entities/material/types";
import { IMaterialCost } from "@entities/material-cost/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export interface FullMaterailData {
  details: Material
  cost: IMaterialCost
}

export default async function createFullMaterial(data: FullMaterailData):Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>("/material-cost/full-material", data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}