import fileDownload from "js-file-download";
import Material from "@entities/material/types";
import { IMaterialCost } from "@entities/material-cost/types";
import IApiResponseFormat from "@shared/api/envelope";
import axiosClient from "@shared/api/client";
import writeOffTypeToRus from "@shared/lib/writeOffTypeToRus";

const URL = "/material-location"

export async function getAmountInWarehouse(materialID: number): Promise<number> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<number>>(`/material-location/total-amount/${materialID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAllMaterialInALocation(locationType: string, locationID: number): Promise<Material[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<Material[]>>(`/material-location/available/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getMaterailCostsInALocation(locationType: string, locationID: number, materialID: number): Promise<IMaterialCost[]> {
  console.log(locationType, locationID, materialID)
  const responseRaw = await axiosClient.get<IApiResponseFormat<IMaterialCost[]>>(`/material-location/costs/${materialID}/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export async function getAmountByCostAndLocation(locationType: string, locationID: number, materialCostID: number): Promise<number> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<number>>(`/material-location/amount/${materialCostID}/${locationType}/${locationID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface MaterialLocationLiveSearchParameters {
  locationType: string
  locationID: number
  materialID: number
}

export interface MaterialLocationLiveView {
  materialID: number
  materialName: number
  materialUnit: string
  materialCostID: number
  materialCostM19: number
  locationType: string
  locationID: number
  amount: number
}

export async function getMaterialLocation(searchParameters: MaterialLocationLiveSearchParameters): Promise<MaterialLocationLiveView[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<MaterialLocationLiveView[]>>(`${URL}/live?locationType=${searchParameters.locationType}&locationID=${searchParameters.locationID}&materialID=${searchParameters.materialID}`)
  const response = responseRaw.data
  if (response.permission && response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface ReportWriteOffBalanceFilter {
  writeOffType: string
  locationID: number
}

export async function buildWriteOffBalanceReport(filter: ReportWriteOffBalanceFilter): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report/balance/writeoff`, filter, { responseType: "blob", })
  if (responseRaw.status == 200) {
    const fileName = writeOffTypeToRus(filter.writeOffType) 
    fileDownload(responseRaw.data, `${fileName}.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}

export async function buildOutOfProjectBalanceReport(): Promise<boolean> {
  const responseRaw = await axiosClient.post(`${URL}/report/balance/out-of-project`, null, { responseType: "blob", })
  if (responseRaw.status == 200) {
    fileDownload(responseRaw.data, `Отчет остатка материалов вне проекта.xlsx`)
    return true
  } else {
    throw new Error(responseRaw.data)
  }
}


