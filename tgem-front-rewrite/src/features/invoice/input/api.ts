import fileDownload from "js-file-download"
import { IInvoiceInput, IInvoiceInputMaterials, IInvoiceInputView } from "./types"
import { InvoiceMaterial, InvoiceMaterialViewWithSerialNumbers, InvoiceMaterialViewWithoutSerialNumbers, } from "@entities/invoice/types/invoiceMaterial"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { IMaterialCost } from "@entities/material-cost/types"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"

const URL = "/input"

export interface InvoiceInputPagianted {
  data: IInvoiceInputView[]
  count: number
  page: number
}

export async function getPaginatedInvoiceInput({ pageParam = 1 }, searchParameters: InvoiceInputSearchParameters): Promise<InvoiceInputPagianted> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceInputPagianted>>(
  `${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&deliveryCode=${searchParameters.deliveryCode}&warehouseManagerWorkerID=${searchParameters.warehouseManagerWorkerID}&releasedWorkerID=${searchParameters.releasedWorkerID}&dateFrom=${searchParameters.dateFrom?.toString().substring(0, 24) ?? ""}&dateTo=${searchParameters.dateTo?.toString().substring(0, 24) ?? ""}&materials=${searchParameters.materials}`
  )
  const response = responseRaw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export async function deleteInvoiceInput(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<string>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceInputMaterial {
  materialData: InvoiceMaterial
  serialNumbers: string[]
}

export interface InvoiceInputMutation {
  details: IInvoiceInput
  items: InvoiceInputMaterial[]
}

export async function createInvoiceInput(data: InvoiceInputMutation): Promise<InvoiceInputMutation> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<InvoiceInputMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function updateInvoiceInput(data: InvoiceInputMutation): Promise<InvoiceInputMutation> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<InvoiceInputMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceInputDocument(deliveryCode: string): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/${deliveryCode}`, { responseType: "blob", })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, `${deliveryCode}.pdf`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export interface InvoiceInputConfirmationData {
  id: number
  file: File
}

export async function sendInvoiceInputConfirmationExcel(data: InvoiceInputConfirmationData): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data.file)
  const responseRaw = await axiosClient.post(`${URL}/confirm/${data.id}`, formData)
  if (responseRaw.status == 200) {
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
    throw new Error(responseRaw.data)
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

export async function getAllUniqueWarehouseManager(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/warehouse-manager`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllUniqueReleased(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/released`)
  const response = responseRaw.data
  if (response.permission || response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceInputReportFilter {
  code: string
  warehouseManagerID: number
  releasedID: number
  dateFrom: Date | null
  dateTo: Date | null
}

export async function buildReport(filter: InvoiceInputReportFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report`, filter, { responseType: "blob", })
  if (responseRaw.status == 200) {
    const date = new Date()
    const fileName = `Отчет Накладной Приход ${date}`
    fileDownload(responseRaw.data, `${fileName}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function createNewMaterialCostFromInvoiceInput(data: IMaterialCost): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<string[]>>(`${URL}/material-cost/new`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export interface CreateFullMaterial {
  category: string
  code: string
  name: string
  unit: string
  notes: string
  article: string
  hasSerialNumber: boolean
  costPrime: number
  costM19: number
  costWithCustomer: number
}

export async function createNewMaterialFromInvoiceInput(data: CreateFullMaterial): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<string[]>>(`${URL}/material/new`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceInputMaterilsWithoutSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithoutSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithoutSerialNumbers[]>>(`${URL}/${id}/materials/without-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceInputMaterilsWithSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithSerialNumbers[]>>(`${URL}/${id}/materials/with-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceInputMaterialsForEdit(id: number): Promise<IInvoiceInputMaterials[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IInvoiceInputMaterials[]>>(`${URL}/invoice-materials/${id}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function importInvoiceInput(file: File): Promise<boolean> {
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

export interface InvoiceInputSearchParametersData {
  deliveryCodes: string[]
  warehouseManagers: IReactSelectOptions<number>[]
  releaseds: IReactSelectOptions<number>[]
  materials: IReactSelectOptions<number>[]
}

export async function getInvoiceInputSearchParametersData(): Promise<InvoiceInputSearchParametersData> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceInputSearchParametersData>>(`${URL}/search-parameters`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceInputSearchParameters {
  deliveryCode: string
  warehouseManagerWorkerID: number
  releasedWorkerID: number
  dateFrom: null | Date
  dateTo: null | Date
  materials: number[]
}
