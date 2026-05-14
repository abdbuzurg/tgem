import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, useEffect } from "react";
import { ENTRY_LIMIT } from "@shared/config/pagination";
import getPaginatedMaterials, { MaterialsGetAllResponse } from "@entities/material/api/getPaginated";
import Material from "@entities/material/types";
import Button from "@shared/ui/Button";
import deleteMaterial from "@entities/material/api/delete";
import LoadingDots from "@shared/ui/LoadingDots";
import DeleteModal from "@shared/components/DeleteModal";
import Modal from "@shared/components/Modal";
import Input from "@shared/ui/Input";
import createMaterial from "@entities/material/api/create";
import updateMaterial from "@entities/material/api/update";
import { MEASUREMENT } from "@shared/lib/measurement";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Select from "react-select"
import { MaterialSearchParameters, exportMaterials, getMaterialTemplateDocument, importMaterials } from "@entities/material/api";
import toast from "react-hot-toast";
import { MATERIALS_CATEGORIES_FOR_SELECT } from "@shared/lib/data/objectStatuses";
import getAllMaterials from "@entities/material/api/getAll";
import useScrollPaginated from "@shared/hooks/useScrollPaginated"

export default function Materials() {
  //FETCHING LOGIC

  const [searchParameters, setSearchParameters] = useState<MaterialSearchParameters>({
    name: "",
    category: "",
    code: "",
    unit: "",
  })
  const tableDataQuery = useInfiniteQuery<MaterialsGetAllResponse, Error>({
    queryKey: ["materials", searchParameters],
    queryFn: ({ pageParam }) => getPaginatedMaterials({ pageParam }, searchParameters),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT > lastPage.count) return undefined
      return lastPage.page + 1
    }
  })
  const [tableData, setTableData] = useState<Material[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const data: Material[] = tableDataQuery.data.pages.reduce<Material[]>((acc, page) => [...acc, ...page.data], [])
      setTableData(data)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  //DELETION LOGIC
  const [showModal, setShowModal] = useState(false)
  const queryClient = useQueryClient()
  const deleteMutation = useMutation({
    mutationFn: deleteMaterial,
    onSuccess: () => {
      queryClient.invalidateQueries(["materials"])
    }
  })
  const [modalProps, setModalProps] = useState({
    setShowModal: setShowModal,
    no_delivery: "",
    deleteFunc: () => { }
  })
  const onDeleteButtonClick = (row: Material) => {
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
  const [materialMutationData, setMaterialMutationData] = useState<Material>({
    category: "",
    code: "",
    id: 0,
    name: "",
    notes: "",
    unit: "",
    article: "",
    hasSerialNumber: false,
    plannedAmountForProject: 0,
    showPlannedAmountInReport: false,
  })

  const measurements = MEASUREMENT.map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))
  const [selectedMeasurement, setSelectedMeasurement] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const onMeasurementSelect = (value: null | IReactSelectOptions<string>) => {
    if (!value) {
      setSelectedMeasurement({ label: "", value: "" })
      setMaterialMutationData({ ...materialMutationData, unit: "" })
      return
    }

    setSelectedMeasurement(value)
    setMaterialMutationData({ ...materialMutationData, unit: value.value })
  }

  const createMaterialMutation = useMutation<Material, Error, Material>({
    mutationFn: createMaterial,
    onSuccess: () => {
      queryClient.invalidateQueries(["materials"])
      setShowMutationModal(false)
    }
  })
  const updateMaterialMutation = useMutation<Material, Error, Material>({
    mutationFn: updateMaterial,
    onSuccess: () => {
      queryClient.invalidateQueries(["materials"])
      setShowMutationModal(false)
    }
  })

  const onMutationSubmit = () => {

    if (materialMutationData.category == "") {
      toast.error("Не указана категория материала")
      return
    }

    if (materialMutationData.name == "") {
      toast.error("Не указано наименование материала")
      return
    }

    if (materialMutationData.unit == "") {
      toast.error("Не указана еденица измерения материала")
      return
    }

    if (materialMutationData.code == "") {
      toast.error("Не указан код материала")
      return
    }

    if (materialMutationData.showPlannedAmountInReport && materialMutationData.plannedAmountForProject <= 0) {
      toast.error('Запланированное количество должны быть больше 0 если вы ставите галочку в - Показать в отчете "Xод работы проекта"')
      return
    }

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

  const [showImportModal, setShowImportModal] = useState(false)

  const importTemplateFile = useQuery<boolean, Error, boolean>({
    queryKey: ["material-import-template"],
    queryFn: getMaterialTemplateDocument,
    enabled: false,
    cacheTime: 0,
  })

  const importMutation = useMutation<boolean, Error, File>({
    mutationFn: importMaterials,
    cacheTime: 0,
  })

  const acceptExcel = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return
    importMutation.mutate(e.target.files[0], {
      onSuccess: () => {
        queryClient.invalidateQueries(["materials"])
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
    queryKey: ["material-export"],
    queryFn: exportMaterials,
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

  const [selectedMaterialCategory, setSelectedMaterialCategory] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allMaterialCategories, setAllMaterialCategories] = useState<IReactSelectOptions<string>[]>([])

  const [selectedMaterialCode, setSelectedMaterialCode] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allMaterialCodes, setAllMaterialCodes] = useState<IReactSelectOptions<string>[]>([])

  const [selectedMaterialUnit, setSelectedMaterialUnit] = useState<IReactSelectOptions<string>>({ label: "", value: "" })

  useEffect(() => {
    if (allMaterialsQuery.isSuccess && allMaterialsQuery.data) {
      setAllMaterialNames(allMaterialsQuery.data.map((val) => ({
        label: val.name,
        value: val.name,
      })))

      const categories: string[] = []
      const codes: string[] = []
      allMaterialsQuery.data.forEach((material) => {
        if (!categories.includes(material.category)) categories.push(material.category)
        if (!codes.includes(material.code)) codes.push(material.code)
      })

      setAllMaterialCategories(categories.map(val => ({ label: val, value: val })))
      setAllMaterialCodes(codes.map(val => ({ label: val, value: val })))

    }
  }, [allMaterialsQuery.data, allMaterialsQuery.isSuccess])

  return (
    <main>
      <div className="mt-2 pl-2 flex space-x-2">
        <span className="text-3xl font-bold">Материалы</span>
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
              unit: "",
              name: "",
              category: "",
              code: "",
            })
            setSelectedMaterialUnit({ label: "", value: "" })
            setSelectedMaterialCategory({ label: "", value: "" })
            setSelectedMaterialName({ label: "", value: "" })
            setSelectedMaterialCode({ label: "", value: "" })
          }}
          className="text-white py-2.5 px-5 rounded-lg bg-red-700 hover:bg-red-800 hover:cursor-pointer"
        >
          Сброс поиска
        </div>
      </div>
      <table className="table-auto text-sm text-left mt-2 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-4 py-3 w-[100px]">
              <span>Код</span>
            </th>
            <th className="px-4 py-3 w-[100px]">
              <span>Пункт</span>
            </th>
            <th className="px-4 py-3">
              <span>Категория</span>
            </th>
            <th className="px-4 py-3">
              <span>Наименование</span>
            </th>
            <th className="px-4 py-3 w-[100px]">
              <span>Ед. Изм.</span>
            </th>
            <th className="px-4 py-3">
              <span>Примичание</span>
            </th>
            <th className="px-4 py-3">
              <Button text="Добавить" onClick={() => {
                setMaterialMutationData({
                  category: "",
                  code: "",
                  id: 0,
                  name: "",
                  notes: "",
                  unit: "",
                  article: "",
                  hasSerialNumber: false,
                  plannedAmountForProject: 0,
                  showPlannedAmountInReport: false,
                })
                setSelectedMeasurement({ label: "", value: "" })
                setSelectedMaterialCategory({ label: "", value: "" })
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
                <td className="px-4 py-3">{row.code}</td>
                <td className="px-4 py-3">{row.article}</td>
                <td className="px-4 py-3">{row.category}</td>
                <td className="px-4 py-3">{row.name}</td>
                <td className="px-4 py-3">{row.unit}</td>
                <td className="px-4 py-3">{row.notes}</td>
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
              <td colSpan={7}>
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
                <label htmlFor="name">Категория материала<span className="text-red-600">*</span></label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-cost-material-select"}
                  placeholder={""}
                  value={{
                    label: materialMutationData.category,
                    value: materialMutationData.category,
                  }}
                  options={MATERIALS_CATEGORIES_FOR_SELECT}
                  onChange={(value) => setMaterialMutationData({
                    ...materialMutationData,
                    category: value?.value ?? "",
                  })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="name">Наименование<span className="text-red-600">*</span></label>
                <Input
                  name="name"
                  type="text"
                  value={materialMutationData.name}
                  onChange={(e) => setMaterialMutationData({ ...materialMutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="unit">Еденица измерения<span className="text-red-600">*</span></label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-cost-material-select"}
                  placeholder={""}
                  value={selectedMeasurement}
                  options={measurements}
                  onChange={(value) => onMeasurementSelect(value)}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="code">Код материала<span className="text-red-600">*</span></label>
                <Input
                  name="code"
                  type="text"
                  value={materialMutationData.code}
                  onChange={(e) => setMaterialMutationData({ ...materialMutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="code">Пункт BOQ контракта с заказчиком/УС</label>
                <Input
                  name="article"
                  type="text"
                  value={materialMutationData.article}
                  onChange={(e) => setMaterialMutationData({ ...materialMutationData, [e.target.name]: e.target.value })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <div className="flex space-x-2 items-center">
                  <input
                    type="checkbox"
                    id="hasSerialNumber"
                    value={1}
                    name="hasSerialNumber"
                    checked={materialMutationData.hasSerialNumber}
                    onChange={(e) => setMaterialMutationData({ ...materialMutationData, hasSerialNumber: e.target.checked })}
                  />
                  <label htmlFor="hasSerialNumber">Необходимость контоля на уровне SN</label>
                </div>
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="code">Запланированное количество<span className="text-red-600">*</span></label>
                <Input
                  name="plannedAmountForProject"
                  type="number"
                  value={materialMutationData.plannedAmountForProject}
                  onChange={(e) => setMaterialMutationData({ ...materialMutationData, [e.target.name]: e.target.valueAsNumber })}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <div className="flex space-x-2 items-center">
                  <input
                    type="checkbox"
                    id="showPlannedAmountInReport"
                    value={1}
                    name="showPlannedAmountInReport"
                    checked={materialMutationData.showPlannedAmountInReport}
                    onChange={(e) => setMaterialMutationData({ ...materialMutationData, showPlannedAmountInReport: e.target.checked })}
                  />
                  <label htmlFor="showPlannedAmountInReport">Показать в отчете "Xод работы проекта"</label>
                </div>
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="notes">Примичание</label>
                <textarea
                  name="notes"
                  value={materialMutationData.notes}
                  onChange={(e) => setMaterialMutationData({ ...materialMutationData, [e.target.name]: e.target.value })}
                  className="py-1.5 px-2 resize-none w-full h-[100px] rounded border  border-gray-800"
                >
                </textarea>
              </div>
              <div>
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
          </div>
        </Modal>
      }
      {showImportModal &&
        <Modal setShowModal={setShowImportModal}>
          <span className="font-bold text-xl px-2 py-1">Импорт данных в Справочник - Материалы</span>
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
                      name: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-category">Категория</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-category"}
                  placeholder={""}
                  value={selectedMaterialCategory}
                  options={allMaterialCategories}
                  onChange={value => {
                    setSelectedMaterialCategory(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      category: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-code">Код</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-code"}
                  placeholder={""}
                  value={selectedMaterialCode}
                  options={allMaterialCodes}
                  onChange={value => {
                    setSelectedMaterialCode(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      code: value?.value ?? "",
                    })
                  }}
                />
              </div>
              <div className="flex flex-col space-y-1">
                <label htmlFor="material-unit">Еденица измерения</label>
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  name={"material-unit"}
                  placeholder={""}
                  value={selectedMaterialUnit}
                  options={measurements}
                  onChange={value => {
                    setSelectedMaterialUnit(value ?? { label: "", value: "" })
                    setSearchParameters({
                      ...searchParameters,
                      unit: value?.value ?? "",
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
