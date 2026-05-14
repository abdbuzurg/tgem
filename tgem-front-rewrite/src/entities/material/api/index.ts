import fileDownload from "js-file-download"
import axiosClient from "@shared/api/client"
import IApiResponseFormat from "@shared/api/envelope"
import isCorrectResponseFormat from "@shared/lib/typeGuardForResponse"

const URL = "/material"

export async function getMaterialTemplateDocument(): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, "Шаблон для импорта материалов.xlsx")
    return true
  } else {
    throw new Error(responseRaw.statusText)
  }
}

export async function importMaterials(data: File): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data)
  const responseRaw = await axiosClient.post<IApiResponseFormat<null>>(`${URL}/document/import`, formData)
  const response = responseRaw.data
  if (response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function exportMaterials(): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (isCorrectResponseFormat<null>(responseRaw.data)) {
    const response = responseRaw.data as IApiResponseFormat<null>
    throw new Error(response.error)
  } else {
    if (responseRaw.status == 200) {
      fileDownload(responseRaw.data, "Эспорт Материалов.xlsx")
      return true
    } else {
      throw new Error(responseRaw.statusText)
    }
  }
}

export interface MaterialSearchParameters {
  name: string
  category: string
  code: string
  unit: string
}
