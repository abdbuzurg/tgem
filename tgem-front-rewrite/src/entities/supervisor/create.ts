import Tbl_Supervisor from "@entities/supervisor/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function createSupervisors(data: Tbl_Supervisor): Promise<Tbl_Supervisor> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<Tbl_Supervisor>>("/tbl_supervisor/", data)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}