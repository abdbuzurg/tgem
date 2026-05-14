import fileDownload from "js-file-download"
import axiosClient from "@shared/api/client"

const URL = "/material-location"

export interface ReportBalanceFilter {
  type: string
  teamID: number
  objectID: number
}

export async function buildReportBalance(filter: ReportBalanceFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report/balance`, filter, { responseType: "blob", })
  if (responseRaw.status == 200) {
    const fileName = "Отчет Остатка"
    fileDownload(responseRaw.data, `${fileName}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}
