import {
  ISubstationCellObjectCreate,
  ISubstationCellObjectPaginated,
  SubstationCellObjectSearchParameters,
  createSubstationCellObject,
  deleteSubstationCellObject,
  exportSubstationCell,
  getPaginatedSubstationCellObjects,
  getSubstationCellObjectNames,
  getSubstationCellTemplateDocument,
  importSubstationCell,
  updateSubstationCellObject,
} from "@features/reference-books/object/api/substationCell"
import { getAllSubstations } from "@features/reference-books/object/api/substation"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

type SubstationCellDetailed = ISubstationCellObjectCreate["detailedInfo"]

const fetchSubstationOptions = async (): Promise<IReactSelectOptions<number>[]> => {
  const subs = await getAllSubstations()
  return subs.map((s) => ({ label: s.name, value: s.id }))
}

export const substationCellConfig: ObjectCrudConfig<ISubstationCellObjectPaginated, SubstationCellDetailed, SubstationCellObjectSearchParameters, ISubstationCellObjectCreate> = {
  typeID: "substation_cell_objects",
  title: "Объекты - Ячейка Подстанции",
  queryKey: "substation-cell-object",

  api: {
    getPaginated: getPaginatedSubstationCellObjects,
    create: createSubstationCellObject,
    update: updateSubstationCellObject,
    delete: deleteSubstationCellObject,
    getNames: getSubstationCellObjectNames,
    getTemplate: getSubstationCellTemplateDocument,
    exportData: exportSubstationCell,
    importData: importSubstationCell,
  },

  emptyDetailed: { amountFeeders: 0 },
  emptySearch: { objectName: "", teamID: 0, supervisorWorkerID: 0, substationObjectID: 0 },

  detailFields: [
    { kind: "number", key: "amountFeeders", label: "Кол-во фидеров", required: true },
  ],

  // SubstationCell's paginated row doesn't expose detailed fields, so no
  // detail columns shown in the table (matches the original page).
  detailColumns: [],

  // SubstationCell's paginated row doesn't include amountFeeders, so on edit
  // we can't pre-populate it from the row (matches original page behavior).
  rowToDetailedInfo: () => ({ amountFeeders: 0 }),
  rowID: (row) => row.objectID,
  rowDetailedID: (row) => row.objectDetailedID,
  rowName: (row) => row.name,
  rowStatus: (row) => row.status,
  rowSupervisors: (row) => row.supervisors,
  rowTeams: (row) => row.teams,

  setSearchObjectName: (s, value) => ({ ...s, objectName: value }),
  setSearchSupervisor: (s, value) => ({ ...s, supervisorWorkerID: value }),
  setSearchTeam: (s, value) => ({ ...s, teamID: value }),

  association: {
    kind: "single",
    label: "Подстанция",
    payloadKey: "substationObjectID",
    tableHeader: "Подстанция",
    rowRenderTableCell: (row) => row.substationName,
    rowSelectedLabel: (row) => row.substationName,
    queryKey: ["all-substations"],
    fetchOptions: fetchSubstationOptions,
  },

  toCreatePayload: (payload, _multi, single) => ({
    ...payload,
    substationObjectID: single,
  }),
}
