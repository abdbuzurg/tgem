import fileDownload from "js-file-download"
import { InvoiceOutputOutOfProject } from "./types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { InvoiceOutputItem } from "@features/invoice/output-in-project/api"
import { InvoiceMaterialViewWithSerialNumbers, InvoiceMaterialViewWithoutSerialNumbers } from "@entities/invoice/types/invoiceMaterial"
import { IInvoiceOutputMaterials } from "@features/invoice/output-in-project/types"

const URL = "/invoice-output-out-of-project"

export interface InvoiceOutputOutOfProjectMutation {
  details: InvoiceOutputOutOfProject
  items: InvoiceOutputItem[]
}

export async function createInvoiceOutputOfOutProject(data: InvoiceOutputOutOfProjectMutation): Promise<InvoiceOutputOutOfProjectMutation> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<InvoiceOutputOutOfProjectMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceOutputOutOfProjectView {
  id: number
  nameOfProject: string
  deliveryCode: string
  releasedWorkerName: string
  dateOfInvoice: Date
  confirmation: boolean
}

export interface InvoiceOutputOutOfProjectPaginated {
  data: InvoiceOutputOutOfProjectView[]
  count: number
  page: number
}

export interface InvoiceOutputOutOfProjectSearchParameters {
  nameOfProject: string
  releasedWorkerID: number
}

export async function getInvoiceOutputOutOfProjectPaginated({ pageParam = 1 }, searchParameters: InvoiceOutputOutOfProjectSearchParameters): Promise<InvoiceOutputOutOfProjectPaginated> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceOutputOutOfProjectPaginated>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&nameOfProject=${searchParameters.nameOfProject}&releasedWorkerID=${searchParameters.releasedWorkerID}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }

}

export async function deleteInvoiceOutputOutOfProject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<string>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceOutputOutOfProjectConfirmation {
  id: number
  file: File
}

export async function sendInvoiceOutputOutOfProjectConfirmationExcel(data: InvoiceOutputOutOfProjectConfirmation): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data.file)
  const responseRaw = await axiosClient.post(`${URL}/confirm/${data.id}`, formData)
  if (responseRaw.status == 200) {
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function getInvoiceOutputOutOfProjectDocument(deliveryCode: string): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/${deliveryCode}`, { responseType: "blob" })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, `${deliveryCode}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function getInvoiceOutputOutOfProjectMaterilsWithoutSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithoutSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithoutSerialNumbers[]>>(`${URL}/${id}/materials/without-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceOutputOutOfProjectMaterilsWithSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithSerialNumbers[]>>(`${URL}/${id}/materials/with-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function updateInvoiceOutputOfOutProject(data: InvoiceOutputOutOfProjectMutation): Promise<InvoiceOutputOutOfProjectMutation> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<InvoiceOutputOutOfProjectMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceOutputOutOfProjectMaterialsForEdit(id: number): Promise<IInvoiceOutputMaterials[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IInvoiceOutputMaterials[]>>(`${URL}/invoice-materials/${id}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getUniqueNamesOfProjects(): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/unique/name-of-project`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceOutputOutOfProjectReportFilter {
  dateFrom: Date | null
  dateTo: Date | null
}

export async function buildReportInvoiceOutputOutOfProject(filter: InvoiceOutputOutOfProjectReportFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report`, filter, { responseType: "blob" })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, `Отсчет накладной отпуск вне проекта.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}
