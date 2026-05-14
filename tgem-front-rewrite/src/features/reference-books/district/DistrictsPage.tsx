import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import Modal from "@shared/components/Modal"
import Button from "@shared/ui/Button"
import LoadingDots from "@shared/ui/LoadingDots"
import DeleteModal from "@shared/components/DeleteModal"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { DistrictsPaginatedData, createDistrict, deleteDistrict, getDistrictsPaginated, updateDistrict } from "@entities/district/api"
import { IDistrict } from "@entities/district/types"
import { useEffect, useState } from "react"
import Input from "@shared/ui/Input"

export default function District() {
  //FETCHING LOGIC
  const tableDataQuery = useInfiniteQuery<DistrictsPaginatedData, Error>({
    queryKey: ["districts"],
    queryFn: ({ pageParam }) => getDistrictsPaginated({ pageParam }),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<IDistrict[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: IDistrict[] = tableDataQuery.data.pages.reduce<IDistrict[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  //DELETION LOGIC
  const [showModal, setShowModal] = useState(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteDistrict,
    onSuccess: () => {
      queryClient.invalidateQueries(["districts"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: IDistrict) => {
    setShowModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.name,
      setShowModal: setShowModal,
    })
  }

  // MUTATION LOGIC
  const [showMutationModal, setShowMutationModal] = useState(false)
  const [mutationModalType, setMutationModalType] = useState<null | "update" | "create">()
  const [materialMutationData, setMaterialMutationData] = useState<IDistrict>({
    id: 0,
    name: "",
  })


  const [mutationModalErrors, setMutationModalErrors] = useState({
    name: false,
  })

  const createMaterialMutation = useMutation<IDistrict, Error, IDistrict>({
    mutationFn: createDistrict,
    onSettled: () => {
      queryClient.invalidateQueries(["districts"])
      setShowMutationModal(false)
    }
  })
  const updateMaterialMutation = useMutation<IDistrict, Error, IDistrict>({
    mutationFn: updateDistrict,
    onSettled: () => {
      queryClient.invalidateQueries(["districts"])
      setShowMutationModal(false)
    }
  })

  const onMutationSubmit = () => {
    if (materialMutationData.name == "") setMutationModalErrors((prev) => ({ ...prev, name: true }))
    else setMutationModalErrors((prev) => ({ ...prev, name: false }))

    const isThereError = Object.keys(materialMutationData).some((value) => {
      if (materialMutationData[value as keyof typeof materialMutationData] == "" && value != "id") {
        return true
      }
    })
    if (isThereError) return

    switch (mutationModalType) {
      case "create":
        createMaterialMutation.mutate(materialMutationData)
        return
      case "update":
        updateMaterialMutation.mutate(materialMutationData)
        return

      default:
        throw new Error("Неправильная операция была выбрана")
    }
  }

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Районы</span>
        {/* <Button text="Экспорт" onClick={() => {}}/> */}
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[100px]">
              <span>Имя</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => {
                setMutationModalType("create")
                setShowMutationModal(true)
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
                <td className="px-4 py-3 border-box flex space-x-3">
                  <Button text="Изменить" buttonType="default" onClick={() => {
                    setShowMutationModal(true)
                    setMutationModalType("update")
                    setMaterialMutationData(row)
                  }}
                  />
                  <Button text="Удалить" buttonType="delete" onClick={() => onDeleteButtonClick(row)} />
                </td>
              </tr>
            ))
          }
          {tableDataQuery.hasNextPage &&
            <tr>
              <td colSpan={2}>
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
          <span>При подтверждении материал под именем {modalProps.no_delivery} и все его данные в ней будут удалены</span>
        </DeleteModal>
      }
      {showMutationModal &&
        <Modal setShowModal={setShowMutationModal}>
          <div className="">
            <h3 className="text-xl font-medium text-gray-800">
              {mutationModalType == "create" && "Добавление матераила"}
              {mutationModalType == "update" && "Изменение матераила"}
            </h3>
            <div className="flex flex-col space-y-3 mt-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="name">Наименование</label>
                <Input
                  name="name"
                  type="text"
                  value={materialMutationData.name}
                  onChange={(e) => setMaterialMutationData({ ...materialMutationData, [e.target.name]: e.target.value })}
                />
                {mutationModalErrors.name && <span className="text-red-600 text-sm font-bold">Не указано наименование материала</span>}
              </div>
              <div>
                <Button
                  text={mutationModalType == "create" ? "Добавить" : "Подтвердить изменения"}
                  onClick={onMutationSubmit}
                />
              </div>
            </div>
          </div>
        </Modal>
      }
    </main>
  )
}
