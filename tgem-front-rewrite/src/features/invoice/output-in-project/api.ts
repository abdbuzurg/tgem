import fileDownload from "js-file-download"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { InvoiceMaterialViewWithSerialNumbers, InvoiceMaterialViewWithoutSerialNumbers } from "@entities/invoice/types/invoiceMaterial"
import { IInvoiceOutputInProject, IInvoiceOutputInProjectView, IInvoiceOutputMaterials } from "./types"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/output"

export interface InvoiceOutputInProjectPagianted {
  data: IInvoiceOutputInProjectView[]
  count: number
  page: number
}

export async function getPaginatedInvoiceOutputInProject({ pageParam = 1 }): Promise<InvoiceOutputInProjectPagianted> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceOutputInProjectPagianted>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export async function deleteInvoiceOutputInProject(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<string>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceOutputItem {
  materialID: number
  amount: number
  serialNumbers: string[]
  notes: string
}

export interface InvoiceOutputInProjectMutation {
  details: IInvoiceOutputInProject
  items: InvoiceOutputItem[]
}

export async function createInvoiceOutputInProject(data: InvoiceOutputInProjectMutation): Promise<InvoiceOutputInProjectMutation> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<InvoiceOutputInProjectMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function updateInvoiceOutputInProject(data: InvoiceOutputInProjectMutation): Promise<InvoiceOutputInProjectMutation> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<InvoiceOutputInProjectMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceOutputInProjectDocument(deliveryCode: string, confirmation: boolean): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/${deliveryCode}`, { responseType: "blob" })
  if (responseRaw.status == 200) {
    if (confirmation) {
      fileDownload(responseRaw.data, `${deliveryCode}.pdf`)
    } else {
      fileDownload(responseRaw.data, `${deliveryCode}.xlsx`)
    }
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export interface InvoiceOutputInProjectConfirmation {
  id: number
  file: File
}

export async function sendInvoiceOutputInProjectConfirmationExcel(data: InvoiceOutputInProjectConfirmation): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data.file)
  const responseRaw = await axiosClient.post(`${URL}/confirm/${data.id}`, formData)
  if (responseRaw.data.success && responseRaw.data.permission) {
    if (typeof responseRaw.data == 'object') {
      const response: IApiResponseFormat<string> = responseRaw.data
      if (response.success && response.permission) {
        return true
      } else {
        throw new Error(response.error)
      }
    }

    return true
  } else {
    throw new Error(responseRaw.data.error)
  }
}

export async function getAllUniqueCode(): Promise<IReactSelectOptions<string>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<string>[]>>(`${URL}/unique/code`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllUniqueDistrict(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/district`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllUniqueWarehouseManager(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/warehouse-manager`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllUniqueRecieved(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/recieved`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllUniqueTeam(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/team`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceOutputReportFilter {
  code: string
  warehouseManagerID: number
  recievedID: number
  districtID: number
  teamID: number
  dateFrom: Date | null
  dateTo: Date | null
}

export async function buildReport(filter: InvoiceOutputReportFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report`, filter, { responseType: "blob", })
  if (responseRaw.status == 200) {
    const date = new Date()
    const fileName = `Отчет Накладной Отпуск - ${date}`
    fileDownload(responseRaw.data, `${fileName}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function getSerialNumberCodesByMaterialID(materialID: number): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/serial-number/material/${materialID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface AvailableMaterial {
  id: number
  name: string
  unit: string
  hasSerialNumber: boolean
  amount: number
}

export async function getAvailableMaterialsInWarehouse(): Promise<AvailableMaterial[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<AvailableMaterial[]>>(`${URL}/material/available-in-warehouse`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceOutputMaterilsWithoutSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithoutSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithoutSerialNumbers[]>>(`${URL}/${id}/materials/without-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceOutputMaterilsWithSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithSerialNumbers[]>>(`${URL}/${id}/materials/with-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceOutputInProjectMaterialsForEdit(id: number): Promise<IInvoiceOutputMaterials[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IInvoiceOutputMaterials[]>>(`${URL}/invoice-materials/${id}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function importInvoiceOutput(file: File): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", file)
  const responseRaw = await axiosClient.post<IApiResponseFormat<unknown>>(`${URL}/import`, formData)
  const response = responseRaw.data
  if (response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}


