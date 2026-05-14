import { useEffect, useState } from "react"
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import useScrollPaginated from "@shared/hooks/useScrollPaginated"
import Select from "react-select"
import toast from "react-hot-toast"

import Button from "@shared/ui/Button"
import Input from "@shared/ui/Input"
import LoadingDots from "@shared/ui/LoadingDots"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { OBJECT_STATUSES_FOR_SELECT } from "@shared/lib/data/objectStatuses"
import arrayListToString from "@shared/lib/arrayListToStringWithCommas"
import IWorker from "@entities/worker/types"
import { getWorkerByJobTitle } from "@entities/worker/api"
import { TeamDataForSelect } from "@entities/team/types"
import { getAllTeamsForSelect } from "@entities/team/api"

import Modal from "@shared/components/Modal"
import DeleteModal from "@shared/components/DeleteModal"

import {
  Association,
  DetailField,
  MutationPayload,
  ObjectCrudConfig,
} from "./ObjectCrudConfig"

interface Props<TPaginated, TDetailed extends Record<string, unknown>, TSearch, TCreate = MutationPayload<TDetailed>> {
  config: ObjectCrudConfig<TPaginated, TDetailed, TSearch, TCreate>
}

export default function ObjectCrudPage<
  TPaginated,
  TDetailed extends Record<string, unknown>,
  TSearch,
  TCreate = MutationPayload<TDetailed>,
