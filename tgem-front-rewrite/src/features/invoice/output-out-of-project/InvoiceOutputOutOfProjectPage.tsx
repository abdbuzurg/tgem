import { useEffect, useState } from "react"
import Select from 'react-select'
import InvoiceOutputOutOfProjectMutationModal from "./InvoiceOutputOutOfProjectMutationModal"
import Button from "@shared/ui/Button"
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { InvoiceOutputOutOfProjectConfirmation, InvoiceOutputOutOfProjectPaginated, InvoiceOutputOutOfProjectSearchParameters, InvoiceOutputOutOfProjectView, deleteInvoiceOutputOutOfProject, getInvoiceOutputOutOfProjectDocument, getInvoiceOutputOutOfProjectPaginated, getUniqueNamesOfProjects, sendInvoiceOutputOutOfProjectConfirmationExcel } from "./api"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import IconButton from "@shared/components/IconButtons"
import { FaDownload, FaEdit, FaRegListAlt, FaRegTrashAlt, FaUpload } from "react-icons/fa"
import LoadingDots from "@shared/ui/LoadingDots"
import DeleteModal from "@shared/components/DeleteModal"
import ShowInvoiceOutputOutOfProjectDetails from "./ShowInvoiceOutputOutOfProjectDetailModal"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Modal from "@shared/components/Modal"
import toast from "react-hot-toast"
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function InvoiceOutputOutOfProject() {

  const [searchParameters, setSearchParameters] = useState<InvoiceOutputOutOfProjectSearchParameters>({
    nameOfProject: "",
    releasedWorkerID: 0,
  })
  //FETCHING LOGIC
  const tableDataQuery = useInfiniteQuery<InvoiceOutputOutOfProjectPaginated, Error>({
    queryKey: ["invoice-output-out-of-project", searchParameters],
    queryFn: ({ pageParam }) => getInvoiceOutputOutOfProjectPaginated({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<InvoiceOutputOutOfProjectView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: InvoiceOutputOutOfProjectView[] = tableDataQuery.data.pages.reduce<InvoiceOutputOutOfProjectView[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //Mutation LOGIC
  const [showAddModal, setShowAddModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [rowToEdit, setRowToEdit] = useState<InvoiceOutputOutOfProjectView>()

  //Details logic
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [detailModalData, setDetailModalData] = useState<InvoiceOutputOutOfProjectView>({
    dateOfInvoice: new Date(),
    deliveryCode: "",
    id: 0,
    releasedWorkerName: "",
    confirmation: false,
    nameOfProject: "",
  })

  const showDetails = (index: number) => {
    setDetailModalData(tableData[index])
    setShowDetailsModal(true)
  }

  //DELETE LOGIC
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteInvoiceOutputOutOfProject,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-output-out-of-project"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowDeleteModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: InvoiceOutputOutOfProjectView) => {
    setShowDeleteModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.deliveryCode,
      setShowModal: setShowDeleteModal,
    })
  }

  //Confirmation logic
  const confirmationFileMutation = useMutation<boolean, Error, InvoiceOutputOutOfProjectConfirmation>({
    mutationFn: sendInvoiceOutputOutOfProjectConfirmationExcel,
    onSuccess: () => queryClient.invalidateQueries(["invoice-output-out-of-project"])
  })

  const acceptConfirmationFile = (e: React.ChangeEvent<HTMLInputElement>, index: number) => {
    e.preventDefault()

    if (!e.target.files) return

    const confirmationFileToast = toast.loading("Загрузка подтвердающего файла...")
    confirmationFileMutation.mutate({
      id: tableData[index].id,
      file: e.target.files[0]!
    }, {
      onSuccess: () => {
        toast.dismiss(confirmationFileToast)
        toast.success("Подтвердающий файл загружен")
        queryClient.invalidateQueries(["invoice-input"])
      },
      onError: (err) => {
        toast.dismiss(confirmationFileToast)
        toast.error(`Ошибка при загрузке файла: ${err.message}`)
      },
    })

    e.target.files = null
    e.target.value = ''
  }

  const [deliveryCodeForDocumentDownload, setDeliveryCodeForDocumentDownload] = useState<string>("")
  useQuery({
    queryKey: ["invoice-output-document", deliveryCodeForDocumentDownload],
    queryFn: async () => {
      const loadingToast = toast.loading("Идет скачка файла")
      return getInvoiceOutputOutOfProjectDocument(deliveryCodeForDocumentDownload)
        .then(() => toast.success("Документ скачан"))
        .catch(err => toast.error(`Ошибка при скачке документа: ${err}`))
        .finally(() => toast.dismiss(loadingToast))
    },
    enabled: deliveryCodeForDocumentDownload != "",
  })

  const [showSearchModal, setShowSearchModal] = useState(false)
  const [selectedNameOfProject, setSelectedNameOfProject] = useState<IReactSelectOptions<string>>({ label: "", value: "", })
  const [allNamesOfProjects, setAllNamesOfProjects] = useState<IReactSelectOptions<string>[]>([])
  const allNamesOfProjectsQuery = useQuery<string[], Error, string[]>({
    queryKey: ["unique-name-of-projects"],
    queryFn: () => getUniqueNamesOfProjects(),
    enabled: false,
  })
  useEffect(() => {
    if (allNamesOfProjectsQuery.isSuccess && allNamesOfProjectsQuery.data) {
      if (allNamesOfProjectsQuery.isSuccess && allNamesOfProjectsQuery.data) {
        setAllNamesOfProjects(allNamesOfProjectsQuery.data.map<IReactSelectOptions<string>>(val => ({
          value: val,
          label: val,
        })))
      }
    }
  }, [allNamesOfProjectsQuery.data, allNamesOfProjectsQuery.isSuccess])

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Накладные отпуск вне проекта</span>
        <div
          onClick={() => {
            setSearchParameters({
              ...searchParameters,
              nameOfProject: "",
            })
            setSelectedNameOfProject({ label: "", value: "" })
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-red-700 hover:bg-red-800 hover:cursor-pointer"
        >
          Сброс поиска
        </div>
        <div>
          <div
            onClick={() => {
              allNamesOfProjectsQuery.refetch()
              setShowSearchModal(true)
            }}
            className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            Поиск
          </div>
          {/* <Button onClick={() => setShowReportModal(true)} text="Отчет" buttonType="default" /> */}

        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[110px]">
              <span>Код</span>
            </th>
            <th className="px-4 py-3">
              <span>Название проекта</span>
            </th>
            <th className="px-4 py-3">
              <span>Составитель</span>
            </th>
            <th className="px-4 py-3 w-[110px]">
              <span>Дата</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => setShowAddModal(true)} />
            </th>
          </tr>
        </thead>
        <tbody>
          {tableData.map((row, index) =>
            <tr key={index} className="text-sm hover:bg-gray-200">
              <td className="px-4 py-3">{row.deliveryCode}</td>
              <td className="px-4 py-3">{row.nameOfProject}</td>
              <td className="px-4 py-3">{row.releasedWorkerName}</td>
              <td className="px-4 py-3">{row.dateOfInvoice.toString().substring(0, 10)}</td>
              <td className="px-4 py-3 border-box flex space-x-3">
                <IconButton
                  icon={<FaRegListAlt size="20px" title={`Просмотр деталей накладной ${row.deliveryCode}`} />}
                  onClick={() => showDetails(index)}
                />
                {row.confirmation &&
                  <IconButton
                    icon={<FaDownload size="20px" title={`Скачать подтвержденный файл накладной ${row.deliveryCode}`} />}
                    onClick={() => setDeliveryCodeForDocumentDownload(row.deliveryCode)}
                  />
                }
                {!row.confirmation &&
                  <>
                    <label
                      htmlFor={`file-${row.id}`}
                      className="px-4 py-2 flex items-center text-white bg-red-700 hover:bg-red-800 rounded-lg text-center cursor-pointer"
                    >
                      <FaUpload
                        size="20px"
                        title={`Подтвердить файлом накладную ${row.deliveryCode}`}
                      />
                    </label>
                    <input
                      name={`file-${row.id}`}
                      type="file"
                      id={`file-${row.id}`}
                      onChange={(e) => acceptConfirmationFile(e, index)}
                      className="hidden"
                      value=''
                    />
                    <IconButton
                      icon={<FaDownload size="20px" title={`Скачать сгенерированный файл накладной ${row.deliveryCode}`} />}
                      onClick={() => setDeliveryCodeForDocumentDownload(row.deliveryCode)}
                    />
                    {/*   icon={<FaRegEdit size="20px" title={`Изменить данные накладной ${row.deliveryCode}`} />} */}
                    {/*   onClick={() => showDetails(index)} */}
                    {/* /> */}
                    <IconButton
                      type="delete"
                      icon={<FaEdit size="20px" title={`Изменение данных накладной ${row.deliveryCode}`} />}
                      onClick={() => {
                        setRowToEdit(row)
                        setShowEditModal(true)
                      }}
                    />
                    <IconButton
                      type="delete"
                      icon={<FaRegTrashAlt size="20px" title={`Удалить все данные накладной ${row.deliveryCode}`} />}
                      onClick={() => onDeleteButtonClick(row)}
                    />
                  </>
                }
              </td>
            </tr>
          )}
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
      {showDetailsModal && <ShowInvoiceOutputOutOfProjectDetails setShowModal={setShowDetailsModal} data={detailModalData} />}
      {showDeleteModal &&
        <DeleteModal {...modalProps}>
          <span>При подтверждении накладая уход с кодом {modalProps.no_delivery} и все связанные материалы будут удалены</span>
        </DeleteModal>
      }
      {showAddModal && <InvoiceOutputOutOfProjectMutationModal mode="create" setShowModal={setShowAddModal} />}
      {showEditModal && <InvoiceOutputOutOfProjectMutationModal mode="edit" setShowModal={setShowEditModal} invoiceOutputOutOfProject={rowToEdit!} />}
      {/* {showReportModal && <ReportInvoiceOutput setShowReportModal={setShowReportModal} />} */}
      {showSearchModal &&
        <Modal setShowModal={setShowSearchModal}>
          <span className="font-bold text-xl py-1">Параметры Поиска по сравочнику материалов</span>
          {allNamesOfProjectsQuery.isLoading && <LoadingDots />}
          {allNamesOfProjectsQuery.isSuccess &&
            <div className="p-2 flex flex-col space-y-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="projects">Проект</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"projects"}
                  placeholder={""}
                  value={selectedNameOfProject}
                  options={allNamesOfProjects}
                  onChange={value => {
                    setSelectedNameOfProject(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      nameOfProject: value?.value ?? "",
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
