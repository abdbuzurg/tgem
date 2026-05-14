import Material from "@entities/material/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function getMaterialByKey(key: number): Promise<Material> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Material>>(`/material/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}