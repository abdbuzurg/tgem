import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import getPaginatedMaterialCost, { MaterialCostGetAllResponse } from "@entities/material-cost/api/getAllPaginated"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import { useEffect, useState } from "react"
import { IMaterialCost, IMaterialCostView } from "@entities/material-cost/types"
import createMaterialCost from "@entities/material-cost/api/create"
import deleteMaterialCost from "@entities/material-cost/api/delete"
import updateMaterialCost from "@entities/material-cost/api/update"
import Button from "@shared/ui/Button"
import LoadingDots from "@shared/ui/LoadingDots"
import DeleteModal from "@shared/components/DeleteModal"
import Modal from "@shared/components/Modal"
import Input from "@shared/ui/Input"
import Select from "react-select"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Material from "@entities/material/types"
import getAllMaterials from "@entities/material/api/getAll"
import toast from "react-hot-toast"
import { MaterialCostSearchParameteres, exportMaterialCosts, getMaterialCostTemplateDocument, importMaterialCost } from "@entities/material-cost/api"
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function MaterialsCosts() {
  //fetching data logic
  const [searchParameters, setSearchParameters] = useState<MaterialCostSearchParameteres>({
    materialName: ""
  })
  const tableDataQuery = useInfiniteQuery<MaterialCostGetAllResponse, Error>({
    queryKey: ["materials-costs", searchParameters],
    queryFn: ({ pageParam }) => getPaginatedMaterialCost({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<IMaterialCostView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: IMaterialCostView[] = tableDataQuery.data.pages.reduce<IMaterialCostView[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //mutation DELETE logic
  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteMaterialCost,
    onSuccess: () => {
      queryClient.invalidateQueries(["materials-costs"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowDeleteModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: IMaterialCostView) => {
    setShowDeleteModal(true)
    setModalProps({
      deleteFunc: () => deleteMutation.mutate(row.id),
      no_delivery: row.materialName,
      setShowModal: setShowDeleteModal,
    })
  }

  //mutation CREATE AND EDIT logic
  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allMaterials, setAllMaterials] = useState<IReactSelectOptions<number>[]>([])
  const allMaterialQuery = useQuery<Material[], Error>({
    queryKey: ["material-all"],
    queryFn: getAllMaterials,
  })
  useEffect(() => {
    if (allMaterialQuery.isSuccess && allMaterialQuery.data) {
      setAllMaterials(allMaterialQuery.data.map<IReactSelectOptions<number>>((v) => ({ value: v.id, label: v.name })))
    }
  }, [allMaterialQuery.isSuccess, allMaterialQuery.data])
  const onMaterialSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedMaterial({ value: 0, label: "" })
      setMutationData({ ...mutationData, materialID: 0 })
      return
    }

    setMutationData({ ...mutationData, materialID: value.value })
    setSelectedMaterial(value)
  }

  const [showMutationModal, setShowMutationModal] = useState<boolean>(false)
  const [mutationModalType, setMutationModalType] = useState<null | "update" | "create">()
  const [mutationData, setMutationData] = useState<IMaterialCost>({
    id: 0,
    costPrime: 0,
    costWithCustomer: 0,
    costM19: 0,
    materialID: 0,
  })

  const createMaterialMutation = useMutation<IMaterialCost, Error, IMaterialCost>({
    mutationFn: createMaterialCost,
    onSettled: () => {
      queryClient.invalidateQueries(["materials-costs"])
      setShowMutationModal(false)
    }
  })

  const updateMaterialMutation = useMutation<IMaterialCost, Error, IMaterialCost>({
    mutationFn: updateMaterialCost,
    onSettled: () => {
      queryClient.invalidateQueries(["materials-costs"])
      setShowMutationModal(false)
    }
  })
  const onMutationSubmit = () => {
    if (mutationData.materialID == 0) {
      toast.error("Не указан материал")
      return
    }

    if (mutationData.costM19 <= 0) {
      toast.error("Цена М19 обязательна")
      return
    }

    if (mutationModalType == "create") {
      createMaterialMutation.mutate(mutationData)
    }

    if (mutationModalType == "update") {
      updateMaterialMutation.mutate(mutationData)
    }

    setMutationData({
      id: 0,
      materialID: 0,
      costPrime: 0,
      costM19: 0,
      costWithCustomer: 0,
    })
    setSelectedMaterial({ label: "", value: 0 })

  }

  const onEditClick = (index: number) => {
    const material = allMaterials.find((val) => val.label == tableData[index].materialName)!
    setSelectedMaterial(material)
    setMutationData({
      id: tableData[index].id,
      materialID: material.value,
      costPrime: tableData[index].costPrime,
      costM19: tableData[index].costM19,
      costWithCustomer: tableData[index].costWithCustomer,
    })
    setShowMutationModal(true)
    setMutationModalType("update")
  }

  const [showImportModal, setShowImportModal] = useState(false)

  const importTemplateFile = useQuery<boolean, Error, boolean>({
    queryKey: ["material-import-template"],
    queryFn: getMaterialCostTemplateDocument,
    enabled: false,
    cacheTime: 0,
  })

  const importMutation = useMutation<boolean, Error, File>({
    mutationFn: importMaterialCost,
    cacheTime: 0,
  })

  const acceptExcel = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importMutation.mutate(e.target.files[0], {
      onSuccess: () => {
        queryClient.invalidateQueries(["materials-costs"])
        setShowImportModal(false)
        toast.success("Импортирование успешно")
      },
      onSettled: () => {
        e.target.value = ""
      },
      onError: (error) => {
        toast.error(`Импортированный файл имеет неправильные данные: ${error.message}`)
      }
    })
  }

  const materialExport = useQuery<boolean, Error, boolean>({
    queryKey: ["material-costs-export"],
    queryFn: exportMaterialCosts,
    enabled: false,
    cacheTime: 0,
  })

  const [showSearchModal, setShowSearchModal] = useState(false)
  const allMaterialsQuery = useQuery<Material[], Error, Material[]>({
    queryKey: ["all-materials"],
    queryFn: getAllMaterials,
    enabled: false,
  })

  const [selectedMaterialName, setSelectedMaterialName] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allMaterialNames, setAllMaterialNames] = useState<IReactSelectOptions<string>[]>([])

  useEffect(() => {
    if (allMaterialsQuery.isSuccess && allMaterialsQuery.data) {
      setAllMaterialNames(allMaterialsQuery.data.map((val) => ({
        label: val.name,
        value: val.name,
      })))
    }
  }, [allMaterialsQuery.data, allMaterialsQuery.isSuccess])

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Ценники Материалов</span>
        <div
          onClick={() => {
            allMaterialsQuery.refetch()
            setShowSearchModal(true)
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          Поиск
        </div>
        <Button text="Импорт" onClick={() => setShowImportModal(true)} />
        <div
          onClick={() => materialExport.refetch()}
          className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
        >
          {materialExport.fetchStatus == "fetching" ? <LoadingDots height={20} /> : "Экспорт"}
        </div>
        <div
          onClick={() => {
            setSearchParameters({
              materialName: "",
            })
            setSelectedMaterialName({ label: "", value: "" })
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
              <span>Наименование материала</span>
            </th>
            <th className="px-4 py-3">
              <span>Изначальная цена</span>
            </th>
            <th className="px-4 py-3">
              <span>Цена М19</span>
            </th>
            <th className="px-4 py-3">
              <span>Цена с заказчиком</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => {
                setMutationData({
                  id: 0,
                  materialID: 0,
                  costPrime: 0,
                  costM19: 0,
                  costWithCustomer: 0,
                })
                setSelectedMaterial({ value: 0, label: "" })
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
                <td className="px-4 py-3">{row.materialName}</td>
                <td className="px-4 py-3">{row.costPrime}</td>
                <td className="px-4 py-3">{row.costM19}</td>
                <td className="px-4 py-3">{row.costWithCustomer}</td>
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
      {showDeleteModal &&
        <DeleteModal {...modalProps}>
          <span>При подтверждении материал под именем {modalProps.no_delivery} и все его данные в ней будут удалены</span>
        </DeleteModal>
      }
      {showMutationModal &&
        <Modal setShowModal={setShowMutationModal}>
          <div className="">
            <h3 className="text-xl font-medium text-gray-800">
              {mutationModalType == "create" && "Добавление сервиса"}
              {mutationModalType == "update" && "Изменение сервиса"}
            </h3>
            <div className="flex flex-col space-y-3 mt-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="name">Наименование<span className="text-red-600">*</span></label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-cost-material-select"}
                  placeholder={""}
                  value={selectedMaterial}
                  options={allMaterials}
                  onChange={(value) => onMaterialSelect(value)}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="costPrime">Изначальная Цена</label>
                <Input
                  name="costPrime"
                  type="number"
                  value={mutationData.costPrime}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: +e.target.valueAsNumber })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="costPrime">Цена M19<span className="text-red-600">*</span></label>
                <Input
                  name="costM19"
                  type="number"
                  value={mutationData.costM19}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: +e.target.valueAsNumber })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="costWithCustomer">Цена с заказчиком</label>
                <Input
                  name="costWithCustomer"
                  type="number"
                  value={mutationData.costWithCustomer}
                  onChange={(e) => setMutationData({ ...mutationData, [e.target.name]: +e.target.valueAsNumber })}
                />
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
      {showImportModal &&
        <Modal setShowModal={setShowImportModal}>
          <span className="font-bold text-xl px-2 py-1">Импорт данных в Справочник - Ценники Материалов</span>
          <div className="grid grid-cols-2 gap-2 items-center px-2 pt-2">
            <div
              onClick={() => importTemplateFile.refetch()}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer text-center"
            >
              {importTemplateFile.fetchStatus == "fetching" ? <LoadingDots height={20} /> : "Скачать шаблон"}
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
          <span className="font-bold text-xl py-1">Параметры Поиска по сравочнику материалов</span>
          {allMaterialsQuery.isLoading && <LoadingDots />}
          {allMaterialsQuery.isSuccess &&
            <div className="p-2 flex flex-col space-y-2">
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-names">Наименование</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-names"}
                  placeholder={""}
                  value={selectedMaterialName}
                  options={allMaterialNames}
                  onChange={value => {
                    setSelectedMaterialName(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      materialName: value?.value ?? "",
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
