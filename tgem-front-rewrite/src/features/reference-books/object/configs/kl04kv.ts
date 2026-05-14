import {
  IKL04KVObjectCreate,
  IKL04KVObjectPaginated,
  KL04KVSearchParameters,
  createKL04KVObject,
  deleteKL04KVObject,
  exportKL04KV,
  getKL04KVObjectNames,
  getKL04KVTemplateDocument,
  getPaginatedKL04KVObjects,
  importKL04KV,
  updateKL04KVObject,
} from "@features/reference-books/object/api/kl04kv"
import { getAllTPs } from "@features/reference-books/object/api/tp"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

type KL04KVDetailed = IKL04KVObjectCreate["detailedInfo"]

const fetchTPOptions = async (): Promise<IReactSelectOptions<number>[]> => {
  const tps = await getAllTPs()
  return tps.map((t) => ({ label: t.name, value: t.id }))
}

export const kl04kvConfig: ObjectCrudConfig<IKL04KVObjectPaginated, KL04KVDetailed, KL04KVSearchParameters, IKL04KVObjectCreate> = {
  typeID: "kl04kv_objects",
  title: "Объекты - КЛ 04 кВ",
  queryKey: "kl04kv-object",

  api: {
    getPaginated: getPaginatedKL04KVObjects,
    create: createKL04KVObject,
    update: updateKL04KVObject,
    delete: deleteKL04KVObject,
    getNames: getKL04KVObjectNames,
    getTemplate: getKL04KVTemplateDocument,
    exportData: exportKL04KV,
    importData: importKL04KV,
  },

  emptyDetailed: { length: 0, nourashes: "" },
  emptySearch: { objectName: "", teamID: 0, supervisorWorkerID: 0, tpObjectID: 0 },

  detailFields: [
    { kind: "number", key: "length", label: "Длина", required: true },
    { kind: "text", key: "nourashes", label: "Питается от" },
  ],

  detailColumns: [
    { header: "Длина", render: (row) => row.length },
    { header: "Питается от", render: (row) => row.nourashes },
  ],

  rowToDetailedInfo: (row) => ({ length: row.length, nourashes: row.nourashes }),
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
