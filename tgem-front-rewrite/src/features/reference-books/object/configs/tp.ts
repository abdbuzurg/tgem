import {
  ITPObjectCreate,
  ITPObjectPaginated,
  TPObjectSearchParameters,
  createTPObject,
  deleteTPObject,
  exportTP,
  getPaginatedTPObjects,
  getTPObjectNames,
  getTPTemplateDocument,
  importTP,
  updateTPObject,
} from "@features/reference-books/object/api/tp"
import { TP_OBJECT_MODELS_FOR_SELECT, TP_OBJECT_VOLTAGE_CLASS_FOR_SELECT } from "@shared/lib/data/objectStatuses"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

type TPDetailed = ITPObjectCreate["detailedInfo"]

export const tpConfig: ObjectCrudConfig<ITPObjectPaginated, TPDetailed, TPObjectSearchParameters> = {
  typeID: "tp_objects",
  title: "Объекты - ТП",
  queryKey: "tp-object",

  api: {
    getPaginated: getPaginatedTPObjects,
    create: createTPObject,
    update: updateTPObject,
    delete: deleteTPObject,
    getNames: getTPObjectNames,
    getTemplate: getTPTemplateDocument,
    exportData: exportTP,
    importData: importTP,
  },

  emptyDetailed: { model: "", voltageClass: "", nourashes: "" },
  emptySearch: { objectName: "", teamID: 0, supervisorWorkerID: 0 },

  detailFields: [
    { kind: "select", key: "model", label: "Модель", options: TP_OBJECT_MODELS_FOR_SELECT, required: true },
    { kind: "select", key: "voltageClass", label: "Класс напряжения", options: TP_OBJECT_VOLTAGE_CLASS_FOR_SELECT, required: true },
    { kind: "text", key: "nourashes", label: "Питается от" },
  ],

  detailColumns: [
    { header: "Модель", render: (row) => row.model },
    { header: "Класс напряжения", render: (row) => row.voltageClass },
  ],

  rowToDetailedInfo: (row) => ({
    model: row.model,
    voltageClass: row.voltageClass,
    nourashes: (row as unknown as { nourashes?: string }).nourashes ?? "",
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
}
