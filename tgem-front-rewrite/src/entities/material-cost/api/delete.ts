import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function deleteMaterialCost(key: number) {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<boolean>>(`/material-cost/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.success
  } else {
    throw new Error(response.error)
  }
}