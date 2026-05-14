import Tbl_Supervisor from "@entities/supervisor/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function updateSupervisor(data: Tbl_Supervisor): Promise<Tbl_Supervisor> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<Tbl_Supervisor>>("/tbl_supervisor/", data)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}