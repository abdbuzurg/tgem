import { useEffect, useState } from "react"
import Input from "@shared/ui/Input"
import Button from "@shared/ui/Button"
import Material from "@entities/material/types"
import { IMaterialCost } from "@entities/material-cost/types"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import getAllMaterials from "@entities/material/api/getAll"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Select from 'react-select'
import { MEASUREMENT } from "@shared/lib/measurement"
import { toast } from 'react-hot-toast'
import { CreateFullMaterial, createNewMaterialCostFromInvoiceInput, createNewMaterialFromInvoiceInput } from "./api"


interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function AddNewMaterialModal({ setShowModal }: Props) {

  const materialQuery = useQuery<Material[], Error, Material[]>({
    queryKey: ["all-materials"],
    queryFn: getAllMaterials,
  })

  const [allData, setAllData] = useState<IReactSelectOptions<number>[]>([])
  const [selectedData, setSelectedData] = useState<IReactSelectOptions<number>>({ value: 0, label: "" })

  useEffect(() => {
    if (materialQuery.isSuccess && materialQuery.data) {
      setAllData([...materialQuery.data.map<IReactSelectOptions<number>>((value) => ({ value: value.id, label: value.name }))])
    }
  }, [materialQuery.data, materialQuery.isSuccess])

  const onSelectChange = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedData({ label: "", value: 0 })
      setMaterialCostData({ ...materialCostData, materialID: 0 })
      return
    }

    setMaterialCostData({ ...materialCostData, materialID: value.value })
    setSelectedData(value)
  }

  const [addType, setAddType] = useState<"newMaterial" | "newMaterialCost" | null | string>(null)
  const onRadioSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAddType(e.target.value)
  }

  const [materialData, setMaterialData] = useState<Material>({
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

  const [materialCostData, setMaterialCostData] = useState<IMaterialCost>({
    costM19: 0,
    costPrime: 0,
    costWithCustomer: 0,
    id: 0,
    materialID: 0,
  })

  const measurements = MEASUREMENT.map<IReactSelectOptions<string>>((value) => ({ label: value, value: value }))
  const [selectedMeasurement, setSelectedMeasurement] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const onMeasurementSelect = (value: null | IReactSelectOptions<string>) => {
    if (!value) {
      setSelectedMeasurement({ label: "", value: "" })
      setMaterialData({ ...materialData, unit: "" })
      return
    }

    setSelectedMeasurement(value)
    setMaterialData({ ...materialData, unit: value.value })
  }

  const queryClient = useQueryClient()
  const newMaterailCostMutation = useMutation<boolean, Error, IMaterialCost>({
    mutationFn: createNewMaterialCostFromInvoiceInput,
    onSettled: () => {
      queryClient.invalidateQueries(["material-cost", materialCostData.materialID])
    }
  })

  const newFullMaterialMutation = useMutation<boolean, Error, CreateFullMaterial>({
    mutationFn: createNewMaterialFromInvoiceInput,
    onSettled: () => {
      queryClient.invalidateQueries(["all-materials"])
    }
  })

  const onSubmit = () => {
    if (addType == "newMaterialCost" && selectedData.value == 0) {
      toast.error("Не выбран материал")
      return
    }

    if (materialCostData.costM19 == 0) {
      toast.error("Не указана цена М19")
      return
    }

    if (addType == "newMaterialCost") {
      newMaterailCostMutation.mutate(materialCostData)
      setShowModal(false)
      return
    }

    if (materialData.name == "") {
      toast.error("Не указано имя материала")
      return
    }

    if (materialData.category == "") {
      toast.error("Не указано категория материла")
      return
    }

    if (materialData.code == "") {
      toast.error("Не указан код материала")
      return
    }

    if (materialData.unit == "") {
      toast.error("Не указано единица измерения материала")
      return
    }

    if (materialData.article == "") {
      toast.error("Не указан пункт материала")
      return
    }

    if (materialData.plannedAmountForProject <= 0) {
      toast.error("Запланированное количество должны быть больше 0")
      return
    }

    if (addType == "newMaterial") {
      newFullMaterialMutation.mutate({
        ...materialData,
        ...materialCostData,
      })
      setShowModal(false)
    }
  }

  return (
    <>
      <div className="fixed inset-0 z-10 overflow-y-auto">
        <div
          className="fixed inset-0 w-full h-full bg-black opacity-40"
        ></div>
        <div className="flex items-center min-h-screen px-4 py-8">
          <div className={`relative w-full px-4 py-2 mx-auto bg-white rounded-md shadow-lg max-w-lg`}>
            <div className="flex justify-end w-full">
              <span
                className="px-3 py-0.5 rounded text-white bg-red-700 hover:bg-red-600 hover:cursor-pointer"
                onClick={() => setShowModal(false)}
              >
                X
              </span>
            </div>

            <div className="flex w-full">
              <div className="text-center sm:text-left w-full">
                <div>
                  <p className="font-bold text-2xl mb-2">Добавление данных материала</p>
                  <div className="flex flex-col space-y-1">
                    <div>
                      <p className="font-semibold text-xl">Что хотите сделать?</p>
                      <div className="flex items-center space-x-2">
                        <div className="flex items-center space-x-2">
                          <input id="newMaterial" name="add_type" type="radio" value="newMaterial" onChange={onRadioSelect} />
                          <label htmlFor="newMaterial">Добавить новый материал</label>
                        </div>
                        <div className="flex items-center space-x-2">
                          <input id="newMaterialCost" name="add_type" type="radio" value="newMaterialCost" onChange={onRadioSelect} />
                          <label htmlFor="newMaterialCost">Добавить новый ценник для материала</label>
                        </div>
                      </div>
                    </div>
                    {addType == "newMaterial" &&
                      <>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="name">Категория материала</label>
                          <Input
                            name="category"
                            type="text"
                            value={materialData.category}
                            onChange={(e) => setMaterialData({ ...materialData, [e.target.name]: e.target.value })}
                          />
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="name">Наименование</label>
                          <Input
                            name="name"
                            type="text"
                            value={materialData.name}
                            onChange={(e) => setMaterialData({ ...materialData, [e.target.name]: e.target.value })}
                          />
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="unit">Еденица измерения</label>
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
                          <label htmlFor="unit">Код</label>
                          <Input
                            name="code"
                            type="text"
                            value={materialData.code}
                            onChange={(e) => setMaterialData({ ...materialData, [e.target.name]: e.target.value })}
                          />
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="unit">Пункт договора закасчика</label>
                          <Input
                            name="article"
                            type="text"
                            value={materialData.article}
                            onChange={(e) => setMaterialData({ ...materialData, [e.target.name]: e.target.value })}
                          />
                        </div>
                        <div className="flex space-x-2">
                          <input type="checkbox" id="hasSerialNumber" value={1} name="hasSerialNumber" onChange={
                            (e) => {
                              if (e.target.checked) {
                                setMaterialData({ ...materialData, hasSerialNumber: true })
                              } else {
                                setMaterialData({ ...materialData, hasSerialNumber: false })
                              }
                            }
                          }
                          />
                          <label htmlFor="hasSerialNumber">Серийный номер</label>
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="code">Запланированное количество<span className="text-red-600">*</span></label>
                          <Input
                            name="plannedAmountForProject"
                            type="number"
                            value={materialData.plannedAmountForProject}
                            onChange={(e) => setMaterialData({ ...materialData, [e.target.name]: e.target.valueAsNumber })}
                          />
                        </div>
                        <div className="flex flex-col space-y-1">
                          <div className="flex space-x-2 items-center">
                            <input type="checkbox" id="hasSerialNumber" value={1} name="hasSerialNumber" onChange={
                              (e) => {
                                if (e.target.checked) {
                                  setMaterialData({ ...materialData, showPlannedAmountInReport: true })
                                } else {
                                  setMaterialData({ ...materialData, showPlannedAmountInReport: false })
                                }
                              }
                            }
                            />
                            <label htmlFor="hasSerialNumber">Показать в отчете "Xод работы проекта"</label>
                          </div>
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="notes">Примичание</label>
                          <textarea
                            name="notes"
                            value={materialData.notes}
                            onChange={(e) => setMaterialData({ ...materialData, [e.target.name]: e.target.value })}
                            className="py-1.5 px-2 resize-none w-full h-[100px] rounded border  border-gray-800"
                          >
                          </textarea>
                        </div>
                      </>
                    }
                    {addType == "newMaterialCost" &&
                      <div>
                        <label htmlFor="materials">Материал</label>
                        <Select
                          className="basic-single"
                          classNamePrefix="select"
                          isSearchable={true}
                          isClearable={true}
                          name={"materials"}
                          placeholder={""}
                          value={selectedData}
                          options={allData}
                          onChange={(value) => onSelectChange(value)}
                        />
                      </div>
                    }
                    {addType &&
                      <>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="costPrime">Изначальная Цена</label>
                          <Input
                            name="costPrime"
                            type="number"
                            value={materialCostData.costPrime}
                            onChange={(e) => setMaterialCostData({ ...materialCostData, [e.target.name]: e.target.valueAsNumber })}
                          />
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="costPrime">Цена M19</label>
                          <Input
                            name="costM19"
                            type="number"
                            value={materialCostData.costM19}
                            onChange={(e) => setMaterialCostData({ ...materialCostData, [e.target.name]: +e.target.valueAsNumber })}
                          />
                        </div>
                        <div className="flex flex-col space-y-1">
                          <label htmlFor="costWithCustomer">Цена с заказчиком</label>
                          <Input
                            name="costWithCustomer"
                            type="number"
                            value={materialCostData.costWithCustomer}
                            onChange={(e) => setMaterialCostData({ ...materialCostData, [e.target.name]: +e.target.valueAsNumber })}
                          />
                        </div>
                        <div className="flex mt-2">
                          <Button text="Добавить" onClick={() => onSubmit()} />
                        </div>
                      </>
                    }
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
