import {
  ISTVTObjectCreate,
  ISTVTObjectPaginated,
  STVTSearchParameters,
  createSTVTObject,
  deleteSTVTObject,
  exportSTVT,
  getPaginatedSTVTObjects,
  getSTVTObjectNames,
  getSTVTTemplateDocument,
  importSTVT,
  updateSTVTObject,
} from "@features/reference-books/object/api/stvt"
import { OBJECT_STATUSES_FOR_SELECT, STVT_OBJECT_VOLTAGE_CLASSES_FOR_SELECT } from "@shared/lib/data/objectStatuses"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

void OBJECT_STATUSES_FOR_SELECT // status options live in the scaffold; imported here only if a per-type override is needed

type STVTDetailed = ISTVTObjectCreate["detailedInfo"]

export const stvtConfig: ObjectCrudConfig<ISTVTObjectPaginated, STVTDetailed, STVTSearchParameters> = {
  typeID: "stvt_objects",
  title: "Объекты - СТВТ",
  queryKey: "stvt-object",

  api: {
    getPaginated: getPaginatedSTVTObjects,
    create: createSTVTObject,
    update: updateSTVTObject,
    delete: deleteSTVTObject,
    getNames: getSTVTObjectNames,
    getTemplate: getSTVTTemplateDocument,
    exportData: exportSTVT,
    importData: importSTVT,
  },

  emptyDetailed: {
    voltageClass: "",
    ttCoefficient: "",
  },
  emptySearch: {
    objectName: "",
    teamID: 0,
    supervisorWorkerID: 0,
  },

  detailFields: [
    {
      kind: "select",
      key: "voltageClass",
      label: "Класс напряжения",
      options: STVT_OBJECT_VOLTAGE_CLASSES_FOR_SELECT,
      required: true,
    },
    {
      kind: "text",
      key: "ttCoefficient",
      label: "Коэффицент ТТ",
    },
  ],

  detailColumns: [
    { header: "Класс напряжение", render: (row) => row.voltageClass },
    { header: "ТТ Коэффицент", render: (row) => row.ttCoefficient },
  ],

  rowToDetailedInfo: (row) => ({
    voltageClass: row.voltageClass,
    ttCoefficient: row.ttCoefficient,
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
