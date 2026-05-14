import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import IWorker from "@entities/worker/types";
import { useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import LoadingDots from "@shared/ui/LoadingDots";
import DeleteModal from "@shared/components/DeleteModal";
import Modal from "@shared/components/Modal";
import Input from "@shared/ui/Input";
import Select from 'react-select'
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { JOB_TITLES } from "@shared/lib/data/jobTitles";
import { WorkerInformation, WorkerPaginatedData, WorkerSearchParameters, createWorker, deleteWorker, exportWorker, getPaginatedWorker, getWorkerInformation, getWorkerTemplateDocument, importWorker, updateWorker } from "@entities/worker/api";
import toast from "react-hot-toast";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function Worker() {
  //fetching data logic
  const [searchParameters, setSearchParameters] = useState<WorkerSearchParameters>({
    name: "",
    jobTitleInCompany: "",
    jobTitleInProject: "",
    companyWorkerID: "",
    mobileNumber: "",
  })
  const tableDataQuery = useInfiniteQuery<WorkerPaginatedData, Error>({
    queryKey: ["workers", searchParameters],
    queryFn: ({ pageParam }) => getPaginatedWorker({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<IWorker[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: IWorker[] = tableDataQuery.data.pages.reduce<IWorker[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])
  useScrollPaginated(tableDataQuery.fetchNextPage)

  //mutation DELETE logic
  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteWorker,
    onSuccess: () => {
      queryClient.invalidateQueries(["workers"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowDeleteModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: IWorker) => {
    setShowDeleteModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id.toString()),
      no_delivery: row.name,
      setShowModal: setShowDeleteModal,
    })
  }

  //mutation CREATE AND EDIT logic

  const [jobTitleInTGEM, setJobTitleInTGEM] = useState<IReactSelectOptions<string> | null>({ label: "", value: "" })
  const [jobTitleInProject, setJobTitleInProject] = useState<IReactSelectOptions<string> | null>({ label: "", value: "" })

  const [showMutationModal, setShowMutationModal] = useState<boolean>(false)
  const [mutationModalType, setMutationModalType] = useState<null | "update" | "create">()
  const [workerMutationData, setWorkerMutationData] = useState<IWorker>({
    id: 0,
    jobTitleInProject: "",
    jobTitleInCompany: "",
    companyWorkerID: "",
    mobileNumber: "+992",
    name: "",
  })
  const createMaterialMutation = useMutation<IWorker, Error, IWorker>({
    mutationFn: createWorker,
    onSuccess: () => {
      queryClient.invalidateQueries(["workers"])
      setShowMutationModal(false)
    }
  })
  const updateMaterialMutation = useMutation<IWorker, Error, IWorker>({
    mutationFn: updateWorker,
    onSuccess: () => {
      queryClient.invalidateQueries(["workers"])
      setShowMutationModal(false)
    }
  })
  const onMutationSubmit = () => {

    if (workerMutationData.name == "") {
      toast.error("Не указано имя работника")
      return
    }

    if (workerMutationData.jobTitleInCompany == "") {
      toast.error("Не указано должность рабоника в ТГЭМ")
      return
    }

    if (workerMutationData.jobTitleInProject == "") {
      toast.error("Не указано должность рабоника в проекте")
      return
    }

    if (workerMutationData.mobileNumber == "") {
      toast.error("Не указано мобильный номер работника")
      return
    }

    switch (mutationModalType) {
      case "create":
        createMaterialMutation.mutate(workerMutationData)
        return

      case "update":
        updateMaterialMutation.mutate(workerMutationData)
        return

      default:
        throw new Error("Неправильная операция была выбрана")
    }

  }

  const [showImportModal, setShowImportModal] = useState(false)

  const importTemplateQuery = useQuery<boolean, Error, boolean>({
    queryKey: ["worker-template"],
    queryFn: getWorkerTemplateDocument,
    enabled: false,
  })

  const importMutation = useMutation<boolean, Error, File>({
    mutationFn: importWorker,
  })

  const acceptExcel = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importMutation.mutate(e.target.files[0], {
      onSuccess: () => {
        queryClient.invalidateQueries(["workers"])
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

  const workerExport = useQuery<boolean, Error, boolean>({
    queryKey: ["worker-export"],
    queryFn: exportWorker,
    enabled: false,
    cacheTime: 0,
  })

  const [showSearchModal, setShowSearchModal] = useState(false)

  const [allNamesForSearch, setAllNamesForSearch] = useState<IReactSelectOptions<string>[]>([])
  const [allJobTitleInProjectForSearch, setAllJobTitleInProjectForSearch] = useState<IReactSelectOptions<string>[]>([])
  const [allJobTitleInCompanyForSearch, setAllJobTitleInCompanyForSearch] = useState<IReactSelectOptions<string>[]>([])
  const [allCompanyWorkerIDsForSearch, setAllCompanyWorkerIDsForSearch] = useState<IReactSelectOptions<string>[]>([])
  const [allMobileNumberForSearch, setAllMobileNumberForSearch] = useState<IReactSelectOptions<string>[]>([])
  const workerInformationQuery = useQuery<WorkerInformation, Error, WorkerInformation>({
    queryKey: ["worker-information"],
    queryFn: getWorkerInformation,
    enabled: showSearchModal,
  })
  useEffect(() => {
    if (workerInformationQuery.isSuccess && workerInformationQuery.data) {
      setAllNamesForSearch(workerInformationQuery.data.name.map(val => ({ label: val, value: val })))
      setAllJobTitleInProjectForSearch(workerInformationQuery.data.jobTitleInProject.map(val => ({ label: val, value: val })))
      setAllJobTitleInCompanyForSearch(workerInformationQuery.data.jobTitleInCompany.map(val => ({ label: val, value: val })))
      setAllCompanyWorkerIDsForSearch(workerInformationQuery.data.companyWorkerID.map(val => ({ label: val, value: val })))
      setAllMobileNumberForSearch(workerInformationQuery.data.mobileNumber.map(val => ({ label: val, value: val })))
    }
  }, [workerInformationQuery.data, workerInformationQuery.isSuccess])


  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Рабочий Персонал</span>
        <div
          onClick={() => {
            setShowSearchModal(true)
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          Поиск
        </div>
        <Button text="Импорт" onClick={() => setShowImportModal(true)} />
        <div
          onClick={() => workerExport.refetch()}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          {workerExport.fetchStatus == "fetching" ? <LoadingDots height={20} /> : "Экспорт"}
        </div>
        <div
          onClick={() => {
            setSearchParameters({
              name: "",
              jobTitleInCompany: "",
              jobTitleInProject: "",
              companyWorkerID: "",
              mobileNumber: "",
            })
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-red-700 hover:bg-red-800 hover:cursor-pointer"
        >
          Сброс поиска
        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[350px]">
              <span>Имя</span>
            </th>
            <th className="px-4 py-3">
              <span>Должность в ТГЭМ</span>
            </th>
            <th className="px-4 py-3">
              <span>ID сотруника</span>
            </th>
            <th className="px-4 py-3">
              <span>Должность в проекте</span>
            </th>
            <th className="px-4 py-3">
              <span>Номера</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => {
                setMutationModalType("create")
                setShowMutationModal(true)
                setWorkerMutationData({
                  id: 0,
                  name: "",
                  jobTitleInCompany: "",
                  companyWorkerID: "",
                  jobTitleInProject: "",
                  mobileNumber: "+992"
                })
                setJobTitleInTGEM({ label: "", value: "" })
                setJobTitleInProject({ label: "", value: "" })
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
                <td className="px-4 py-3">{row.name}</td>
                <td className="px-4 py-3">{row.jobTitleInCompany}</td>
                <td className="px-4 py-3">{row.companyWorkerID}</td>
                <td className="px-4 py-3">{row.jobTitleInProject}</td>
                <td className="px-4 py-3">{row.mobileNumber}</td>
                <td className="px-4 py-3 border-box flex space-x-3">
                  <Button text="Изменить" buttonType="default" onClick={() => {
                    setShowMutationModal(true)
                    setMutationModalType("update")
                    setWorkerMutationData(row)
                    setJobTitleInProject({ label: row.jobTitleInProject, value: row.jobTitleInProject })
                    setJobTitleInTGEM({ label: row.jobTitleInCompany, value: row.jobTitleInCompany })
                  }}
                  />
                  <Button text="Удалить" buttonType="delete" onClick={() => onDeleteButtonClick(row)} />
                </td>
              </tr>
            ))
          }
          {tableDataQuery.hasNextPage &&
            <tr>
              <td colSpan={6}>
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
        <DeleteModal {...modalProps}>
          <span>При подтверждении материал под именем {modalProps.no_delivery} и все его данные в ней будут удалены</span>
        </DeleteModal>
      }
      {showMutationModal &&
        <Modal setShowModal={setShowMutationModal}>
          <div className="">
            <h3 className="text-xl font-medium text-gray-800">
              {mutationModalType == "create" && "Добавление работника"}
              {mutationModalType == "update" && "Изменение работника"}
            </h3>
            <div className="flex flex-col space-y-3 mt-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="name">Фамилия Имя Отчество<span className="text-red-600">*</span></label>
                <Input
                  name="name"
                  type="text"
                  value={workerMutationData.name}
                  onChange={(e) => setWorkerMutationData({ ...workerMutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="jobTitle">Должность в ТГЭМ<span className="text-red-600">*</span></label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"job-title-in-company"}
                  placeholder={""}
                  value={jobTitleInTGEM}
                  options={JOB_TITLES.map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))}
                  onChange={(value) => {
                    setJobTitleInTGEM(value)
                    setWorkerMutationData({
                      ...workerMutationData,
                      jobTitleInCompany: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="mobileNumber">ID сотрудника<span className="text-red-600">*</span></label>
                <Input
                  name="companyWorkerID"
                  type="text"
                  value={workerMutationData.companyWorkerID}
                  onChange={(e) => setWorkerMutationData({ ...workerMutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="jobTitle">Должность в Проекте<span className="text-red-600">*</span></label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"job-title-in-project"}
                  placeholder={""}
                  value={jobTitleInProject}
                  options={JOB_TITLES.map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))}
                  onChange={(value) => {
                    setJobTitleInProject(value)
                    setWorkerMutationData({
                      ...workerMutationData,
                      jobTitleInProject: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="mobileNumber">Номера телефона<span className="text-red-600">*</span></label>
                <Input
                  name="mobileNumber"
                  type="text"
                  value={workerMutationData.mobileNumber}
                  onChange={(e) => setWorkerMutationData({ ...workerMutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="mt-4 flex">
                <div
                  onClick={() => onMutationSubmit()}
                  className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
                >
                  {(createMaterialMutation.isLoading || updateMaterialMutation.isLoading) && <LoadingDots height={30} />}
                  {!createMaterialMutation.isLoading && mutationModalType == "create" && "Опубликовать"}
                  {!updateMaterialMutation.isLoading && mutationModalType == "update" && "Изменить"}
                </div>
              </div>
            </div>
          </div>
        </Modal>
      }
      {showImportModal &&
        <Modal setShowModal={setShowImportModal}>
          <span className="font-bold text-xl px-2 py-1">Импорт данных в Справочник - Персонал</span>
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
          <span className="font-bold text-xl py-1">Параметры Поиска по сравочнику - Рабочий Персонал</span>
          {workerInformationQuery.isLoading && <LoadingDots />}
          {workerInformationQuery.isSuccess &&
            <div className="p-2 flex flex-col space-y-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-names">Имя</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-names"}
                  placeholder={""}
                  value={{ label: searchParameters.name, value: searchParameters.name }}
                  options={allNamesForSearch}
                  onChange={value => {
                    setSearchParameters({
                      ...searchParameters,
                      name: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-category">Должность в ТГЭМ</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-category"}
                  placeholder={""}
                  value={{ label: searchParameters.jobTitleInCompany, value: searchParameters.jobTitleInCompany }}
                  options={allJobTitleInCompanyForSearch}
                  onChange={value => {
                    setSearchParameters({
                      ...searchParameters,
                      jobTitleInCompany: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-code">ID Сотрудника</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-code"}
                  placeholder={""}
                  value={{ label: searchParameters.companyWorkerID, value: searchParameters.companyWorkerID }}
                  options={allCompanyWorkerIDsForSearch}
                  onChange={value => {
                    setSearchParameters({
                      ...searchParameters,
                      companyWorkerID: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-unit">Должность в Проекте</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-unit"}
                  placeholder={""}
                  value={{ label: searchParameters.jobTitleInProject, value: searchParameters.jobTitleInProject }}
                  options={allJobTitleInProjectForSearch}
                  onChange={value => {
                    setSearchParameters({
                      ...searchParameters,
                      jobTitleInProject: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-unit">Номера</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-unit"}
                  placeholder={""}
                  value={{ label: searchParameters.mobileNumber, value: searchParameters.mobileNumber }}
                  options={allMobileNumberForSearch}
                  onChange={value => {
                    setSearchParameters({
                      ...searchParameters,
                      mobileNumber: value?.value ?? "",
                    })
                  }}
                />
              </div>
            </div>
          }
        </Modal>
      }
    </main>
  )
}
