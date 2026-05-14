import { Fragment, useEffect, useState } from "react"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Select from 'react-select'
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Button from "@shared/ui/Button"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { TeamDataForSelect } from "@entities/team/types"
import { IObject } from "@entities/object/types"
import { IDistrict } from "@entities/district/types";
import { IInvoiceObjectMaterials, IInvoiceObjectOperations } from "@entities/invoice/types/invoiceObject"
import toast from "react-hot-toast"
import { getAllObjects } from "@entities/object/api"
import { getAllDistricts } from "@entities/district/api";
import { InvoiceObjectCreateItems, InvoiceObjectOperations, InvoiceObjectOperationsCreate, InvoiceObjectTeamMaterialData, createInvoiceObject, getMaterialsDataFromTeam, getOperationsBasedOnTeamID, getSerialNumbersOfMaterial, getTeamsFromObjectID } from "@features/invoice/object/api"
import { useNavigate } from "react-router-dom"
import { INVOICE_OBJECT_USER } from "@routes/paths"
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses"
import LoadingDots from "@shared/ui/LoadingDots"

export default function InvoiceObjectMutationAdd() {
  const [invoiceDate, setInvoiceDate] = useState<Date>(new Date())

  // Select District Logic
  const [selectedDistrict, setSelectedDistrict] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [availableDistricts, setAvailableDistricts] = useState<IReactSelectOptions<number>[]>([])
  const allDistrictsQuery = useQuery<IDistrict[], Error, IDistrict[]>({
    queryKey: ["all-districts"],
    queryFn: getAllDistricts,
  })
  useEffect(() => {
    if (allDistrictsQuery.isSuccess && allDistrictsQuery.data) {
      setAvailableDistricts(
        allDistrictsQuery.data.map<IReactSelectOptions<number>>((value) => ({
          label: value.name,
          value: value.id,
        }))
      )
    }
  }, [allDistrictsQuery.data, allDistrictsQuery.isSuccess])

  // Select Object Logic
  const [selectedObject, setSelectedObject] = useState<IReactSelectOptions<number>>({
    label: "",
    value: 0,
  })
  const [availableObjects, setAvailableObject] = useState<IReactSelectOptions<number>[]>([])
  const allObjectsQuery = useQuery<IObject[], Error, IObject[]>({
    queryKey: ["all-objects"],
    queryFn: getAllObjects,
  })
  useEffect(() => {
    if (allObjectsQuery.isSuccess && allObjectsQuery.data) {
      setAvailableObject([
        ...allObjectsQuery.data.map<IReactSelectOptions<number>>((val) => {
          const objectType = objectTypeIntoRus(val.type)
          return {
            label: val.name + " (" + objectType + ")",
            value: val.id,
          }
        })
      ])
    }
  }, [allObjectsQuery.data, allObjectsQuery.isSuccess])

  // Select Team Logic
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({
    label: "",
    value: 0,
  })
  const [availableTeams, setAvailableTeams] = useState<IReactSelectOptions<number>[]>([])
  const allTeamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["all-teams-in-object", selectedObject.value],
    queryFn: () => getTeamsFromObjectID(selectedObject.value),
    enabled: selectedObject.value != 0,
  })
  useEffect(() => {
    if (allTeamsQuery.isSuccess && allTeamsQuery.data) {
      setAvailableTeams([
        ...allTeamsQuery.data.map<IReactSelectOptions<number>>((val) => ({
          label: val.teamNumber + " (" + val.teamLeaderName + ")",
          value: val.id
        }))
      ])
    }
  }, [allTeamsQuery.data, allTeamsQuery.isSuccess])

  // Select Material Logic
  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({
    label: "",
    value: 0,
  })
  const [availableMaterials, setAvailableMaterials] = useState<IReactSelectOptions<number>[]>([])
  const allMaterialsQuery = useQuery<InvoiceObjectTeamMaterialData[], Error, InvoiceObjectTeamMaterialData[]>({
    queryKey: [`materials-in-team`, selectedTeam.value],
    queryFn: () => getMaterialsDataFromTeam(selectedTeam.value),
    enabled: selectedTeam.value != 0,
  })
  useEffect(() => {
    if (allMaterialsQuery.isSuccess && allMaterialsQuery.data) {
      setAvailableMaterials([
        ...allMaterialsQuery.data.map<IReactSelectOptions<number>>((val) => ({
          label: val.materialName,
          value: val.materialID,
        }))
      ])
    }
  }, [allMaterialsQuery.data, allMaterialsQuery.isSuccess])

  const onMaterialSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedMaterial({ label: "", value: 0 })
      resetInvoiceMaterialObject()
      return
    }

    setSelectedMaterial(value)
    if (allMaterialsQuery.isSuccess && allMaterialsQuery.data) {
      const material = allMaterialsQuery.data.find((val) => val.materialID == value.value)!
      setInvoiceMaterial({
        ...invoiceMaterial,
        materialID: material.materialID,
        materialName: material.materialName,
        hasSerialNumbers: material.hasSerialNumber,
        unit: material.materialUnit,
        availableMaterial: material.amount,
      })
    }
  }

  const [selectedOperation, setSelectedOperation] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [availableOperations, setAvailableOperations] = useState<IReactSelectOptions<number>[]>([])
  const allOperationsQuery = useQuery<InvoiceObjectOperations[], Error, InvoiceObjectOperations[]>({
    queryKey: ["operations-based-on-team", selectedTeam.value],
    queryFn: () => getOperationsBasedOnTeamID(selectedTeam.value),
    enabled: selectedTeam.value != 0,
  })
  useEffect(() => {
    if (allOperationsQuery.isSuccess && allOperationsQuery.data) {
      setAvailableOperations(allOperationsQuery.data
        .filter(val => val.materialID == 0)
        .map<IReactSelectOptions<number>>(val => ({
          label: val.operationName,
          value: val.operationID,
        })))
    }
  }, [allOperationsQuery.data, allOperationsQuery.isSuccess])
  const onOperationSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedOperation({ label: "", value: 0 })
      return
    }

    setSelectedOperation(value)
    setInvoiceOperation({
      ...invoiceOperation,
      operationID: value.value,
      operationName: value.label,
    })
  }

  //Operation List
  const [invoiceOperations, setInvoiceOperations] = useState<IInvoiceObjectOperations[]>([])
  const [invoiceOperation, setInvoiceOperation] = useState<IInvoiceObjectOperations>({
    operationID: 0,
    operationName: "",
    amount: 0,
    notes: "",
  })

  // Material list
  const [invoiceMaterials, setInvoiceMaterials] = useState<IInvoiceObjectMaterials[]>([])
  const [invoiceMaterial, setInvoiceMaterial] = useState<IInvoiceObjectMaterials>({
    materialID: 0,
    materialName: "",
    availableMaterial: 0,
    unit: "",
    amount: 0,
    notes: "",
    hasSerialNumbers: false,
    serialNumbers: [],
  })

  const resetInvoiceMaterialObject = () => {
    setInvoiceOperations([])
    setInvoiceOperation({
      operationID: 0,
      operationName: "",
      amount: 0,
      notes: "",
    })
    setInvoiceMaterials([])
    setInvoiceMaterial({
      materialID: 0,
      materialName: "",
      availableMaterial: 0,
      unit: "",
      amount: 0,
      notes: "",
      hasSerialNumbers: false,
      serialNumbers: [],
    })
  }

  // Serial Numbers logic
  const [availableSerialNumbers, setAvaiableSerialNumbers] = useState<IReactSelectOptions<string>[]>([])
  const [selectedSerialNumber, setSelectedSerialNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [alreadySelectedSerialNumbers, setAlreadySelectedSerialNumbers] = useState<IReactSelectOptions<string>[]>([])
  const [toBeDeletedSerialNumber, setToBeDeletedSerialNumber] = useState<IReactSelectOptions<string>>({ label: "", value: "" })

  const availableSerialNumbersQuery = useQuery<string[], Error, string[]>({
    queryKey: [`serial-numbers-of-material-${selectedMaterial.value}`],
    queryFn: () => getSerialNumbersOfMaterial(selectedMaterial.value, selectedTeam.value),
    enabled: selectedMaterial.value != 0 && selectedTeam.value != 0,
  })
  useEffect(() => {

    if (availableSerialNumbersQuery.isSuccess && availableSerialNumbersQuery.data) {
      setAvaiableSerialNumbers([
        ...availableSerialNumbersQuery.data.map<IReactSelectOptions<string>>((val) => ({ value: val, label: val }))
      ])
    }

  }, [availableSerialNumbersQuery.data, availableSerialNumbersQuery.isSuccess])

  const addToSerialNumberList = () => {
    const list = [...alreadySelectedSerialNumbers, selectedSerialNumber]
    setAlreadySelectedSerialNumbers(list)

    setAvaiableSerialNumbers(
      availableSerialNumbers.filter((value) => value.value != selectedSerialNumber.value)
    )

    setSelectedSerialNumber({ label: "", value: "" })
    setInvoiceMaterial({
      ...invoiceMaterial,
      serialNumbers: list.map((val) => val.value)
    })
  }

  const deleteSerialNumberFromList = () => {
    const list = alreadySelectedSerialNumbers.filter((value) => value.value != toBeDeletedSerialNumber.value)
    setAlreadySelectedSerialNumbers(list)

    setAvaiableSerialNumbers([
      toBeDeletedSerialNumber,
      ...availableSerialNumbers,
    ])

    setToBeDeletedSerialNumber({ label: "", value: "" })
    setInvoiceMaterial({
      ...invoiceMaterial,
      serialNumbers: list.map((val) => val.value)
    })
  }

  // Adding materials to the list
  const addMaterialToTheList = () => {

    if (selectedMaterial.value == 0) {
      toast.error("Не был выбран материал")
      return
    }

    if (invoiceMaterials.findIndex((val) => val.materialID == invoiceMaterial.materialID) != -1) {
      toast.error("Такой материал уже в списке")
      return
    }

    if (invoiceMaterial.amount <= 0) {
      toast.error("Неправильно указано количество материала")
      return
    }

    if (invoiceMaterial.amount > invoiceMaterial.availableMaterial) {
      toast.error("Количество материала превышает доступное количество")
      return
    }

    if (invoiceMaterial.hasSerialNumbers && alreadySelectedSerialNumbers.length != invoiceMaterial.amount) {
      toast.error("Количество серийных номеров не совпадает с указанным количеством материла")
      return
    }

    if (allOperationsQuery.data) {
      const operation = allOperationsQuery.data.find(val => val.materialID == invoiceMaterial.materialID)
      if (operation) {
        setInvoiceOperations([
          ...invoiceOperations,
          {
            operationID: operation.operationID,
            operationName: operation.operationName,
            amount: invoiceMaterial.amount,
            notes: "",
          }
        ])
      }
    }

    setInvoiceMaterials([invoiceMaterial, ...invoiceMaterials])
    setInvoiceMaterial({
      materialID: 0,
      materialName: "",
      availableMaterial: 0,
      unit: "",
      amount: 0,
      notes: "",
      hasSerialNumbers: false,
      serialNumbers: [],
    })
    setSelectedMaterial({ label: "", value: 0 })
    setSelectedSerialNumber({ label: "", value: "" })
    setAvaiableSerialNumbers([])
    setToBeDeletedSerialNumber({ label: "", value: "" })
    setAlreadySelectedSerialNumbers([])


  }

  const addOperationToTheList = () => {
    if (invoiceOperation.operationID == 0) {
      toast.error("Не выбрана услуга")
      return
    }

    if (invoiceOperations.findIndex((val) => val.operationID == invoiceOperation.operationID) != -1) {
      toast.error("Такая услуга уже выбрана")
      return
    }

    if (invoiceOperation.amount <= 0) {
      toast.error("Неправильно указано количество услуги")
      return
    }

    setInvoiceOperations([...invoiceOperations, invoiceOperation])
    setInvoiceOperation({
      operationID: 0,
      operationName: "",
      amount: 0,
      notes: "",
    })
    setSelectedOperation({label: "", value: 0})
  }

  // Delete from the list
  const deleteFromMaterialsList = (index: number) => {
    const materialID = invoiceMaterials[index].materialID
    setInvoiceMaterials(invoiceMaterials.filter((_, i) => i != index))
    if (allOperationsQuery.data) {
      const operation = allOperationsQuery.data.find(val => val.materialID == materialID)
      if (operation){
        setInvoiceOperations(invoiceOperations.filter((val) => val.operationID != operation.operationID))
      }
    }
  }

  const deleteFromOperationsList = (index: number) => {
    const operation = invoiceOperations[index] 
    if (allOperationsQuery.data) {
      const operationWithMaterial = allOperationsQuery.data.find(val => val.operationID == operation.operationID)
      if (operationWithMaterial) {
        if (operationWithMaterial.materialID != 0) {
          setInvoiceMaterials(invoiceMaterials.filter(val => val.materialID != operationWithMaterial.materialID))
        }
      }
    }
    setInvoiceOperations(invoiceOperations.filter((_, i) => i != index))
  }

  const createInvoiceObjectMutation = useMutation({
    mutationFn: createInvoiceObject,
  })

  const queryClient = useQueryClient()
  const navigate = useNavigate()
  const submitInvoice = () => {
    if (selectedDistrict.value == 0) {
      toast.error("Не выбран район")
      return
    }

    if (!invoiceDate || Number.isNaN(invoiceDate.getTime())) {
      toast.error("Не выбрана дата")
      return
    }

    if (selectedObject.value == 0) {
      toast.error("Не выбрана объект")
      return
    }

    if (selectedTeam.value == 0) {
      toast.error("Не выбрана бригада")
      return
    }

    if (invoiceMaterials.length == 0) {
      toast.error("Не указаны материалы для поступления")
      return
    }

    createInvoiceObjectMutation.mutate({
      details: {
        objectID: selectedObject.value,
        teamID: selectedTeam.value,
        districtid: selectedDistrict.value,
        dateOfInvoice: invoiceDate,
        id: 0,
        deliveryCode: "",
        projectID: 0,
        supervisorWorkerID: 0,
      },
      items: [
        ...invoiceMaterials.map<InvoiceObjectCreateItems>((val) => ({
          materialID: val.materialID,
          amount: val.amount,
          serialNumbers: val.serialNumbers,
          notes: val.notes,
        }))
      ],
      operations: [
        ...invoiceOperations.map<InvoiceObjectOperationsCreate>(val => ({
          operationID: val.operationID,
          amount: val.amount,
          notes: val.notes,
        }))
      ]
    }, {
      onSuccess: () => {
        queryClient.invalidateQueries(["invoice-objects-paginated"])
        navigate(INVOICE_OBJECT_USER)
      }
    })
  }

  const [activeTab, setActiveTab] = useState<"operations" | "materials">("materials")
  const onTabClick = (tabName: "operations" | "materials") => {
    setActiveTab(tabName)
  }

  return (
    <main>
      <div className="px-2 py-1">
        <span className="font-bold text-xl">Поступление материалов на объект</span>
      </div>
      <div className="px-2 py-1">
        <div className="w-full flex">
          <div
            onClick={() => submitInvoice()}
            className="text-white py-2.5 px-5 rounded-lg bg-gray-800 hover:bg-gray-700 hover:cursor-pointer"
          >
            {createInvoiceObjectMutation.isLoading ? <LoadingDots height={30} /> : "Опубликовать"}
          </div>
        </div>
        <span className="font-semibold text-lg">Основная информация</span>
        <div className="px-3 py-4 bg-gray-800 text-white rounded-md ">
          {allDistrictsQuery.isLoading &&
            <div className="flex items-center h-[40px]">
              <LoadingDots height={40} />
            </div>
          }
          {allDistrictsQuery.isSuccess &&
            <div className="flex flex-col space-y-1 pb-2">
              <span className="font-semibold">Район</span>
              <Select
                className="basic-single text-black"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name={"district-select"}
                placeholder={""}
                value={selectedDistrict}
                options={availableDistricts}
                onChange={(value) => {
                  setSelectedDistrict({
                    label: value?.label ?? "",
                    value: value?.value ?? 0,
                  })
                }}
              />
            </div>
          }
          <div className="flex flex-col space-y-1 pb-2">
            <span className="font-semibold">Дата накладной</span>
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px] max-w-[220px]">
              <DatePicker
                name="dateOfInvoice"
                className="outline-none w-full text-black"
                dateFormat={"dd-MM-yyyy"}
                selected={invoiceDate}
                onChange={(date) => setInvoiceDate(date ?? new Date(+0))}
              />
            </div>
          </div>
          {allObjectsQuery.isLoading &&
            <div className="flex items-center">
              <LoadingDots height={40} />
            </div>
          }
          {allObjectsQuery.isSuccess &&
            <div className="flex flex-col space-y-1">
              <span className="font-semibold">Объект</span>
              <Select
                className="basic-single text-black"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name={"material-cost-material-select"}
                placeholder={""}
                value={selectedObject}
                options={availableObjects}
                onChange={(value) => {
                  setSelectedObject({
                    label: value?.label ?? "",
                    value: value?.value ?? 0,
                  })
                  setSelectedTeam({ value: 0, label: "" })
                  setSelectedMaterial({ value: 0, label: "" })
                  resetInvoiceMaterialObject()
                }}
              />
            </div>
          }

          {allTeamsQuery.isLoading && allTeamsQuery.fetchStatus == "fetching" &&
            <div className="flex items-center h-[40px]">
              <LoadingDots height={40} />
            </div>
          }
          {(allTeamsQuery.isSuccess || allTeamsQuery.fetchStatus == "idle") &&
            <div className="flex flex-col space-y-1">
              <span className="font-semibold">Бригада</span>
              <Select
                className="basic-single text-black"
                classNamePrefix="select"
                isSearchable={true}
                isClearable={true}
                name={"material-cost-material-select"}
                placeholder={""}
                value={selectedTeam}
                options={availableTeams}
                onChange={(value) => {
                  setSelectedTeam({
                    label: value?.label ?? "",
                    value: value?.value ?? 0,
                  })
                  setSelectedMaterial({ value: 0, label: "" })
                  resetInvoiceMaterialObject()
                }}
              />
            </div>
          }
        </div>
      </div>
      <div className="px-2 py-1">
        <span className="font-semibold text-lg">Расход</span>
        <div className="bg-gray-800 text-white rounded-md px-2 py-2">
          <div
            className="flex w-full bg-gray-700 rounded-md"
            id="tabs"
          >
            <div className={`py-2 w-full text-center rounded-md cursor-pointer ${activeTab == "materials" ? "bg-gray-500" : ""}`} onClick={() => onTabClick("materials")}>Материалы</div>
            <div className={`py-2 w-full text-center rounded-md cursor-pointer ${activeTab == "operations" ? "bg-gray-500" : ""}`} onClick={() => onTabClick("operations")}>Услуги</div>
          </div>
        </div>
      </div>

      {activeTab == "materials" &&
        <>
          <div className="px-2 py-1">
            <span className="font-semibold text-lg">Материалы</span>
            <div className="grid grid-cols-1 gap-2">
              <div className="flex flex-col space-y-3 px-3 py-2 bg-gray-800 text-white rounded-md">
                {allMaterialsQuery.isLoading && allMaterialsQuery.fetchStatus == "fetching" &&
                  <div className="flex items-center">
                    <LoadingDots height={40} />
                  </div>
                }
                {(allMaterialsQuery.isSuccess || allMaterialsQuery.fetchStatus == "idle") &&
                  <div className="flex flex-col space-y-1">
                    <span className="font-semibold">Материалы бригады</span>
                    <Select
                      className="basic-single text-black"
                      classNamePrefix="select"
                      isSearchable={true}
                      isClearable={true}
                      name={"material-cost-material-select"}
                      placeholder={""}
                      value={selectedMaterial}
                      options={availableMaterials}
                      onChange={(value) => onMaterialSelect(value)}
                    />
                  </div>
                }

                <div className="flex flex-col space-y-1">
                  <span className="font-semibold">Количество</span>
                  <div className="flex space-x-2 items-center">
                    <input
                      type="number"
                      value={invoiceMaterial.amount}
                      onChange={(e) => setInvoiceMaterial({
                        ...invoiceMaterial,
                        amount: e.target.valueAsNumber
                      })}
                      className="text-black rounded-sm px-2 py-1.5"
                    />
                    <span>Доступно: {invoiceMaterial.availableMaterial} {invoiceMaterial.unit}</span>
                  </div>
                </div>
                {invoiceMaterial.hasSerialNumbers &&
                  <div className="flex flex-col space-y-4">
                    <div className="flex space-x-2">
                      <div className="flex flex-col space-y-1">
                        <label>Доступные серийные номера</label>
                        <Select
                          className="basic-single text-black"
                          classNamePrefix="select"
                          isSearchable={true}
                          isClearable={true}
                          name={"material-cost-material-select"}
                          placeholder={""}
                          value={selectedSerialNumber}
                          options={availableSerialNumbers}
                          onChange={(value) => setSelectedSerialNumber({
                            label: value?.label ?? "",
                            value: value?.value ?? "",
                          })}
                        />
                      </div>
                      <Button onClick={() => addToSerialNumberList()} text="Выбрать" />
                    </div>
                    <div className="flex space-x-2">
                      <div className="flex flex-col space-y-1">
                        <label>Выбранные серийные номера</label>
                        <Select
                          className="basic-single text-black"
                          classNamePrefix="select"
                          isSearchable={true}
                          isClearable={true}
                          name={"material-cost-material-select"}
                          placeholder={""}
                          value={toBeDeletedSerialNumber}
                          options={alreadySelectedSerialNumbers}
                          onChange={(value) => setToBeDeletedSerialNumber({
                            label: value?.label ?? "",
                            value: value?.value ?? "",
                          })}
                        />
                      </div>
                      <Button onClick={() => deleteSerialNumberFromList()} text="Удалить" buttonType="delete" />
                    </div>
                  </div>
                }
                <div className="flex flex-col space-y-1">
                  <span className="font-semibold">Примичание</span>
                  <textarea
                    className="w-full text-black"
                    value={invoiceMaterial.notes}
                    onChange={(e) => setInvoiceMaterial({
                      ...invoiceMaterial,
                      notes: e.target.value,
                    })}
                  />
                </div>
                <div className="col-span-2 flex space-x-2">
                  <Button
                    text="Добавить в список"
                    onClick={() => addMaterialToTheList()}
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="px-2 py-1">
            <span className="font-semibold text-lg">Список материалов на объект</span>
            <div className="grid grid-cols-1 gap-2">
              {invoiceMaterials.length == 0 &&
                <div className="px-3 py-2 bg-gray-800 text-white italic rounded-md">
                  Список материалов пока пустой
                </div>
              }
              {invoiceMaterials.map((value, index) => (
                <div key={index} className="grid grid-cols-2 gap-3 px-3 py-2 bg-gray-800 text-white rounded-md overflow-auto">
                  <div className="font-bold">Материал:</div>
                  <div>{value.materialName}</div>
                  <div className="font-bold">Количество:</div>
                  <div>{value.amount} {value.unit}</div>
                  {value.notes == ""
                    ?
                    <Fragment>
                      <div className="font-bold">Примичание:</div>
                      <div className="font-light italic">Отсутсвует</div>
                    </Fragment>
                    :
                    <Fragment>
                      <div className="font-bold col-span-2">Примичание:</div>
                      <div className="col-span-2">{value.notes}</div>
                    </Fragment>
                  }
                  <div className="col-span-2 flex space-x-2">
                    <Button
                      text="Удалить из списка"
                      buttonType="delete"
                      onClick={() => deleteFromMaterialsList(index)}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      }
      {activeTab == "operations" &&
        <>
          <div className="px-2 py-1">
            <span className="font-semibold text-lg">Услуги</span>
            <div className="grid grid-cols-1 gap-2">
              <div className="flex flex-col space-y-3 px-3 py-2 bg-gray-800 text-white rounded-md">
                {allOperationsQuery.isLoading && allOperationsQuery.fetchStatus == "fetching" &&
                  <div className="flex items-center">
                    <LoadingDots height={40} />
                  </div>
                }
                {(allOperationsQuery.isSuccess || allOperationsQuery.fetchStatus == "idle") &&
                  <div className="flex flex-col space-y-1">
                    <span className="font-semibold">Доступные услуги</span>
                    <Select
                      className="basic-single text-black"
                      classNamePrefix="select"
                      isSearchable={true}
                      isClearable={true}
                      name={"available-operations-select"}
                      placeholder={""}
                      value={selectedOperation}
                      options={availableOperations}
                      onChange={(value) => onOperationSelect(value)}
                    />
                    <span className="italic text-sm">
                      Примичание: в данном списке доступны услуги которые не привязаны к материалам. Остальные услуги будут добавлятся в список услуг на объект автоматически при выборе материалов
                    </span>
                  </div>
                }
                <div className="flex flex-col space-y-1">
                  <span className="font-semibold">Количество</span>
                  <div className="flex space-x-2 items-center">
                    <input
                      type="number"
                      value={invoiceOperation.amount}
                      onChange={(e) => setInvoiceOperation({
                        ...invoiceOperation,
                        amount: e.target.valueAsNumber
                      })}
                      className="text-black rounded-sm px-2 py-1.5"
                    />
                  </div>
                </div>
                <div className="flex flex-col space-y-1">
                  <span className="font-semibold">Примичание</span>
                  <textarea
                    className="w-full text-black"
                    value={invoiceOperation.notes}
                    onChange={(e) => setInvoiceOperation({
                      ...invoiceOperation,
                      notes: e.target.value,
                    })}
                  />
                </div>
                <div className="col-span-2 flex space-x-2">
                  <Button
                    text="Добавить в список"
                    onClick={() => addOperationToTheList()}
                  />
                </div>
              </div>
            </div>
          </div>
          <div className="px-2 py-1">
            <span className="font-semibold text-lg">Список услуг на объект</span>
            <div className="grid grid-cols-1 gap-2">
              {invoiceOperations.length == 0 &&
                <div className="px-3 py-2 bg-gray-800 text-white italic rounded-md">
                  Список услуг пока пустой
                </div>
              }
              {invoiceOperations.map((value, index) => (
                <div key={index} className="grid grid-cols-2 gap-3 px-3 py-2 bg-gray-800 text-white rounded-md overflow-auto">
                  <div className="font-bold">Услуга:</div>
                  <div>{value.operationName}</div>
                  <div className="font-bold">Количество:</div>
                  <div>{value.amount}</div>
                  {value.notes == ""
                    ?
                    <Fragment>
                      <div className="font-bold">Примичание:</div>
                      <div className="font-light italic">Отсутсвует</div>
                    </Fragment>
                    :
                    <Fragment>
                      <div className="font-bold col-span-2">Примичание:</div>
                      <div className="col-span-2">{value.notes}</div>
                    </Fragment>
                  }
                  <div className="col-span-2 flex space-x-2">
                    <Button
                      text="Удалить из списка"
                      buttonType="delete"
                      onClick={() => deleteFromOperationsList(index)}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      }
    </main>
  )
}
