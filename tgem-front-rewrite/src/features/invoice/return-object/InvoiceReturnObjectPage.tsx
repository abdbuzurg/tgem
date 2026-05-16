import { useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import { IInvoiceReturnView } from "@entities/invoice-return/types";
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { InvoiceReturnConfirmation, InvoiceReturnPagianted, deleteInvoiceReturn, getInvoiceReturnDocument, getPaginatedInvoiceReturn, sendInvoiceReturnConfirmationExcel } from "@entities/invoice-return/api";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import DeleteModal from "@shared/components/DeleteModal";
import ReportInvoiceReturn from "@entities/invoice-return/ui/ReportInvoiceReturn";
import IconButton from "@shared/components/IconButtons";
import { FaDownload, FaEdit, FaRegListAlt, FaRegTrashAlt, FaUpload } from "react-icons/fa";
import ShowInvoiceReturnDetails from "@entities/invoice-return/ui/ShowInvoiceReturnDetails";
import LoadingDots from "@shared/ui/LoadingDots";
import InvoiceReturnObjectMutationModal from "./InvoiceReturnObjectMutationModal";
import toast from "react-hot-toast";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function InvoiceReturnObject() {

  //FETCHING LOGIC
  const tableDataQuery = useInfiniteQuery<InvoiceReturnPagianted, Error>({
    queryKey: ["invoice-return-object"],
    queryFn: ({ pageParam }) => getPaginatedInvoiceReturn({ pageParam }, "object"),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<IInvoiceReturnView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: IInvoiceReturnView[] = tableDataQuery.data.pages.reduce<IInvoiceReturnView[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //Confirmation logic
  const confirmationFileMutation = useMutation<boolean, Error, InvoiceReturnConfirmation>({
    mutationFn: sendInvoiceReturnConfirmationExcel,
    onSuccess: () => queryClient.invalidateQueries(["invoice-return-object"])
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
        queryClient.invalidateQueries(["invoice-return-object"])
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
    queryKey: ["invoice-return-document", informationToGetDocument],
    queryFn: async () => {
      const loadingToast = toast.loading("Идет скачка файла")
      return getInvoiceReturnDocument(informationToGetDocument.deliveryCode, informationToGetDocument.confirmation)
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
    mutationFn: deleteInvoiceReturn,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-return-object"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: IInvoiceReturnView) => {
    setShowModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.deliveryCode,
      setShowModal: setShowModal,
    })
  }

  //Add LOGIC
  const [showAddModal, setShowAddModal] = useState(false)
  //Edit Logic
  const [showEditModal, setShowEditModal] = useState(false)
  const [rowToEdit, setRowToEdit] = useState<IInvoiceReturnView>()

  //Report Modal
  const [showReportModal, setShowReportModal] = useState(false)

  //DETAIL MODAL DATA
  const [showDetailsModal, setShowDetailsModal] = useState(false)
  const [detailModalData, setDetailModalData] = useState<IInvoiceReturnView>({
    dateOfInvoice: new Date(),
    deliveryCode: "",
    districtName: "",
    id: 0,
    confirmation: false,
    projectID: 1,
    teamNumber: "",
    teamLeaderName: "",
    acceptorName: "",
    objectName: "",
    objectSupervisorNames: [],
    objectType: "",
  })

  const showDetails = (index: number) => {
    setDetailModalData(tableData[index])
    setShowDetailsModal(true)
  }

  return (
    <main>
      <div className="mt-2 px-2 flex justify-between">
        <span className="text-3xl font-bold">Накладные возврат из объекта</span>
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
              <span>Объект</span>
            </th>
            <th className="px-4 py-3">
              <span>Супервайзеры</span>
            </th>
            <th className="px-4 py-3">
              <span>Бригада</span>
            </th>
            <th className="px-4 py-3">
              <span>Бригадир</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Дата</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => setShowAddModal(true)} />
            </th>
          </tr>
        </thead>
        <tbody>
          {tableData.map((value, index) =>
            <tr key={index}>
              <td className="px-4 py-3">{value.deliveryCode}</td>
              <td className="px-4 py-3">{value.objectName}</td>
              <td className="px-4 py-3">{value.objectSupervisorNames.reduce((acc, val) => acc + ", " + val)}</td>
              <td className="px-4 py-3">{value.teamNumber}</td>
              <td className="px-4 py-3">{value.teamLeaderName}</td>
              <td className="px-4 py-3">{value.dateOfInvoice.toString().substring(0, 10)}</td>
              <td className="px-4 py-3 border-box flex space-x-3">
                <IconButton
                  icon={<FaRegListAlt size="20px" title={`Просмотр деталей накладной ${value.deliveryCode}`} />}
                  onClick={() => showDetails(index)}
                />
                {value.confirmation &&
                  <IconButton
                    icon={<FaDownload size="20px" title={`Скачать подтвержденный файл накладной ${value.deliveryCode}`} />}
                    onClick={() => setInformationToGetDocument({deliveryCode: value.deliveryCode, confirmation: value.confirmation})}
                  />
                }
                {!value.confirmation && <>
                  <label
                    htmlFor="file"
                    className="px-4 py-2 flex items-center text-white bg-red-700 hover:bg-red-800 rounded-lg text-center cursor-pointer"
                  >
                    <FaUpload
                      size="20px"
                      title={`Подтвердить файлом накладную ${value.deliveryCode}`}
                    />
                  </label>
                  <input
                    name={`file-${value.id}`}
                    type="file"
                    id="file"
                    onChange={(e) => acceptConfirmationFile(e, index)}
                    className="hidden"
                    value=''
                  />
                  <IconButton
                    icon={<FaDownload size="20px" title={`Скачать сгенерированный файл накладной ${value.deliveryCode}`} />}
                    onClick={() => setInformationToGetDocument({deliveryCode: value.deliveryCode, confirmation: value.confirmation})}
                  />               {/* <IconButton */}
                  {/*   icon={<FaRegEdit size="20px" title={`Изменить данные накладной ${row.deliveryCode}`} />} */}
                  {/*   onClick={() => showDetails(index)} */}
                  {/* /> */}
                  <IconButton
                    type="delete"
                    icon={<FaEdit size="20px" title={`Изменение данных накладной ${value.deliveryCode}`} />}
                    onClick={() => {
                      setRowToEdit(value)
                      setShowEditModal(true)
                    }}
                  />
                  <IconButton
                    type="delete"
                    icon={<FaRegTrashAlt size="20px" title={`Удалить все данные накладной ${value.deliveryCode}`} />}
                    onClick={() => onDeleteButtonClick(value)}
                  />
                </>
                }
              </td>
            </tr>
          )}
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
      {showDetailsModal && <ShowInvoiceReturnDetails setShowModal={setShowDetailsModal} data={detailModalData} />}
      {showModal &&
        <DeleteModal {...modalProps}>
          <span>При подтверждении накладая приход с кодом {modalProps.no_delivery} и все связанные материалы будут удалены</span>
        </DeleteModal>
      }
      {showAddModal && <InvoiceReturnObjectMutationModal mode="create" setShowModal={setShowAddModal} />}
      {showEditModal && <InvoiceReturnObjectMutationModal mode="edit" setShowModal={setShowEditModal} invoiceReturnObject={rowToEdit!} />}
      {showReportModal && <ReportInvoiceReturn setShowReportModal={setShowReportModal} />}
    </main>
  )
}
