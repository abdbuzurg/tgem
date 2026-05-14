import {
  ISIPObjectCreate,
  ISIPObjectPaginated,
  SIPObjectSearchParameters,
  createSIPObject,
  deleteSIPObject,
  exportSIP,
  getPaginatedSIPObjects,
  getSIPObjectNames,
  getSIPTemplateDocument,
  importSIP,
  updateSIPObject,
} from "@features/reference-books/object/api/sip"
import { getAllTPs } from "@features/reference-books/object/api/tp"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

type SIPDetailed = ISIPObjectCreate["detailedInfo"]

const fetchTPOptions = async (): Promise<IReactSelectOptions<number>[]> => {
  const tps = await getAllTPs()
  return tps.map((t) => ({ label: t.name, value: t.id }))
}

export const sipConfig: ObjectCrudConfig<ISIPObjectPaginated, SIPDetailed, SIPObjectSearchParameters, ISIPObjectCreate> = {
  typeID: "sip_objects",
  title: "Объекты - СИП",
  queryKey: "sip-object",

  api: {
    getPaginated: getPaginatedSIPObjects,
    create: createSIPObject,
    update: updateSIPObject,
    delete: deleteSIPObject,
    getNames: getSIPObjectNames,
    getTemplate: getSIPTemplateDocument,
    exportData: exportSIP,
    importData: importSIP,
  },

  emptyDetailed: { amountFeeders: 0 },
  emptySearch: { objectName: "", teamID: 0, supervisorWorkerID: 0 },

  detailFields: [
    { kind: "number", key: "amountFeeders", label: "Кол-во фидеров", required: true },
  ],

  detailColumns: [
    { header: "Кол-во фидеров", render: (row) => row.amountFeeders },
  ],

  rowToDetailedInfo: (row) => ({ amountFeeders: row.amountFeeders }),
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
