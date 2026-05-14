import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function deleteSupervisor(key: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<undefined>>(`/tbl_supervisor/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}