import IApiResponseFormat from "@shared/api/envelope"
import axiosClient from "@shared/api/client"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { UserActionFilter, UserActionFilterUserOption, UserActionPaginated, UserActionView } from "./types"

const URL = "/user-action"

function buildQuery(page: number, filter: UserActionFilter): string {
  const params = new URLSearchParams()
  params.set("page", String(page))
  params.set("limit", String(ENTRY_LIMIT))
  if (filter.userID && filter.userID > 0) params.set("userID", String(filter.userID))
  if (filter.projectID && filter.projectID > 0) params.set("projectID", String(filter.projectID))
  if (filter.actionType) params.set("actionType", filter.actionType)
  if (filter.httpMethod) params.set("httpMethod", filter.httpMethod)
  if (typeof filter.status === "boolean") params.set("status", String(filter.status))
  if (filter.dateFrom) params.set("dateFrom", filter.dateFrom)
  if (filter.dateTo) params.set("dateTo", filter.dateTo)
  return params.toString()
}

export async function getPaginatedUserActions(
  pageParam: number,
  filter: UserActionFilter,
): Promise<UserActionPaginated> {
  const qs = buildQuery(pageParam, filter)
  const responseRaw = await axiosClient.get<IApiResponseFormat<UserActionPaginated>>(
    `${URL}/paginated?${qs}`,
  )
  const response = responseRaw.data
  if (response.success && response.permission) {
    return { ...response.data, page: pageParam }
  }
  throw new Error(response.error)
}

export async function getUserActionsByUserID(userID: number): Promise<UserActionView[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<UserActionView[]>>(
    `${URL}/user/${userID}`,
  )
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data
  }
  throw new Error(response.error)
}

export async function getUserActionFilterUsers(): Promise<UserActionFilterUserOption[]> {
  const responseRaw = await axiosClient.get<IApiResponseFormat<UserActionFilterUserOption[]>>(
    `${URL}/filter-users`,
  )
  const response = responseRaw.data
  if (response.success && response.permission) {
    return response.data ?? []
  }
  throw new Error(response.error)
}
