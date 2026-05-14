import Material from "@entities/material/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function createMaterial(data: Material): Promise<Material> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<Material>>(`/material/`, data)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}