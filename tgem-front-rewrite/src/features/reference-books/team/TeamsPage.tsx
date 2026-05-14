import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, useEffect } from "react";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import Button from "@shared/ui/Button";
import LoadingDots from "@shared/ui/LoadingDots";
import DeleteModal from "@shared/components/DeleteModal";
import { ITeam } from "@entities/team/types";
import Modal from "@shared/components/Modal";
import Input from "@shared/ui/Input";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Select from 'react-select'
import toast from "react-hot-toast";
import IWorker from "@entities/worker/types";
import { getAllWorkers } from "@entities/worker/api";
import { TeamMutation, TeamGetAllResponse, TeamPaginated, createTeam, deleteTeam, getPaginatedTeams, updateTeam, importTeam, getTeamTemplateDocument, TeamSearchParameters, exportTeam, getAllUniqueNumber, getAllUniqueMobileNumber, getAllUniqueCompany } from "@entities/team/api";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function Team() {
  //fetching data logic
  const [searchParameters, setSearchParameters] = useState<TeamSearchParameters>({
    number: "",
    leaderID: 0,
    mobileNumber: "",
    company: "",
  })
  const [tableData, setTableData] = useState<TeamPaginated[]>([])
  const tableDataQuery = useInfiniteQuery<TeamGetAllResponse, Error>({
    queryKey: ["teams", searchParameters],
    queryFn: ({ pageParam }) => getPaginatedTeams({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: TeamPaginated[] = tableDataQuery.data.pages.reduce<TeamPaginated[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //delete logic
  const [showModal, setShowModal] = useState(false)
  const queryClient = useQueryClient()

  const deleteMutation = useMutation({
    mutationFn: deleteTeam,
    onSuccess: () => {
      queryClient.invalidateQueries(["teams"])
    }
  })

  const [modalProps, setModalProps] = useState({
    setShowModal: setShowModal,
    no_delivery: "",
    deleteFunc: () => { }
  })

  const onDeleteButtonClick = (row: TeamPaginated) => {
    setShowModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.number,
      setShowModal: setShowModal,
    })
  }

  //mutation CREATE AND EDIT logic
  const [showMutationModal, setShowMutationModal] = useState<boolean>(false)
  const [mutationModalType, setMutationModalType] = useState<null | "update" | "create">()

  const [mutationData, setMutationData] = useState<TeamMutation>({
    company: "",
    id: 0,
    leaderIDs: [],
    mobileNumber: "",
    number: "",
  })

  const [selectedTeamLeaderWorkerID, setSelectedTeamLeaderWorkerID] = useState<IReactSelectOptions<number>[]>([])
  const [availableTeamLeaders, setAvailableTeamLeaders] = useState<IReactSelectOptions<number>[]>([])
  const allWorkersQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["all-workers"],
    queryFn: () => getAllWorkers(),
  })
  useEffect(() => {
    if (allWorkersQuery.data && allWorkersQuery.isSuccess) {
      setAvailableTeamLeaders([
        ...allWorkersQuery.data.map<IReactSelectOptions<number>>((val) => ({
          label: val.name,
          value: val.id,
        }))
      ])
    }

  }, [allWorkersQuery.data, allWorkersQuery.isSuccess])

  const createTeamMutation = useMutation<ITeam, Error, TeamMutation>({
    mutationFn: createTeam,
    onSuccess: () => {
      queryClient.invalidateQueries(["teams"])
      setShowMutationModal(false)
    }
  })

  const updateTeamMutation = useMutation<ITeam, Error, TeamMutation>({
    mutationFn: updateTeam,
    onSuccess: () => {
      queryClient.invalidateQueries(["teams"])
      setShowMutationModal(false)
    }
  })

  const onEditClick = (index: number) => {

    const leaders = tableData[index].leaderNames.map<IReactSelectOptions<number>>((value) => {
      const subIndex = availableTeamLeaders.findIndex((val) => val.label == value)!
      return availableTeamLeaders[subIndex]
    }).filter((val) => val)!
    setSelectedTeamLeaderWorkerID(leaders)

    setMutationData({
      id: tableData[index].id,
      company: tableData[index].company,
      mobileNumber: tableData[index].mobileNumber,
      number: tableData[index].number,
      leaderIDs: leaders.map(v => v.value),
    })

    setShowMutationModal(true)
    setMutationModalType("update")
  }

  const onMutationSubmit = () => {

    if (!mutationData.number) {
      toast.error("Не указано номер бригады")
      return
    }

    if (selectedTeamLeaderWorkerID.length == 0) {
      toast.error("Не указан лидер бригады")
      return
    }

    if (!mutationData.mobileNumber) {
      toast.error("Не указан номер телефона бригады")
      return
    }

    switch (mutationModalType) {
      case "create":

        createTeamMutation.mutate({
          ...mutationData,
          leaderIDs: [...selectedTeamLeaderWorkerID.map(v => v.value)],
        })
        return
      case "update":

        updateTeamMutation.mutate({
          ...mutationData,
          leaderIDs: [...selectedTeamLeaderWorkerID.map(v => v.value)],
        })
        return

      default:
        throw new Error("Неправильная операция была выбрана")
    }
  }

  const [showImportModal, setShowImportModal] = useState(false)

  const importTemplateQuery = useQuery<boolean, Error, boolean>({
    queryKey: ["worker-template"],
    queryFn: getTeamTemplateDocument,
    enabled: false,
  })

  const importMutation = useMutation<boolean, Error, File>({
    mutationFn: importTeam,
    onError: (err: Error) => toast.error(`${err.message}`)
  })

  const acceptExcel = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importMutation.mutate(e.target.files[0], {
      onSuccess: () => {
        queryClient.invalidateQueries(["teams"])
        setShowImportModal(false)
      },
      onSettled: () => {
        e.target.value = ""
      },
      onError: (error) => {
        toast.error(`Импортированный файл имеет неправильные данные: ${error.message}`)
      }
    })
  }

  const teamExport = useQuery<boolean, Error, boolean>({
    queryKey: ["team-export"],
    queryFn: exportTeam,
    enabled: false,
    cacheTime: 0,
  })

  const [showSearchModal, setShowSearchModal] = useState(false)

  const [selectedTeamNumber, setSelectedTeamNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allTeamNumbers, setAllTeamNumbers] = useState<IReactSelectOptions<string>[]>([])
  const uniqueTeamNumberQuery = useQuery<string[], Error, string[]>({
    queryKey: ["unique-team-number"],
    queryFn: getAllUniqueNumber,
    enabled: false,
  })
  useEffect(() => {
    if (uniqueTeamNumberQuery.isSuccess && uniqueTeamNumberQuery.data) {
      setAllTeamNumbers(uniqueTeamNumberQuery.data.map<IReactSelectOptions<string>>(val => ({
        value: val,
        label: val,
      })))
    }
  }, [uniqueTeamNumberQuery.data, uniqueTeamNumberQuery.isSuccess])

  const [selectedTeamLeaderID, setSelectedTeamLeaderID] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })

  const [selectedMobileNumber, setSelectedMobileNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allMobileNumbers, setAllMobileNumbers] = useState<IReactSelectOptions<string>[]>([])
  const uniqueTeamMobileNumbersQuery = useQuery<string[], Error, string[]>({
    queryKey: ["unique-mobile-number"],
    queryFn: getAllUniqueMobileNumber,
    enabled: false,
  })
  useEffect(() => {
    if (uniqueTeamMobileNumbersQuery.isSuccess && uniqueTeamMobileNumbersQuery.data) {
      setAllMobileNumbers(uniqueTeamMobileNumbersQuery.data.map<IReactSelectOptions<string>>(val => ({
        value: val,
        label: val,
      })))
    }
  }, [uniqueTeamMobileNumbersQuery.data, uniqueTeamMobileNumbersQuery.isSuccess])

  const [selectedCompany, setSelectedCompany] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allCompanies, setAllCompanies] = useState<IReactSelectOptions<string>[]>([])
  const uniqueTeamCompaniesQuery = useQuery<string[], Error, string[]>({
    queryKey: ["unique-companies"],
    queryFn: getAllUniqueCompany,
    enabled: false,
  })
  useEffect(() => {
    if (uniqueTeamCompaniesQuery.isSuccess && uniqueTeamCompaniesQuery.data) {
      setAllCompanies(uniqueTeamCompaniesQuery.data.map<IReactSelectOptions<string>>(val => ({
        value: val,
        label: val,
      })))
    }
  }, [uniqueTeamCompaniesQuery.data, uniqueTeamCompaniesQuery.isSuccess])

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Бригады</span>
        <div
          onClick={() => {
            uniqueTeamNumberQuery.refetch()
            uniqueTeamMobileNumbersQuery.refetch()
            uniqueTeamCompaniesQuery.refetch()
            setShowSearchModal(true)
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          Поиск
        </div>
        <Button text="Импорт" onClick={() => setShowImportModal(true)} />
        <div
          onClick={() => teamExport.refetch()}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          {teamExport.fetchStatus == "fetching" ? <LoadingDots height={20} /> : "Экспорт"}
        </div>
        <div
          onClick={() => {
            setSearchParameters({
              number: "",
              mobileNumber: "",
              company: "",
              leaderID: 0,
            })
            setSelectedTeamNumber({ label: "", value: "" })
            setSelectedMobileNumber({ label: "", value: "" })
            setSelectedMobileNumber({ label: "", value: "" })
            setSelectedTeamLeaderID({ label: "", value: 0 })
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-red-700 hover:bg-red-800 hover:cursor-pointer"
        >
          Сброс поиска
        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[150px]">
              <span>Бригада №</span>
            </th>
            <th className="px-4 py-3">
              <span>Бригадир</span>
            </th>
            <th className="px-4 py-3">
              <span>Телефон</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Компания</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => {
                setMutationModalType("create")
                setShowMutationModal(true)
                setMutationData({
                  id: 0,
                  company: "",
                  mobileNumber: "",
                  number: "",
                  leaderIDs: [],
                })
                setSelectedTeamLeaderWorkerID([])
              }} />
            </th>
          </tr>
        </thead>
        <tbody>
          {tableDataQuery.isLoading &&
            <tr>
              <td colSpan={6}>
                <LoadingDots />
              </td>
            </tr>
          }
          {tableDataQuery.isError &&
            <tr>
              <td colSpan={6} className="text-red font-bold text-center">
                {tableDataQuery.error.message}
              </td>
            </tr>
          }
          {tableDataQuery.isSuccess &&
            tableData.map((row, index) => (
              <tr key={index} className="border-b">
                <td className="px-4 py-3">{row.number}</td>
                <td className="px-4 py-3">{row.leaderNames.reduce((acc, val) => acc + ", " + val)}</td>
                <td className="px-4 py-3">{row.mobileNumber}</td>
                <td className="px-4 py-3">{row.company}</td>
                <td className="px-4 py-3 border-box flex space-x-3">
                  <Button text="Изменить" buttonType="default" onClick={() => onEditClick(index)} />
                  <Button text="Удалить" buttonType="delete" onClick={() => onDeleteButtonClick(row)} />
                </td>
              </tr>
            ))
          }
          {tableDataQuery.hasNextPage &&
            <tr>
              <td colSpan={5}>
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
      {showModal &&
        <DeleteModal {...modalProps}>
          <span>При подтверждении бригада под номером {modalProps.no_delivery} и все их данные будут удалены</span>
        </DeleteModal>
      }
      {showMutationModal &&
        <Modal setShowModal={setShowMutationModal}>
          <div className="">
            <h3 className="text-xl font-medium text-gray-800">
              {mutationModalType == "create" && "Добавление бригады"}
              {mutationModalType == "update" && "Изменение бригады"}
            </h3>
            <div className="flex flex-col space-y-3 mt-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="number">Номер Бригады<span className="text-red-600">*</span></label>
                <Input
                  name="number"
                  type="text"
                  value={mutationData.number}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label>Бригадир<span className="text-red-600">*</span></label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  isMulti
                  name={"job-title"}
                  placeholder={""}
                  value={selectedTeamLeaderWorkerID}
                  options={availableTeamLeaders}
                  onChange={(value) => setSelectedTeamLeaderWorkerID([...value])}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="mobileNumber">Мобильный номер бригадира<span className="text-red-600">*</span></label>
                <Input
                  name="mobileNumber"
                  type="text"
                  value={mutationData.mobileNumber}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="company">Компания (если субподрядчик)</label>
                <Input
                  name="company"
                  type="text"
                  value={mutationData.company}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex">
                <div className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer" onClick={() => onMutationSubmit()}>
                  {mutationModalType == "create" && !createTeamMutation.isLoading && "Добавить"}
                  {mutationModalType == "create" && createTeamMutation.isLoading && <LoadingDots height={30} />}
                  {mutationModalType == "update" && !updateTeamMutation.isLoading && "Подтвердить изменения"}
                  {mutationModalType == "update" && updateTeamMutation.isLoading && <LoadingDots height={30} />}
                </div>
              </div>
            </div>
          </div>
        </Modal>
      }
      {showImportModal &&
        <Modal setShowModal={setShowImportModal}>
          <span className="font-bold text-xl px-2 py-1">Импорт данных в Справочник - Бригады</span>
          <div className="grid grid-cols-2 gap-2 items-center px-2 pt-2">
            <div
              onClick={() => importTemplateQuery.refetch()}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
            >
              {importTemplateQuery.fetchStatus == "fetching" ? <LoadingDots height={20} /> : "Скачать шаблон"}
            </div>
            <div className="w-full">
              {importMutation.status == "loading"
                ?
                <div className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800">
                  <LoadingDots height={25} />
                </div>
                :
                <label
                  htmlFor="file"
                  className="w-full text-white py-3 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
                >
                  Импортировать данные
                </label>
              }
              <input
                name="file"
                type="file"
                id="file"
                onChange={(e) => acceptExcel(e)}
                className="hidden"
              />
            </div>
          </div>
          <span className="text-sm italic px-2 w-full text-center">При импортировке система будет следовать правилам шаблона</span>
        </Modal>
      }
      {showSearchModal &&
        <Modal setShowModal={setShowSearchModal}>
          <span className="font-bold text-xl py-1">Параметры Поиска по сравочнику бригад</span>
            <div className="p-2 flex flex-col space-y-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-names">Номер Бригады</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"team-number"}
                  placeholder={""}
                  value={selectedTeamNumber}
                  options={allTeamNumbers}
                  onChange={value => {
                    setSelectedTeamNumber(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      number: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-category">Бригадир</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"team-leader-select"}
                  placeholder={""}
                  value={selectedTeamLeaderID}
                  options={availableTeamLeaders}
                  onChange={value => {
                    setSelectedTeamLeaderID(value ?? { label: "", value: 0 })
                    setSearchParameters({
                      ...searchParameters,
                      leaderID: value?.value ?? 0,
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-code">Телефон</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"team-mobile-number-select"}
                  placeholder={""}
                  value={selectedMobileNumber}
                  options={allMobileNumbers}
                  onChange={value => {
                    setSelectedMobileNumber(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      mobileNumber: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-unit">Компания</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"team-company-select"}
                  placeholder={""}
                  value={selectedCompany}
                  options={allCompanies}
                  onChange={value => {
                    setSelectedCompany(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      company: value?.value ?? "",
                    })
                  }}
                />
              </div>
            </div>
        </Modal>
      }
    </main>
  )
}
