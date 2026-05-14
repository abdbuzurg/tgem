import {
  ISubstationObjectCreate,
  ISubstationObjectPaginated,
  SubstationObjectSearchParameters,
  createSubstationObject,
  deleteSubstationObject,
  exportSubstation,
  getPaginatedSubstationObjects,
  getSubstationObjectNames,
  getSubstationTemplateDocument,
  importSubstation,
  updateSubstationObject,
} from "@features/reference-books/object/api/substation"
import { SUBSTATION_OBJECT_VOLTAGE_CLASS_FOR_SELECT } from "@shared/lib/data/objectStatuses"
import { ObjectCrudConfig } from "../_scaffold/ObjectCrudConfig"

type SubstationDetailed = ISubstationObjectCreate["detailedInfo"]

export const substationConfig: ObjectCrudConfig<ISubstationObjectPaginated, SubstationDetailed, SubstationObjectSearchParameters> = {
  typeID: "substation_objects",
  title: "Объекты - Подстанция",
  queryKey: "substation-object",

  api: {
    getPaginated: getPaginatedSubstationObjects,
    create: createSubstationObject,
    update: updateSubstationObject,
    delete: deleteSubstationObject,
    getNames: getSubstationObjectNames,
    getTemplate: getSubstationTemplateDocument,
    exportData: exportSubstation,
    importData: importSubstation,
  },

  emptyDetailed: { voltageClass: "", numberOfTransformers: 0 },
  emptySearch: { objectName: "", teamID: 0, supervisorWorkerID: 0 },

  detailFields: [
    { kind: "select", key: "voltageClass", label: "Класс напряжения", options: SUBSTATION_OBJECT_VOLTAGE_CLASS_FOR_SELECT, required: true },
    { kind: "number", key: "numberOfTransformers", label: "Кол-во трансформаторов", required: true },
  ],

  detailColumns: [
    { header: "Класс напряжения", render: (row) => row.voltageClass },
    { header: "Трансформаторы", render: (row) => row.numberOfTransformers },
  ],

  rowToDetailedInfo: (row) => ({
    voltageClass: row.voltageClass,
    numberOfTransformers: +row.numberOfTransformers,
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
