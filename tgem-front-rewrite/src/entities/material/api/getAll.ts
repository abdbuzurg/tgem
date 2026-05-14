import Material from "@entities/material/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function getAllMaterials(): Promise<Material[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Material[]>>("/material/all")
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}