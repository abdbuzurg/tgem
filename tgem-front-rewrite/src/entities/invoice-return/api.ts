import fileDownload from "js-file-download"
import { IInvoiceReturn, IInvoiceReturnMaterials, IInvoiceReturnView } from "./types"
import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { IMaterialCost } from "@entities/material-cost/types"
import { InvoiceMaterialViewWithSerialNumbers, InvoiceMaterialViewWithoutSerialNumbers } from "@entities/invoice/types/invoiceMaterial"

const URL = "/return"

export interface InvoiceReturnPagianted {
  data: IInvoiceReturnView[]
  count: number
  page: number
}

export async function getPaginatedInvoiceReturn({ pageParam = 1 }, type: string): Promise<InvoiceReturnPagianted> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceReturnPagianted>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&type=${type}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  } else {
    throw new Error(response.error)
  }
}

export async function deleteInvoiceReturn(id: number): Promise<boolean> {
  const responseRaw = await axiosClient.delete<IApiResponseFormat<string>>(`${URL}/${id}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return true
  } else {
    throw new Error(response.error)
  }
}
export interface InvoiceReturnItem {
  materialID: number
  amount: number
  serialNumbers: string[]
  isDefected: boolean
  notes: string
}

export interface InvoiceReturnMutation {
  details: IInvoiceReturn
  items: InvoiceReturnItem[]
}

export async function createInvoiceReturn(data: InvoiceReturnMutation): Promise<InvoiceReturnMutation> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<InvoiceReturnMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function updateInvoiceReturn(data: InvoiceReturnMutation): Promise<InvoiceReturnMutation> {
  const responseRaw = await axiosClient.patch<IApiResponseFormat<InvoiceReturnMutation>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceReturnDocument(deliveryCode: string, confirmation: boolean): Promise<boolean> {
  const responseRaw = await axiosClient.get(`${URL}/document/${deliveryCode}`, { responseType: "blob", })
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

export interface InvoiceReturnConfirmation {
  id: number
  file: File
}

export async function sendInvoiceReturnConfirmationExcel(data: InvoiceReturnConfirmation): Promise<boolean> {
  const formData = new FormData()
  formData.append("file", data.file)
  const responseRaw = await axiosClient.post(`${URL}/confirm/${data.id}`, formData)
  if (responseRaw.status == 200) {
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function getUniqueCode(): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/unique/code`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getUniqueTeam(): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/unique/team`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getUniqueObject(): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/unique/object`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceReturnReportFilter {
  code: string
  returnerType: "all" | "teams" | "objects"
  returner: string
  dateFrom: Date | null
  dateTo: Date | null
}

export async function buildInvoiceReturnReport(filter: InvoiceReturnReportFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report`, filter, { responseType: "blob", })
  if (responseRaw.status == 200) {
    const fileName = "Отчет"
    fileDownload(responseRaw.data, `${fileName}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export interface InvoiceReturnMaterialsForSelect {
  materialID: number
  materialName: string
  materialUnit: string
  amount: number
  hasSerialNumber: boolean
}

export async function getUniqueMaterialsInLocation(locationType: string, locationID: number): Promise<InvoiceReturnMaterialsForSelect[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceReturnMaterialsForSelect[]>>(`${URL}/material/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getMaterialCostsInLocation(materialID: number, locationType: string, locationID: number): Promise<IMaterialCost[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IMaterialCost[]>>(`${URL}/material-cost/${materialID}/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getMaterialAmountInLocation(materialCostID: number, locationType: string, locationID: number): Promise<number> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<number>>(`${URL}/material-amount/${materialCostID}/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getSerialNumberCodesInLocation(materialID: number, locationType: string, locationID: number): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/serial-number/${locationType}/${locationID}/${materialID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}


export async function getInvoiceReturnMaterilsWithoutSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithoutSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithoutSerialNumbers[]>>(`${URL}/${id}/materials/without-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceReturnMaterilsWithSerialNumbersByID(id: number): Promise<InvoiceMaterialViewWithSerialNumbers[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceMaterialViewWithSerialNumbers[]>>(`${URL}/${id}/materials/with-serial-number`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceReturnMaterialsForEdit(id: number, locationType: string, locationID: number): Promise<IInvoiceReturnMaterials[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IInvoiceReturnMaterials[]>>(`${URL}/invoice-materials/${id}/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getMaterialAmountByMaterialID(materialID: number, locationID: number, locationType: string): Promise<number> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<number>>(`${URL}/amount/${locationType}/${locationID}/${materialID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}
