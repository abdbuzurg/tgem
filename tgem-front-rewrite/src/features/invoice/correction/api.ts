import fileDownload from "js-file-download";
import { IInvoiceObject } from "@entities/invoice/types/invoiceObject";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";
import { ObjectDataForSelect } from "@entities/object/types";
import { ENTRY_LIMIT } from "@shared/config/pagination";

const URL = "/invoice-correction"

export interface InvoiceCorrectionPaginated {
  data: InvoiceCorrectionPaginatedView[]
  count: number
  page: number
}

export interface InvoiceCorrectionPaginatedView {
  id: number
  deliveryCode: string
  supervisorName: string
  objectName: string
  objectType: string
  teamID: number
  teamLeaderName: string
  dateOfInvoice: Date
}

export interface InvoiceCorrectionSearchParameters {
  teamID: number
  objectID: number
}

export async function getPaginatedInvoiceCorrection({ pageParam = 1 }, searchParamters: InvoiceCorrectionSearchParameters): Promise<InvoiceCorrectionPaginated> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceCorrectionPaginated>>(`${URL}/paginated?page=${pageParam}&limit=${ENTRY_LIMIT}&teamID=${searchParamters.teamID}&objectID=${searchParamters.objectID}`)
  const responseData = responseRaw.data
  if (responseData.success && responseData.permission) {
    return { ...responseData.data, page: pageParam }
  } else {
    throw new Error(responseData.error)
  }
}


export async function getAllInvoiceObjectsForCorrect(): Promise<InvoiceCorrectionPaginatedView[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceCorrectionPaginatedView[]>>(`${URL}/`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceCorrectionMaterial {
  invoiceMaterialID: number
  materialName: string
  materialID: number
  materialAmount: number
  materialAvailableAmount: number
  materialUnit: string
  notes: string
}

export async function getInvoiceMaterialsForCorrect(invoiceID: number): Promise<InvoiceCorrectionMaterial[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceCorrectionMaterial[]>>(`${URL}/materials/${invoiceID}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getTotalAmounByTeamNumber(materialID: number, teamNumber: string): Promise<number> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<number>>(`${URL}/total-amount/${materialID}/team/${teamNumber}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getSerialNumbersOfMaterialInTeam(materialID: number, teamID: number): Promise<string[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<string[]>>(`${URL}/serial-number/material/${materialID}/teams/${teamID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceCorrectionMaterialMutation {
  details: {
    id: number
    dateOfCorrection: Date
  },
  items: InvoiceCorrectionMaterial[]
  operations: InvoiceCorrectionOperation[]
}

export async function createInvoiceCorrection(data: InvoiceCorrectionMaterialMutation): Promise<IInvoiceObject> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<IInvoiceObject>>(`${URL}/`, data)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceCorrectionUniqueObjects(): Promise<ObjectDataForSelect[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<ObjectDataForSelect[]>>(`${URL}/unique/object`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getInvoiceCorrectionUniqueTeams(): Promise<IReactSelectOptions<number>[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<IReactSelectOptions<number>[]>>(`${URL}/unique/team`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceCorrectionReportFilter {
  teamID: number
  objectID: number
  dateTo: Date | null
  dateFrom: Date | null
}

export async function buildInvoiceCorrectionReport(filter: InvoiceCorrectionReportFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report`, filter, { responseType: "blob", })
  if (responseRaw.status == 200) {
    const date = new Date()
    const fileName = `Отчет Расхода ${date}`
    fileDownload(responseRaw.data, `${fileName}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export interface InvoiceCorrectionOperation {
  operationID: number
  operationName: string
  materialName: string
  amount: number
}

export async function getOperationsForCorrect(invoiceID: number): Promise<InvoiceCorrectionOperation[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceCorrectionOperation[]>>(`${URL}/operations/${invoiceID}`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface InvoiceCorrectionSearchParametersData {
  teams: IReactSelectOptions<number>[]
  objects: IReactSelectOptions<number>[]
}

export async function getInvoiceCorrectionSearchParameters(): Promise<InvoiceCorrectionSearchParametersData> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<InvoiceCorrectionSearchParametersData>>(`${URL}/search-parameters`)
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}
