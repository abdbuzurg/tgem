import axiosClient from "@shared/api/client";

export default async function deleteMaterial(key: number) {
  const responseRaw = await axiosClient.delete(`/material/${key}`)
  const response = responseRaw.data
  return response.success 
}