import fileDownload from "js-file-download"
import IWorker from "./types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import isCorrectResponseFormat from "@shared/lib/typeGuardForResponse"

const URL = "/worker"

export async function getWorkerByJobTitle(jobTitle: string): Promise<IWorker[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IWorker[]>>(`${URL}/job-title/${jobTitle}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function createWorker(data: IWorker): Promise<IWorker> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<IWorker>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function deleteWorker(key: string) {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<IWorker[]>>(`${URL}/${key}`)
  const response = responseRaw.data
  return response.success
}

export async function getAllWorkers(): Promise<IWorker[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IWorker[]>>(`${URL}/all`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface WorkerSearchParameters {
  name: string
  jobTitleInCompany: string
  companyWorkerID: string
  jobTitleInProject: string
  mobileNumber: string
}

export interface WorkerPaginatedData {
  data: IWorker[]
  count: number
  page: number
}

export async function getPaginatedWorker({ pageParam = 1 }, searchParameters: WorkerSearchParameters): Promise<WorkerPaginatedData> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<WorkerPaginatedData>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&name=${searchParameters.name}&jobTitleInCompany=${searchParameters.jobTitleInCompany}&jobTitleInProject=${searchParameters.jobTitleInProject}&companyWorkerID=${searchParameters.companyWorkerID}&mobileNumber=${searchParameters.mobileNumber}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export async function updateWorker(data: IWorker): Promise<IWorker> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<IWorker>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getWorkerTemplateDocument(): Promise<boolean> {
  const response = await axiosClient.get(`${URL}/document/template`, { responseType: "blob" })
  if (response.status == 200) {
    fileDownload(response.data, "Шаблон для импорта Рабочего Персонала.xlsx")
    return true
  } else {
    throw new Error(response.statusText)
  }
}

export async function importWorker(data: File): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data)
  const responseRaw = await axiosClient.post(`${URL}/document/import`, formData)
  if (responseRaw.status == 200) {
    if (typeof responseRaw.data == 'object') {
      const response: IApiResponseFormat<string> = responseRaw.data
      if (!response.success) {
        throw new Error(response.error)
      } else {
        return true
      }
    } else {
      return true
    }
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function exportWorker(): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/export`, { responseType: "blob" })
  if (isCorrectResponseFormat<null>(responseRaw.data)) {
    const response = responseRaw.data as IApiResponseFormat<null>
    throw new Error(response.error)
  } else {
    if (responseRaw.status == 200) {
      fileDownload(responseRaw.data, "Эспорт Рабочего Персонала.xlsx")
      return true
    } else {
      throw new Error(responseRaw.statusText)
    }
  }

}

export interface WorkerInformation {
  name: string[]
  jobTitleInCompany: string[]
  companyWorkerID: string[]
  jobTitleInProject: string[]
  mobileNumber: string[]
}

export async function getWorkerInformation(): Promise<WorkerInformation> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<WorkerInformation>>(`${URL}/unique/worker-information`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }

}
