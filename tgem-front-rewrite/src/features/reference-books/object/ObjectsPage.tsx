import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState, useEffect } from "react"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { IObject, SecondaryObjectData } from "@entities/object/types"
import Input from "@shared/ui/Input"
import Modal from "@shared/components/Modal"
import Button from "@shared/ui/Button"
import LoadingDots from "@shared/ui/LoadingDots"
import DeleteModal from "@shared/components/DeleteModal"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Select from "react-select"
import IWorker from "@entities/worker/types"
import { getWorkerByJobTitle } from "@entities/worker/api"
import toast from "react-hot-toast"
import { IObjectGetAllResponse, IObjectPaginated, ObjectCreateShape, createObject, deleteObject, getPaginatedObjects, updateObject } from "@entities/object/api"
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

const objectTypes = [
  { label: "КЛ 04 КВ", value: "kl04kv_objects" },
  { label: "МЖД", value: "mjd_objects" },
  { label: "СИП", value: "sip_objects" },
  { label: "СТВТ", value: "stvt_objects" },
  { label: "ТП", value: "tp_objects" },
]

export default function Objects() {
  //fetching data logic
  const tableDataQuery = useInfiniteQuery<IObjectGetAllResponse, Error>({
    queryKey: ["objects"],
    queryFn: ({ pageParam = 1 }) => getPaginatedObjects({ pageParam }),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<IObjectPaginated[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: IObjectPaginated[] = tableDataQuery.data.pages.reduce<IObjectPaginated[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //delete logic
  const [showModal, setShowModal] = useState(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteObject,
    onSuccess: () => {
      queryClient.invalidateQueries(["objects"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: IObjectPaginated) => {
    setShowModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.name,
      setShowModal: setShowModal,
    })
  }

  //mutation CREATE AND EDIT logic
  const [showMutationModal, setShowMutationModal] = useState<boolean>(false)
  const [mutationModalType, setMutationModalType] = useState<null | "update" | "create">()

  const [selectedSupervisorsWorkerID, setselectedSupervisorsWorkerID] = useState<IReactSelectOptions<number>[]>([])
  const [avaiableSupervisors, setAvailableSupervisors] = useState<IReactSelectOptions<number>[]>([])
  const supervisorsQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["worker-supervisors"],
    queryFn: () => getWorkerByJobTitle("Супервайзер")
  })
  useEffect(() => {
    if (supervisorsQuery.isSuccess && supervisorsQuery.data) {
      setAvailableSupervisors([
        ...supervisorsQuery.data.map<IReactSelectOptions<number>>((val) => ({ label: val.name, value: val.id }))
      ])
    }
  }, [supervisorsQuery.data, supervisorsQuery.isSuccess])

  const [mutationData, setMutationData] = useState<IObject>({
    projectID: 1,
    id: 0,
    name: "",
    objectDetailedID: 0,
    status: "",
    type: "",
  })
  const [secondaryMutationData, setSecondaryMutation] = useState<SecondaryObjectData>({
    model: "",
    amountStores: 0,
    amountEntrances: 0,
    hasBasement: true,
    voltageClass: "",
    nourashes: "",
    ttCoefficient: "",
    amountFeeders: 0,
    length: 0,
  })
  const [selectedObjectType, setSelectedObjectType] = useState<IReactSelectOptions<string>>({ label: "", value: "" })

  const createMaterialMutation = useMutation<IObject, Error, ObjectCreateShape>({
    mutationFn: createObject,
    onSettled: () => {
      queryClient.invalidateQueries(["objects"])
      setShowMutationModal(false)
    }
  })

  const updateMaterialMutation = useMutation<IObject, Error, IObject>({
    mutationFn: updateObject,
    onSettled: () => {
      queryClient.invalidateQueries(["objects"])
      setShowMutationModal(false)
    }
  })

  const onMutationSubmit = () => {

    if (selectedObjectType.value == "") {
      toast.error("Не указан тип объекта")
      return
    }

    if (mutationData.name == "") {
      toast.error("Не указано имя объекта")
      return
    }

    if (selectedSupervisorsWorkerID.length == 0) {
      toast.error("Не указан супервайзер(-ы)")
      return
    }

    if (mutationData.status == "") {
      toast.error("Не указан статус объекта")
      return
    }

    if ((selectedObjectType.value == "mjd_objects"
      || selectedObjectType.value == "tp_objects")
      && secondaryMutationData.model == "") {
      toast.error("Не указан тип")
      return
    }

    if (selectedObjectType.value == "mjd_objects" && secondaryMutationData.amountStores == 0) {
      toast.error("Не указано кол-во этажей")
      return
    }

    if (selectedObjectType.value == "mjd_objects" && secondaryMutationData.amountEntrances == 0) {
      toast.error("Не указано кол-во подъездов")
      return
    }

    if ((selectedObjectType.value == "tp_objects"
      || selectedObjectType.value == "stvt_objects")
      && secondaryMutationData.voltageClass == "") {
      toast.error("Не указан класс напряжения")
      return
    }

    if ((selectedObjectType.value == "tp_objects"
      || selectedObjectType.value == "kl04kv_objects")
      && secondaryMutationData.nourashes == "") {
      toast.error("Не указано кого питает")
      return
    }

    if (selectedObjectType.value == "stvt_objects" && secondaryMutationData.ttCoefficient == "") {
      toast.error("Не указан коэффицент ТТ")
      return
    }

    if (selectedObjectType.value == "sip_objects" && secondaryMutationData.amountFeeders == 0) {
      toast.error("Не указано количество питающих")
      return
    }

    if (selectedObjectType.value == "kl04kv_objects" && secondaryMutationData.length == 0) {
      toast.error("Не указана длина для КЛ 04 КВ")
      return
    }

    switch (mutationModalType) {
      case "create":
        createMaterialMutation.mutate({
          ...mutationData,
          ...secondaryMutationData,
          supervisors: [...selectedSupervisorsWorkerID.map<number>(val => val.value)],
          type: selectedObjectType.value,
        })

        return
      case "update":
        updateMaterialMutation.mutate(mutationData)
        return

      default:
        throw new Error("Неправильная операция была выбрана")
    }

  }

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Объекты</span>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[150px]">
              <span>Тип</span>
            </th>
            <th className="px-4 py-3">
              <span>Наименование</span>
            </th>
            <th className="px-4 py-3">
              <span>Статус</span>
            </th>
            <th className="px-4 py-3 w-[150px]">
              <span>Супервайзер</span>
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
          {tableDataQuery.isSuccess && tableData.length != 0 &&
            tableData.map((row, index) => (
              <tr key={index} className="border-b">
                <td className="px-4 py-3">{row.type}</td>
                <td className="px-4 py-3">{row.name}</td>
                <td className="px-4 py-3">{row.status}</td>
                <td className="px-4 py-3">
                  {row.supervisors.reduce((acc, value) => acc + ", " + value, "")}
                </td>
                <td className="px-4 py-3 border-box flex space-x-3">
                  <Button text="Изменить" buttonType="default" onClick={() => {
                    setShowMutationModal(true)
                    setMutationModalType("update")
                    // setMutationData({
                    //   id: row.id,
                    //   name: row.name,
                    //   status: row.status,
                    //   type: row.type,
                    // })
                    setSelectedObjectType({ label: row.type, value: row.type })
                  }}
                  />
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
              {mutationModalType == "create" && "Добавление матераила"}
              {mutationModalType == "update" && "Изменение матераила"}
            </h3>
            <div className="flex flex-col space-y-3 mt-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="type">Тип объекта</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"objects"}
                  placeholder={""}
                  value={selectedObjectType}
                  options={objectTypes}
                  onChange={(value) => setSelectedObjectType({
                    label: value?.label ?? "",
                    value: value?.value ?? "",
                  })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="name">Наименование</label>
                <Input
                  name="name"
                  type="text"
                  value={mutationData.name}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div>
                <label htmlFor="">Супервайзеры объекта</label>
                <Select
                  className="basic-single text-black"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  isMulti
                  name={"material-cost-material-select"}
                  placeholder={""}
                  value={selectedSupervisorsWorkerID}
                  options={avaiableSupervisors}
                  onChange={(value) => setselectedSupervisorsWorkerID([...value])}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="status">Статус</label>
                <Input
                  name="status"
                  type="text"
                  value={mutationData.status}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: e.target.value })}
                />
              </div>
              {(selectedObjectType.value == "mjd_objects" || selectedObjectType.value == "tp_objects") &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="model">Тип</label>
                  <Input
                    name="model"
                    type="text"
                    value={secondaryMutationData.model}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: e.target.value })}
                  />
                </div>
              }
              {selectedObjectType.value == "mjd_objects" &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="amountStores">Количество этажей</label>
                  <Input
                    name="amountStores"
                    type="number"
                    value={secondaryMutationData.amountStores}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: +e.target.value })}
                  />
                </div>
              }
              {selectedObjectType.value == "mjd_objects" &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="amountEntrances">Количество подъездов</label>
                  <Input
                    name="amountEntrances"
                    type="number"
                    value={secondaryMutationData.amountEntrances}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: +e.target.value })}
                  />
                </div>
              }
              {selectedObjectType.value == "mjd_objects" &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="amountEntrances">Присутсвует падвал?</label>
                  <div className="flex space-x-3">
                    <div className="flex space-x-1">
                      <input
                        type="radio"
                        name="hasBasement"
                        value={"1"}
                        onChange={() => setSecondaryMutation({ ...secondaryMutationData, hasBasement: true })}
                        checked={secondaryMutationData.hasBasement}
                        id="hasBasementTrue"
                      />
                      <label htmlFor="hasBasementTrue">Да</label>
                    </div>
                    <div className="flex space-x-1">
                      <input
                        type="radio"
                        name="hasBasement"
                        value={"2"}
                        onChange={() => setSecondaryMutation({ ...secondaryMutationData, hasBasement: false })}
                        checked={!secondaryMutationData.hasBasement}
                        id="hasBasementFalse"
                      />
                      <label htmlFor="hasBasementFalse">Нет</label>
                    </div>
                  </div>
                </div>
              }
              {(selectedObjectType.value == "tp_objects" || selectedObjectType.value == "stvt_objects") &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="voltageClass">Класс напряжения</label>
                  <Input
                    name="voltageClass"
                    type="text"
                    value={secondaryMutationData.voltageClass}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: e.target.value })}
                  />
                </div>
              }
              {(selectedObjectType.value == "tp_objects" || selectedObjectType.value == "kl04kv_objects") &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="nourashes">Питает</label>
                  <Input
                    name="nourashes"
                    type="text"
                    value={secondaryMutationData.nourashes}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: e.target.value })}
                  />
                </div>
              }
              {selectedObjectType.value == "stvt_objects" &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="ttCoefficient">ТТ Коэффицент</label>
                  <Input
                    name="ttCoefficient"
                    type="text"
                    value={secondaryMutationData.ttCoefficient}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: e.target.value })}
                  />
                </div>
              }
              {selectedObjectType.value == "sip_objects" &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="amountFeeders">Количество питающих</label>
                  <Input
                    name="amountFeeders"
                    type="number"
                    value={secondaryMutationData.amountFeeders}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: +e.target.value })}
                  />
                </div>
              }
              {selectedObjectType.value == "kl04kv_objects" &&
                <div className="flex flex-col space-y-1">
                  <label htmlFor="length">Длина</label>
                  <Input
                    name="length"
                    type="text"
                    value={secondaryMutationData.length}
                    onChange={(e) => setSecondaryMutation({ ...secondaryMutationData, [e.target.name]: +e.target.value })}
                  />
                </div>
              }
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