>({
  config,
}: Props<TPaginated, TDetailed, TSearch, TCreate>) {
  const queryClient = useQueryClient()

  // SEARCH STATE
  const [searchParameters, setSearchParameters] = useState<TSearch>(config.emptySearch)

  // PAGINATED LIST
  const tableDataQuery = useInfiniteQuery<{ page: number; count: number; data: TPaginated[] }, Error>({
    queryKey: [config.queryKey, searchParameters],
    queryFn: ({ pageParam }) => config.api.getPaginated({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    },
  })

  const [tableData, setTableData] = useState<TPaginated[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data = tableDataQuery.data.pages.reduce<TPaginated[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  // DELETE
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const deleteMutation = useMutation({
    mutationFn: config.api.delete,
    onSuccess: () => queryClient.invalidateQueries([config.queryKey]),
  })
  const [deleteModalProps, setDeleteModalProps] = useState({
    setShowModal: setShowDeleteModal,
    no_delivery: "",
    deleteFunc: () => {},
  })
  const onDeleteButtonClick = (row: TPaginated) => {
    setShowDeleteModal(true)
    setDeleteModalProps({
      deleteFunc: () => deleteMutation.mutate(config.rowDetailedID(row)),
      no_delivery: config.rowName(row),
      setShowModal: setShowDeleteModal,
    })
  }

  // MUTATION (create + update share state, gate by mutationType)
  const [showMutationModal, setShowMutationModal] = useState(false)
  const [mutationType, setMutationType] = useState<null | "create" | "update">(null)

  const buildEmptyMutationData = (): MutationPayload<TDetailed> => ({
    baseInfo: {
      id: 0,
      projectID: 0,
      objectDetailedID: 0,
      type: config.typeID,
      name: "",
      status: "",
    },
    detailedInfo: { ...config.emptyDetailed },
    supervisors: [],
    teams: [],
  })

  const [mutationData, setMutationData] = useState<MutationPayload<TDetailed>>(buildEmptyMutationData())

  // SUPERVISORS
  const [selectedSupervisors, setSelectedSupervisors] = useState<IReactSelectOptions<number>[]>([])
  const [availableSupervisors, setAvailableSupervisors] = useState<IReactSelectOptions<number>[]>([])
  const supervisorsQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["worker-supervisors"],
    queryFn: () => getWorkerByJobTitle("Супервайзер"),
  })
  useEffect(() => {
    if (supervisorsQuery.isSuccess && supervisorsQuery.data) {
      setAvailableSupervisors(
        supervisorsQuery.data.map<IReactSelectOptions<number>>((val) => ({ label: val.name, value: val.id })),
      )
    }
  }, [supervisorsQuery.data, supervisorsQuery.isSuccess])

  // TEAMS
  const [selectedTeams, setSelectedTeams] = useState<IReactSelectOptions<number>[]>([])
  const [availableTeams, setAvailableTeams] = useState<IReactSelectOptions<number>[]>([])
  const teamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["all-teams-for-select"],
    queryFn: getAllTeamsForSelect,
  })
  useEffect(() => {
    if (teamsQuery.isSuccess && teamsQuery.data) {
      setAvailableTeams(
        teamsQuery.data.map<IReactSelectOptions<number>>((val) => ({
          label: val.teamNumber + " (" + val.teamLeaderName + ")",
          value: val.id,
        })),
      )
    }
  }, [teamsQuery.data, teamsQuery.isSuccess])

  // OPTIONAL EXTRA ASSOCIATION (multi or single)
  const [selectedAssocMulti, setSelectedAssocMulti] = useState<IReactSelectOptions<number>[]>([])
  const [selectedAssocSingle, setSelectedAssocSingle] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [availableAssoc, setAvailableAssoc] = useState<IReactSelectOptions<number>[]>([])
  const assocQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: config.association ? config.association.queryKey : ["unused-assoc"],
    queryFn: () => (config.association ? config.association.fetchOptions() : Promise.resolve([])),
    enabled: !!config.association,
  })
  useEffect(() => {
    if (assocQuery.isSuccess && assocQuery.data) {
      setAvailableAssoc(assocQuery.data)
    }
  }, [assocQuery.data, assocQuery.isSuccess])

  // CREATE / UPDATE
  const createMutation = useMutation<boolean, Error, TCreate>({
    mutationFn: config.api.create,
  })
  const updateMutation = useMutation<boolean, Error, TCreate>({
    mutationFn: config.api.update,
  })

  const buildCreatePayload = (): TCreate => {
    if (config.toCreatePayload) {
      return config.toCreatePayload(
        mutationData,
        selectedAssocMulti.map((v) => v.value),
        selectedAssocSingle.value,
      )
    }
    return mutationData as unknown as TCreate
  }

  const validateDetailedInfo = (info: TDetailed): string | null => {
    for (const f of config.detailFields) {
      if (f.kind === "checkbox") continue
      if (!f.required) continue
      const value = info[f.key]
      if (f.kind === "text" || f.kind === "select") {
        if (!value || value === "") return `Не указано: ${f.label}`
      } else if (f.kind === "number") {
        if (!value || value === 0) return `Не указано: ${f.label}`
      }
    }
    return null
  }

  const onMutationSubmitClick = () => {
    if (mutationData.baseInfo.name === "") {
      toast.error("Не указано наименование объекта.")
      return
    }
    if (mutationData.baseInfo.status === "") {
      toast.error("Не указан статус объекта.")
      return
    }
    const detailErr = validateDetailedInfo(mutationData.detailedInfo)
    if (detailErr) {
      toast.error(detailErr)
      return
    }

    const onSuccess = () => {
      queryClient.invalidateQueries([config.queryKey])
      setShowMutationModal(false)
    }

    const payload = buildCreatePayload()
    if (mutationType === "create") createMutation.mutate(payload, { onSuccess })
    if (mutationType === "update") updateMutation.mutate(payload, { onSuccess })
  }

  const onAddClick = () => {
    setMutationType("create")
    setMutationData(buildEmptyMutationData())
    setSelectedSupervisors([])
    setSelectedTeams([])
    setSelectedAssocMulti([])
    setSelectedAssocSingle({ label: "", value: 0 })
    setShowMutationModal(true)
  }

  const onEditClick = (index: number) => {
    const row = tableData[index]
    const supervisors = availableSupervisors.filter((val) => config.rowSupervisors(row).includes(val.label))

    const teamNamesOnly = availableTeams.map<IReactSelectOptions<number>>((val) => ({
      ...val,
      label: val.label.split(" ")[0],
    }))
    const teams = teamNamesOnly.filter((val) => config.rowTeams(row).includes(val.label))
    const fullTeamNames = availableTeams.filter((val) => teams.find((t) => t.value === val.value))

    const next: MutationPayload<TDetailed> = {
      baseInfo: {
        id: config.rowID(row),
        projectID: 0,
        objectDetailedID: config.rowDetailedID(row),
        type: config.typeID,
        name: config.rowName(row),
        status: config.rowStatus(row),
      },
      detailedInfo: config.rowToDetailedInfo(row),
      supervisors: supervisors.map((s) => s.value),
      teams: teams.map((t) => t.value),
    }

    if (config.association?.kind === "multi") {
      const assoc = config.association
      const picked = availableAssoc.filter((val) => assoc.rowSelectedNames(row).includes(val.label))
      setSelectedAssocMulti(picked)
    } else if (config.association?.kind === "single") {
      const assoc = config.association
      const wantedLabel = assoc.rowSelectedLabel(row)
      const picked = availableAssoc.find((val) => val.label === wantedLabel) ?? { label: "", value: 0 }
      setSelectedAssocSingle(picked)
    }

    setMutationData(next)
    setSelectedSupervisors(supervisors)
    setSelectedTeams(fullTeamNames)
    setShowMutationModal(true)
    setMutationType("update")
  }

  // IMPORT / EXPORT
  const [showImportModal, setShowImportModal] = useState(false)
  const importTemplateQuery = useQuery<boolean, Error, boolean>({
    queryKey: [`${config.queryKey}-template`],
    queryFn: config.api.getTemplate,
    enabled: false,
  })
  const importMutation = useMutation<boolean, Error, File>({
    mutationFn: config.api.importData,
  })
  const acceptExcel = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importMutation.mutate(e.target.files[0], {
      onSuccess: () => {
        queryClient.invalidateQueries([config.queryKey])
        setShowImportModal(false)
      },
      onSettled: () => {
        e.target.value = ""
      },
      onError: (error) => {
        toast.error(`Импортированный файл имеет неправильные данные: ${error.message}`)
      },
    })
  }
  const exportQuery = useQuery<boolean, Error, boolean>({
    queryKey: [`${config.queryKey}-export`],
    queryFn: config.api.exportData,
    enabled: false,
  })

  // SEARCH MODAL
  const [showSearchModal, setShowSearchModal] = useState(false)
  const [selectedSearchObjectName, setSelectedSearchObjectName] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allObjectNames, setAllObjectNames] = useState<IReactSelectOptions<string>[]>([])
  const allObjectNamesQuery = useQuery<IReactSelectOptions<string>[], Error, IReactSelectOptions<string>[]>({
    queryKey: [`${config.queryKey}-names`],
    queryFn: config.api.getNames,
    enabled: showSearchModal,
  })
  useEffect(() => {
    if (allObjectNamesQuery.isSuccess && allObjectNamesQuery.data) {
      setAllObjectNames(allObjectNamesQuery.data)
    }
  }, [allObjectNamesQuery.data, allObjectNamesQuery.isSuccess])

  const [selectedSearchSupervisor, setSelectedSearchSupervisor] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allSearchSupervisors, setAllSearchSupervisors] = useState<IReactSelectOptions<number>[]>([])
  const allSearchSupervisorsQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["all-workers", "Супервайзер"],
    queryFn: () => getWorkerByJobTitle("Супервайзер"),
    enabled: showSearchModal,
  })
  useEffect(() => {
    if (allSearchSupervisorsQuery.isSuccess && allSearchSupervisorsQuery.data) {
      setAllSearchSupervisors(
        allSearchSupervisorsQuery.data.map<IReactSelectOptions<number>>((val) => ({ label: val.name, value: val.id })),
      )
    }
  }, [allSearchSupervisorsQuery.data, allSearchSupervisorsQuery.isSuccess])

  const [selectedSearchTeam, setSelectedSearchTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allSearchTeams, setAllSearchTeams] = useState<IReactSelectOptions<number>[]>([])
  const allSearchTeamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["all-teams-for-select"],
    queryFn: getAllTeamsForSelect,
    enabled: showSearchModal,
  })
  useEffect(() => {
    if (allSearchTeamsQuery.isSuccess && allSearchTeamsQuery.data) {
      setAllSearchTeams(
        allSearchTeamsQuery.data.map<IReactSelectOptions<number>>((val) => ({
          label: val.teamNumber + " (" + val.teamLeaderName + ")",
          value: val.id,
        })),
      )
    }
  }, [allSearchTeamsQuery.data, allSearchTeamsQuery.isSuccess])

  const onResetSearch = () => {
    setSearchParameters(config.emptySearch)
    setSelectedSearchObjectName({ label: "", value: "" })
    setSelectedSearchSupervisor({ label: "", value: 0 })
    setSelectedSearchTeam({ label: "", value: 0 })
  }

  const renderDetailField = (f: DetailField) => {
    const value = mutationData.detailedInfo[f.key] as unknown
    const requireMark = f.kind !== "checkbox" && f.required ? <span className="text-red-600">*</span> : null
    const setDetail = (next: unknown) => {
      setMutationData({
        ...mutationData,
        detailedInfo: { ...mutationData.detailedInfo, [f.key]: next } as TDetailed,
      })
    }

    if (f.kind === "text") {
      return (
        <div className="flex flex-col space-y-1" key={f.key}>
          <label htmlFor={f.key}>
            {f.label}
            {requireMark}
          </label>
          <Input
            name={f.key}
            type="text"
            value={(value as string) ?? ""}
            onChange={(e) => setDetail(e.target.value)}
          />
        </div>
      )
    }

    if (f.kind === "number") {
      return (
        <div className="flex flex-col space-y-1" key={f.key}>
          <label htmlFor={f.key}>
            {f.label}
            {requireMark}
          </label>
          <Input
            name={f.key}
            type="number"
            value={(value as number) ?? 0}
            onChange={(e) => setDetail(e.target.valueAsNumber)}
          />
        </div>
      )
    }

    if (f.kind === "checkbox") {
      return (
        <div className="flex space-x-2 items-center" key={f.key}>
          <input
            id={f.key}
            type="checkbox"
            checked={!!value}
            onChange={(e) => setDetail(e.currentTarget.checked)}
          />
          <label htmlFor={f.key}>{f.label}</label>
        </div>
      )
    }

    // select
    return (
      <div className="flex flex-col space-y-1" key={f.key}>
        <label htmlFor={f.key}>
          {f.label}
          {requireMark}
        </label>
        <Select
          className="basic-single text-black"
          classNamePrefix="select"
          isSearchable={true}
          isClearable={true}
          name={f.key}
          placeholder={""}
          value={{ label: (value as string) ?? "", value: (value as string) ?? "" }}
          options={f.options}
          onChange={(opt) => setDetail(opt?.value ?? "")}
        />
      </div>
    )
  }

  const renderAssociationField = (assoc: Association<TPaginated>) => {
    if (assoc.kind === "multi") {
      return (
        <div>
          <label>{assoc.label}</label>
          <Select
            className="basic-single text-black"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            isMulti
            name={`${assoc.payloadKey}-select`}
            placeholder={""}
            value={selectedAssocMulti}
            options={availableAssoc}
            onChange={(value) => {
              setSelectedAssocMulti([...value])
            }}
          />
        </div>
      )
    }
    return (
      <div>
        <label>{assoc.label}</label>
        <Select
          className="basic-single text-black"
          classNamePrefix="select"
          isSearchable={true}
          isClearable={true}
          name={`${assoc.payloadKey}-select`}
          placeholder={""}
          value={selectedAssocSingle}
          options={availableAssoc}
          onChange={(value) => {
            const v = value ?? { label: "", value: 0 }
            setSelectedAssocSingle(v)
          }}
        />
      </div>
    )
  }

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">{config.title}</span>
        <div onClick={() => setShowSearchModal(true)} className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer">
          Поиск
        </div>
        <Button text="Импорт" onClick={() => setShowImportModal(true)} />
        <div
          onClick={() => exportQuery.refetch()}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          {exportQuery.fetchStatus === "fetching" ? <LoadingDots height={20} /> : "Экспорт"}
        </div>
        <div
          onClick={onResetSearch}
          className="text-white py-2.5 px-5 rounded-lg bg-red-700 hover:bg-red-800 hover:cursor-pointer"
        >
          Сброс поиска
        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3"><span>Наименование</span></th>
            <th className="px-4 py-3"><span>Статус</span></th>
            {config.detailColumns.map((c) => (
              <th key={c.header} className="px-4 py-3"><span>{c.header}</span></th>
            ))}
            {config.association && (
              <th className="px-4 py-3 w-[150px]"><span>{config.association.tableHeader}</span></th>
            )}
            <th className="px-4 py-3 w-[150px]"><span>Супервайзер</span></th>
            <th className="px-4 py-3 w-[150px]"><span>Бригады</span></th>
            <th className="px-4 py-3"><Button text="Добавить" onClick={onAddClick} /></th>
          </tr>
        </thead>
        <tbody>
          {tableDataQuery.isLoading &&
            <tr><td colSpan={6}><LoadingDots /></td></tr>
          }
          {tableDataQuery.isError &&
            <tr>
              <td colSpan={6} className="text-red font-bold text-center">
                {tableDataQuery.error.message}
              </td>
            </tr>
          }
          {tableDataQuery.isSuccess && tableData.length !== 0 &&
            tableData.map((row, index) => (
              <tr key={index} className="border-b">
                <td className="px-4 py-3">{config.rowName(row)}</td>
                <td className="px-4 py-3">{config.rowStatus(row)}</td>
                {config.detailColumns.map((c) => (
                  <td key={c.header} className="px-4 py-3">{c.render(row)}</td>
                ))}
                {config.association?.kind === "single" && (
                  <td className="px-4 py-3">{config.association.rowRenderTableCell(row)}</td>
                )}
                {config.association?.kind === "multi" && (
                  <td className="px-4 py-3">{arrayListToString(config.association.rowSelectedNames(row))}</td>
                )}
                <td className="px-4 py-3">{arrayListToString(config.rowSupervisors(row))}</td>
                <td className="px-4 py-3">{arrayListToString(config.rowTeams(row))}</td>
                <td className="px-4 py-3 border-box flex space-x-3">
                  <Button text="Изменить" onClick={() => onEditClick(index)} />
                  <Button text="Удалить" buttonType="delete" onClick={() => onDeleteButtonClick(row)} />
                </td>
              </tr>
            ))
          }
          {tableDataQuery.hasNextPage &&
            <tr>
              <td colSpan={8}>
                <div className="w-full py-4 flex justify-center">
                  <div
                    onClick={() => tableDataQuery.fetchNextPage()}
                    className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
                  >
                    {tableDataQuery.isLoading && <LoadingDots height={30} />}
                    {!tableDataQuery.isLoading && "Загрузить еще"}
                  </div>
                </div>
              </td>
            </tr>
          }
        </tbody>
      </table>

      {showDeleteModal &&
        <DeleteModal {...deleteModalProps}>
          <span>При подтверждении бригада под номером {deleteModalProps.no_delivery} и все их данные будут удалены</span>
        </DeleteModal>
      }

      {showMutationModal &&
        <Modal setShowModal={setShowMutationModal}>
          <div>
            {mutationType === "create" && <span className="font-bold text-xl">Добавление: {config.title}</span>}
            {mutationType === "update" && <span className="font-bold text-xl">Изменение: {config.title}</span>}
          </div>
          <div className="flex flex-col space-y-2 py-2">
            <div className="flex flex-col space-y-1">
              <label htmlFor="name">Наименование<span className="text-red-600">*</span></label>
              <Input
                name="name"
                type="text"
                value={mutationData.baseInfo.name}
                onChange={(e) => setMutationData({
                  ...mutationData,
                  baseInfo: { ...mutationData.baseInfo, name: e.target.value },
                })}
              />
            </div>
            <div className="flex flex-col space-y-1">
              <label htmlFor="status">Статус<span className="text-red-600">*</span></label>
              <Select
                className="basic-single text-black"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name="object-status-select"
                placeholder={""}
                value={{ label: mutationData.baseInfo.status, value: mutationData.baseInfo.status }}
                options={OBJECT_STATUSES_FOR_SELECT}
                onChange={(value) => setMutationData({
                  ...mutationData,
                  baseInfo: { ...mutationData.baseInfo, status: value?.value ?? "" },
                })}
              />
            </div>
            <div>
              <label>Супервайзеры объекта</label>
              <Select
                className="basic-single text-black"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                isMulti
                name="supervisors-select"
                placeholder={""}
                value={selectedSupervisors}
                options={availableSupervisors}
                onChange={(value) => {
                  setSelectedSupervisors([...value])
                  setMutationData({ ...mutationData, supervisors: value.map((v) => v.value) })
                }}
              />
            </div>
            <div>
              <label>Бригадиры Объекта</label>
              <Select
                className="basic-single text-black"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                isMulti
                name="teams-select"
                placeholder={""}
                value={selectedTeams}
                options={availableTeams}
                onChange={(value) => {
                  setSelectedTeams([...value])
                  setMutationData({ ...mutationData, teams: value.map((v) => v.value) })
                }}
              />
            </div>
            {config.detailFields.map(renderDetailField)}
            {config.association && renderAssociationField(config.association)}
          </div>
          <div className="mt-4 flex">
            <div
              onClick={onMutationSubmitClick}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
            >
              {(createMutation.isLoading || updateMutation.isLoading) && <LoadingDots height={30} />}
              {!createMutation.isLoading && mutationType === "create" && "Опубликовать"}
              {!updateMutation.isLoading && mutationType === "update" && "Изменить"}
            </div>
          </div>
        </Modal>
      }

      {showImportModal &&
        <Modal setShowModal={setShowImportModal}>
          <span className="font-bold text-xl px-2 py-1">Импорт данных в Справочник</span>
          <div className="grid grid-cols-2 gap-2 items-center px-2 pt-2">
            <div
              onClick={() => importTemplateQuery.refetch()}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
            >
              {importTemplateQuery.fetchStatus === "fetching" ? <LoadingDots height={20} /> : "Скачать шаблон"}
            </div>
            <div className="w-full">
              {importMutation.status === "loading" ? (
                <div className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800">
                  <LoadingDots height={25} />
                </div>
              ) : (
                <label
                  htmlFor="file"
                  className="w-full text-white py-3 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
                >
                  Импортировать данные
                </label>
              )}
              <input
                name="file"
                type="file"
                id="file"
                onChange={acceptExcel}
                className="hidden"
              />
            </div>
          </div>
          <span className="text-sm italic px-2 w-full text-center">При импортировке система будет следовать правилам шаблона</span>
        </Modal>
      }

      {showSearchModal &&
        <Modal setShowModal={setShowSearchModal}>
          <span className="font-bold text-xl py-1">Параметры Поиска</span>
          <div className="p-2 flex flex-col space-y-2">
            <div className="flex flex-col space-y-1">
              <label htmlFor="object-names">Наименование Объекта</label>
              <Select
                className="basic-single"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name="object-names"
                placeholder={""}
                value={selectedSearchObjectName}
                options={allObjectNames}
                onChange={(value) => {
                  const v = value ?? { label: "", value: "" }
                  setSelectedSearchObjectName(v)
                  setSearchParameters(config.setSearchObjectName(searchParameters, v.value))
                }}
              />
            </div>
            <div className="flex flex-col space-y-1">
              <label htmlFor="supervisors">Супервайзеры</label>
              <Select
                className="basic-single"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name="supervisors"
                placeholder={""}
                value={selectedSearchSupervisor}
                options={allSearchSupervisors}
                onChange={(value) => {
                  const v = value ?? { label: "", value: 0 }
                  setSelectedSearchSupervisor(v)
                  setSearchParameters(config.setSearchSupervisor(searchParameters, v.value))
                }}
              />
            </div>
            <div className="flex flex-col space-y-1">
              <label htmlFor="team">Бригада</label>
              <Select
                className="basic-single"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name="team"
                placeholder={""}
                value={selectedSearchTeam}
                options={allSearchTeams}
                onChange={(value) => {
                  const v = value ?? { label: "", value: 0 }
                  setSelectedSearchTeam(v)
                  setSearchParameters(config.setSearchTeam(searchParameters, v.value))
                }}
              />
            </div>
          </div>
        </Modal>
      }
    </main>
  )
}
