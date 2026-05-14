import {
  IMJDObjectCreate,
  IMJDObjectPaginated,
  MJDObjectSearchParameters,
  createMJDObject,
  deleteMJDObject,
  exportMJD,
  getMJDObjectNames,
  getMJDTemplateDocument,
  getPaginatedMJDObjects,
  importMJD,
  updateMJDObject,
} from "@features/reference-books/object/api/mjd"
import { getAllTPs } from "@features/reference-books/object/api/tp"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { MJD_OBJECT_TYPES_FOR_SELECT } from "@shared/lib/data/objectStatuses"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

type MJDDetailed = IMJDObjectCreate["detailedInfo"]

const fetchTPOptions = async (): Promise<IReactSelectOptions<number>[]> => {
  const tps = await getAllTPs()
  return tps.map((t) => ({ label: t.name, value: t.id }))
}

export const mjdConfig: ObjectCrudConfig<IMJDObjectPaginated, MJDDetailed, MJDObjectSearchParameters, IMJDObjectCreate> = {
  typeID: "mjd_objects",
  title: "Объекты - МЖД",
  queryKey: "mjd-objects",

  api: {
    getPaginated: getPaginatedMJDObjects,
    create: createMJDObject,
    update: updateMJDObject,
    delete: deleteMJDObject,
    getNames: getMJDObjectNames,
    getTemplate: getMJDTemplateDocument,
    exportData: exportMJD,
    importData: importMJD,
  },

  emptyDetailed: { model: "", amountStores: 0, amountEntrances: 0, hasBasement: true },
  emptySearch: { objectName: "", teamID: 0, supervisorWorkerID: 0, tpObjectID: 0 },

  detailFields: [
    { kind: "select", key: "model", label: "Тип здания", options: MJD_OBJECT_TYPES_FOR_SELECT },
    { kind: "number", key: "amountStores", label: "Кол-во этажей", required: true },
    { kind: "number", key: "amountEntrances", label: "Кол-во подъездов", required: true },
    { kind: "checkbox", key: "hasBasement", label: "Есть подвал" },
  ],

  detailColumns: [
    { header: "Тип", render: (row) => row.model },
    { header: "Этажи", render: (row) => row.amountStores },
    { header: "Подъезды", render: (row) => row.amountEntrances },
    { header: "Подвал", render: (row) => (row.hasBasement ? "ДА" : "НЕТ") },
  ],

  rowToDetailedInfo: (row) => ({
    model: row.model,
    amountStores: row.amountStores,
    amountEntrances: row.amountEntrances,
    hasBasement: row.hasBasement,
  }),
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
    kind: "multi",
    label: "Питается от ТП",
    payloadKey: "nourashedByTP",
    tableHeader: "ТП",
    rowSelectedNames: (row) => row.tpNames,
    queryKey: ["all-tp-objects"],
    fetchOptions: fetchTPOptions,
  },

  toCreatePayload: (payload, multi) => ({
    ...payload,
    nourashedByTP: multi,
  }),
}
