import Tbl_Supervisor from "@entities/supervisor/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function getByKeySupervisor(key: number):Promise<Tbl_Supervisor> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Tbl_Supervisor>>(`/tbl_supervisor/${key}`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}