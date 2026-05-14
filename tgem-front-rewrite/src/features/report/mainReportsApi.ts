import fileDownload from "js-file-download"
import axiosClient from "@shared/api/client"

const URL = "/main-reports"

export async function projectProgressReport(date: Date): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/project-progress`, {date: date}, { responseType: "blob", })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, `Прогресс Реализации Проекта ${date.getDate()}-${date.getMonth() + 1}-${date.getFullYear()}.xlsx`)
    return true
  } else {
    throw new Error("Неизветсная ошибка при создании отсчета")
  }
}

export async function analysisOfRemainingMaterials(): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/analysis-of-remaining-materials`, { responseType: "blob", })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, "Аналитика Остатка Материалов.xlsx")
    return true
  } else {
    throw new Error("Неизветсная ошибка при создании отсчета")
  }
}
