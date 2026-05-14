import { IMaterialCost } from "@entities/material-cost/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function getMaterailCostByMaterialID(materialID: number): Promise<IMaterialCost[]> {
  if (materialID == 0) {
    return []
  }
  const responseRaw = await axiosClient.get<IApiResponseFormat<IMaterialCost[]>>(`/material-cost/material-id/${materialID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  }  else {
    throw new Error(response.error)
  }
}