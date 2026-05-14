import Tbl_Supervisor from "@entities/supervisor/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";

export default async function getAllSupervisors() {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Tbl_Supervisor[]>>("/tbl_supervisor/")
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}