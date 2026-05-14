import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

interface UniqueColumnData {
  kod_material: string[]
  name_material: string[]
  cat_material: string[]
}

export default async function getMaterialCategories():Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<UniqueColumnData>>("/material/unique-column-data")
  const response = responseRaw.data
  if (response.success) {
    return response.data.cat_material
  } else {
    throw new Error(response.error)
  }
}