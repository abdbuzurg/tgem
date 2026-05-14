import { IMaterialCostView } from "@entities/material-cost/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function getMaterialCostByKey(key: number): Promise<IMaterialCostView> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IMaterialCostView>>(`/material-cost/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}