import { useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import DeleteModal from "@shared/components/DeleteModal";
import ReportInvoiceOutput from "./ReportInvoiceOutput";
import IconButton from "@shared/components/IconButtons";
import { FaDownload, FaEdit, FaRegListAlt, FaRegTrashAlt, FaUpload } from "react-icons/fa";
import ShowInvoiceOutputInProjectDetails from "./ShowInvoiceOutputInProjectDetails";
import { InvoiceOutputInProjectConfirmation, InvoiceOutputInProjectPagianted, deleteInvoiceOutputInProject, getInvoiceOutputInProjectDocument, getPaginatedInvoiceOutputInProject, sendInvoiceOutputInProjectConfirmationExcel } from "./api";
import { IInvoiceOutputInProjectView } from "./types";
import InvoiceOutputInProjectMutationModal from "./InvoiceOutputInProjectMutationModal";
import LoadingDots from "@shared/ui/LoadingDots";
import toast from "react-hot-toast";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function InvoiceOutputInProject() {
  //FETCHING LOGIC
  const tableDataQuery = useInfiniteQuery<InvoiceOutputInProjectPagianted, Error>({
    queryKey: ["invoice-output-in-project"],
    queryFn: ({ pageParam }) => getPaginatedInvoiceOutputInProject({ pageParam }),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<IInvoiceOutputInProjectView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: IInvoiceOutputInProjectView[] = tableDataQuery.data.pages.reduce<IInvoiceOutputInProjectView[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //Confirmation logic
  const confirmationFileMutation = useMutation<boolean, Error, InvoiceOutputInProjectConfirmation>({
    mutationFn: sendInvoiceOutputInProjectConfirmationExcel,
  })

  const acceptConfirmationFile = (e: React.ChangeEvent<HTMLInputElement>, index: number) => {
    e.preventDefault()

    if (!e.target.files) return

    const confirmationFileToast = toast.loading("Загрузка подтвердающего файла...")

    const fileName = e.target.files[0].name.split(".").pop()
    if (fileName != "pdf") {
      toast.dismiss(confirmationFileToast)
      toast.error("Подтверждающий файл должен быть формата PDF")
      return 
    }

    confirmationFileMutation.mutate({
      id: tableData[index].id,
      file: e.target.files[0]!
    }, {
      onSuccess: () => {
        toast.dismiss(confirmationFileToast)
        toast.success("Подтвердающий файл загружен")
        queryClient.invalidateQueries(["invoice-output-in-project"])
      },
      onError: (err) => {
        toast.dismiss(confirmationFileToast)
        toast.error(`Ошибка при загрузке файла: ${err.message}`)
      },
    })

    e.target.files = null
    e.target.value = ''
  }

  const [informationToGetDocument, setInformationToGetDocument] = useState({
    deliveryCode: "",
    confirmation: false,
  })
  useQuery({
    queryKey: ["invoice-output-document", informationToGetDocument],
    queryFn: async () => {
      const loadingToast = toast.loading("Идет скачка файла")
      return getInvoiceOutputInProjectDocument(informationToGetDocument.deliveryCode, informationToGetDocument.confirmation)
        .then(() => toast.success("Документ скачан"))
        .catch(err => toast.error(`Ошибка при скачке документа: ${err}`))
        .finally(() => toast.dismiss(loadingToast))
    },
    enabled: informationToGetDocument.deliveryCode != "",
  })

  //DELETE LOGIC
  const [showModal, setShowModal] = useState(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteInvoiceOutputInProject,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-output-in-project"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: IInvoiceOutputInProjectView) => {
    setShowModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.deliveryCode,
      setShowModal: setShowModal,
    })
  }

  //Mutation LOGIC
  const [showAddModal, setShowAddModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [rowToEdit, setRowToEdit] = useState<IInvoiceOutputInProjectView>()

  //Details logic
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [detailModalData, setDetailModalData] = useState<IInvoiceOutputInProjectView>({
    dateOfInvoice: new Date(),
    projectID: 0,
    deliveryCode: "",
    id: 0,
    notes: "",
    releasedWorkerID: 0,
    releasedName: "",
    warehouseManagerName: "",
    warehouseManagerWorkerID: 0,
    teamName: "",
    teamID: 0,
    confirmation: false,
    districtName: "",
    districtID: 0,
    recipientName: "",
    recipientWorkerID: 0,
  })

  const showDetails = (index: number) => {
    setDetailModalData(tableData[index])
    setShowDetailsModal(true)
  }

  //Report Modal
  const [showReportModal, setShowReportModal] = useState(false)

  return (
    <main>
      <div className="mt-2 px-2 flex justify-between">
        <span className="text-3xl font-bold">Накладные отпуск внутри проекта</span>
        <div>
          <Button onClick={() => setShowReportModal(true)} text="Отчет" buttonType="default" />
        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[110px]">
              <span>Код</span>
            </th>
            <th className="px-4 py-3">
              <span>Район</span>
            </th>
            <th className="px-4 py-3">
              <span>Бригада</span>
            </th>
            <th className="px-4 py-3 min-w-[110px]">
              <span>Зав. Склад</span>
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
              <td className="px-4 py-3">{row.districtName}</td>
              <td className="px-4 py-3">{row.teamName}</td>
              <td className="px-4 py-3">{row.warehouseManagerName}</td>
              <td className="px-4 py-3">{row.releasedName}</td>
              <td className="px-4 py-3">{row.dateOfInvoice.toString().substring(0, 10)}</td>
              <td className="px-4 py-3 border-box flex space-x-3">
                <IconButton
                  icon={<FaRegListAlt size="20px" title={`Просмотр деталей накладной ${row.deliveryCode}`} />}
                  onClick={() => showDetails(index)}
                />
                {row.confirmation &&
                  <IconButton
                    icon={<FaDownload size="20px" title={`Скачать подтвержденный файл накладной ${row.deliveryCode}`} />}
                    onClick={() => setInformationToGetDocument({deliveryCode: row.deliveryCode, confirmation: row.confirmation})}
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
                      onClick={() => setInformationToGetDocument({deliveryCode: row.deliveryCode, confirmation: row.confirmation})}
                    />               {/* <IconButton */}
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
      {showDetailsModal && <ShowInvoiceOutputInProjectDetails setShowModal={setShowDetailsModal} data={detailModalData} />}
      {showModal &&
        <DeleteModal {...modalProps}>
          <span>При подтверждении накладая уход с кодом {modalProps.no_delivery} и все связанные материалы будут удалены</span>
        </DeleteModal>
      }
      {showAddModal && <InvoiceOutputInProjectMutationModal mode="create" setShowModal={setShowAddModal} />}
      {showEditModal && <InvoiceOutputInProjectMutationModal mode="edit" setShowModal={setShowEditModal} invoiceOutputInProject={rowToEdit!} />}
      {showReportModal && <ReportInvoiceOutput setShowReportModal={setShowReportModal} />}
    </main>
  )
}
