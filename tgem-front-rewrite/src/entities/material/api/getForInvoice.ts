import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"

export interface MaterialForInvoice {
  key_material: number
  name_material: string
  unit: string
}

export async function getMaterialsForInvoice(): Promise<MaterialForInvoice[]>{
  const responseRaw = await axiosClient.get<IApiResponseFormat<MaterialForInvoice[]>>(`/material/warehouse`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}