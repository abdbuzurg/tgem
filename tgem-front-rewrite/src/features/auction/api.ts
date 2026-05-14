import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"

const URL = "/auction"

interface AuctionItemDataPublic {
  name: string
  description: string
  unit: string
  quantity: number
  note: string
}

interface AuctionParticipantForPublic {
  participantTitle: string
  totalPrice: number
}

export interface AuctionPublicData {
  packageName: string
  packageItems: AuctionItemDataPublic[]
  participantTotalPriceForPackage: AuctionParticipantForPublic[]
}

export async function getDataForPublicAuction(auctionID: number): Promise<AuctionPublicData[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<AuctionPublicData[]>>(`${URL}/${auctionID}`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

interface AuctionItemDataPrivate {
  itemID: number
  name: string
  description: string
  unit: string
  quantity: number
  note: string
  userUnitPrice: number
  comment: string
}

export interface AuctionPrivateData {
  packageName: string
  packageItems: AuctionItemDataPrivate[]
  minimumPackagePrice: number
}

export async function getDataForPrivateAuction(auctionID: number): Promise<AuctionPrivateData[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<AuctionPrivateData[]>>(`${URL}/private/${auctionID}`)
  const response = responseRaw.data
  if (response.success) {
    return response.data
  } else {
    throw new Error(response.error)
  }
}

export interface AuctionPrivateParticipantDataForSave {
  itemID: number
  comment: string
  unitPrice: number
}

export async function savePartisipantChanges(data: AuctionPrivateParticipantDataForSave[]): Promise<boolean> {
  const responseRaw = await axiosClient.post<IApiResponseFormat<boolean>>(`${URL}/private`, data)
  const response = responseRaw.data
  if (response.success) {
    return true
  } else {
    throw new Error(response.error)
  }
}
