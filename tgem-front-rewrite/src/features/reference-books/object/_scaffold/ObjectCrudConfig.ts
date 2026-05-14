import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { ReactNode } from "react"

export type DetailField =
  | { kind: "text"; key: string; label: string; required?: boolean }
  | { kind: "number"; key: string; label: string; required?: boolean }
  | { kind: "checkbox"; key: string; label: string }
  | { kind: "select"; key: string; label: string; options: IReactSelectOptions<string>[]; required?: boolean }

export interface DetailColumn<TPaginated> {
  header: string
  render: (row: TPaginated) => ReactNode
}

export type Association<TPaginated> =
  | {
      kind: "multi"
      label: string
      payloadKey: string
      tableHeader: string
      // names of the picked options, as exposed on the paginated row
      // (used to pre-select on edit)
      rowSelectedNames: (row: TPaginated) => string[]
      queryKey: readonly unknown[]
      fetchOptions: () => Promise<IReactSelectOptions<number>[]>
    }
  | {
      kind: "single"
      label: string
      payloadKey: string
      tableHeader: string
      rowRenderTableCell: (row: TPaginated) => ReactNode
      // label of the option currently selected on this row; the scaffold
      // resolves it to an option by matching against fetched options.
      rowSelectedLabel: (row: TPaginated) => string
      queryKey: readonly unknown[]
      fetchOptions: () => Promise<IReactSelectOptions<number>[]>
    }

export interface PaginatedResponse<TRow> {
  page: number
  count: number
  data: TRow[]
}

export interface BaseInfo {
  id: number
  projectID: number
  objectDetailedID: number
  type: string
  name: string
  status: string
}

export interface MutationPayload<TDetailed> {
  baseInfo: BaseInfo
  detailedInfo: TDetailed
  supervisors: number[]
  teams: number[]
}

export interface ObjectCrudApi<TPaginated, TSearch, TCreate> {
  getPaginated: (
    page: { pageParam?: number },
    search: TSearch,
  ) => Promise<PaginatedResponse<TPaginated>>
  create: (data: TCreate) => Promise<boolean>
  update: (data: TCreate) => Promise<boolean>
  delete: (id: number) => Promise<boolean>
  getNames: () => Promise<IReactSelectOptions<string>[]>
  getTemplate: () => Promise<boolean>
  exportData: () => Promise<boolean>
  importData: (file: File) => Promise<boolean>
}

export interface ObjectCrudConfig<
  TPaginated,
  TDetailed extends Record<string, unknown>,
  TSearch,
  TCreate = MutationPayload<TDetailed>,
> {
  typeID: string
  title: string
  queryKey: string
  api: ObjectCrudApi<TPaginated, TSearch, TCreate>
  emptyDetailed: TDetailed
  emptySearch: TSearch
  detailFields: DetailField[]
  detailColumns: DetailColumn<TPaginated>[]
  rowToDetailedInfo: (row: TPaginated) => TDetailed
  rowID: (row: TPaginated) => number
  rowDetailedID: (row: TPaginated) => number
  rowName: (row: TPaginated) => string
  rowStatus: (row: TPaginated) => string
  rowSupervisors: (row: TPaginated) => string[]
  rowTeams: (row: TPaginated) => string[]
  setSearchObjectName: (search: TSearch, value: string) => TSearch
  setSearchSupervisor: (search: TSearch, value: number) => TSearch
  setSearchTeam: (search: TSearch, value: number) => TSearch
  association?: Association<TPaginated>
  // Required when TCreate ≠ MutationPayload<TDetailed> (i.e., the API expects
  // an extra association key on the payload). Receives the scaffold's
  // MutationPayload plus the picked association IDs (one of the two is unused
  // depending on association.kind). Returns the per-type create payload.
  toCreatePayload?: (
    payload: MutationPayload<TDetailed>,
    assocMultiIds: number[],
    assocSingleId: number,
  ) => TCreate
}
