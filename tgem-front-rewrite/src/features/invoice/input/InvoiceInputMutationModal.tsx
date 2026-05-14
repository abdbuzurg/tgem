import Modal from "@shared/components/Modal";
import Select from 'react-select'
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import AddNewMaterialModal from "./AddNewMaterialModal";
import Button from "@shared/ui/Button";
import { Fragment, useEffect, useState } from "react";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import Input from "@shared/ui/Input";
import { IInvoiceInput, IInvoiceInputMaterials, IInvoiceInputView } from "./types";
import getAllMaterials from "@entities/material/api/getAll";
import Material from "@entities/material/types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { IMaterialCost } from "@entities/material-cost/types";
import getMaterailCostByMaterialID from "@entities/material-cost/api/getByMaterailID";
import { InvoiceInputMaterial, InvoiceInputMutation, createInvoiceInput, getInvoiceInputMaterialsForEdit, updateInvoiceInput } from "./api";
import SerialNumberAddModal from "./SerialNumerAddModal";
import toast from "react-hot-toast";
import IconButton from "@shared/components/IconButtons";
import { FaBarcode } from "react-icons/fa";
import { IoIosAddCircleOutline } from "react-icons/io";
import LoadingDots from "@shared/ui/LoadingDots";
import { getWorkerByJobTitle } from "@entities/worker/api";
import IWorker from "@entities/worker/types";

type Props =
  | {
      mode: "create"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
    }
  | {
      mode: "edit"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
      invoiceInput: IInvoiceInputView
    }

export default function InvoiceInputMutationModal(props: Props) {
  const isEdit = props.mode === "edit"

  const [invoiceData, setInvoiceData] = useState<IInvoiceInput>(
    isEdit
      ? {
          projectID: 0,
          dateOfInvoice: new Date(props.invoiceInput.dateOfInvoice.toString().substring(0, 10)),
          deliveryCode: props.invoiceInput.deliveryCode,
          id: props.invoiceInput.id,
          notes: props.invoiceInput.notes,
          releasedWorkerID: 0,
          warehouseManagerWorkerID: 0,
          confirmation: false,
        }
      : {
          projectID: 1,
          dateOfInvoice: new Date(),
          deliveryCode: "",
          id: 0,
          notes: "",
          releasedWorkerID: 0,
          warehouseManagerWorkerID: 0,
          confirmation: false,
        }
  )

  // SELECT WAREHOUSE MANAGER LOGIC
  const [selectedWarehouseManager, setSelectedWarehouseManager] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allWarehouseManagers, setAllWarehouseManagers] = useState<IReactSelectOptions<number>[]>([])
  const warehouseManagerQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: [`worker-warehouse-manager`],
    queryFn: () => getWorkerByJobTitle("Заведующий складом"),
  })
  useEffect(() => {
    if (warehouseManagerQuery.isSuccess && warehouseManagerQuery.data) {
      setAllWarehouseManagers(warehouseManagerQuery.data.map<IReactSelectOptions<number>>((val) => ({
        label: val.name,
        value: val.id,
      })))

      if (isEdit) {
        const alreadyWarehouseManager = warehouseManagerQuery.data.find((val) => val.name === props.invoiceInput.warehouseManagerName)!
        setSelectedWarehouseManager({
          label: alreadyWarehouseManager.name,
          value: alreadyWarehouseManager.id,
        })
        setInvoiceData((prev) => ({
          ...prev,
          warehouseManagerWorkerID: alreadyWarehouseManager.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [warehouseManagerQuery.data, isEdit, warehouseManagerQuery.isSuccess])

  // Invoice materials information
  const [invoiceMaterials, setInvoiceMaterials] = useState<IInvoiceInputMaterials[]>([])
  const invoiceMaterialsForEditQuery = useQuery<IInvoiceInputMaterials[], Error, IInvoiceInputMaterials[]>({
    queryKey: ["invoice-input-materials", isEdit ? props.invoiceInput.id : 0],
    queryFn: () => getInvoiceInputMaterialsForEdit(isEdit ? props.invoiceInput.id : 0),
    enabled: isEdit,
  })
  useEffect(() => {
    if (invoiceMaterialsForEditQuery.isSuccess && invoiceMaterialsForEditQuery.data) {
      setInvoiceMaterials(invoiceMaterialsForEditQuery.data)
    }
  }, [invoiceMaterialsForEditQuery.data, invoiceMaterialsForEditQuery.isSuccess])

  const [invoiceMaterial, setInvoiceMaterial] = useState<IInvoiceInputMaterials>({
    amount: 0,
    materialID: 0,
    materialName: "",
    notes: "",
    materialCostID: 0,
    materialCost: 0,
    unit: "",
    hasSerialNumber: false,
    serialNumbers: []
  })

  // LOGIC OF ADDING NEW MATERIAL
  const [showAddNewMaterialDetaisModal, setShowAddNewMaterialDetailsModal] = useState(false)
  useEffect(() => {
    materialCostQuery.refetch()
    materialQuery.refetch()
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [showAddNewMaterialDetaisModal])

  // MATERIAL SELECT LOGIC
  const materialQuery = useQuery<Material[], Error, Material[]>({
    queryKey: ["all-materials"],
    queryFn: getAllMaterials,
  })
  const [allMaterialData, setAllMaterialData] = useState<IReactSelectOptions<number>[]>([])
  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ value: 0, label: "" })
  useEffect(() => {
    if (materialQuery.isSuccess && materialQuery.data) {
      setAllMaterialData([
        ...materialQuery
          .data
          .map<IReactSelectOptions<number>>((value) => ({ value: value.id, label: value.name })),
      ])
    }
  }, [materialQuery.data, materialQuery.isSuccess])
  const onMaterialSelect = (value: IReactSelectOptions<number> | null) => {
    setSelectedMaterialCost({ label: "", value: 0 })
    if (!value) {
      setSelectedMaterial({ label: "", value: 0 })
      setInvoiceMaterial({
        ...invoiceMaterial,
        unit: "",
        materialID: 0,
        materialName: "",
        hasSerialNumber: false,
      })
      return
    }
    setSelectedMaterial(value)
    if (materialQuery.data && materialQuery.isSuccess) {
      const material = materialQuery.data.find((material) => material.id == value.value)!
      setInvoiceMaterial({
        ...invoiceMaterial,
        unit: material.unit,
        materialID: material.id,
        materialName: material.name,
        hasSerialNumber: material.hasSerialNumber,
      })
    }
  }

  // MATERIAL COST SELECT LOGIC
  const materialCostQuery = useQuery<IMaterialCost[], Error>({
    queryKey: ["material-cost", invoiceMaterial.materialID],
    queryFn: () => getMaterailCostByMaterialID(invoiceMaterial.materialID),
  })
  const [allMaterialCostData, setAllMaterialCostData] = useState<IReactSelectOptions<number>[]>([])
  const [selectedMaterialCost, setSelectedMaterialCost] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  useEffect(() => {
    if (materialCostQuery.isSuccess && materialCostQuery.data) {
      setAllMaterialCostData([...materialCostQuery.data.map<IReactSelectOptions<number>>((value) => ({ label: value.costM19.toString(), value: value.id }))])
      if (materialCostQuery.data.length == 1) {
        setSelectedMaterialCost({ label: materialCostQuery.data[0].costM19.toString(), value: materialCostQuery.data[0].id })
        setInvoiceMaterial({
          ...invoiceMaterial,
          materialCost: materialCostQuery.data[0].costM19,
          materialCostID: materialCostQuery.data[0].id,
        })
      }
    }
  }, [materialCostQuery.data, invoiceMaterial, materialCostQuery.isSuccess])
  const onMaterialCostSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedMaterialCost({ label: "", value: 0 })
      setInvoiceMaterial({ ...invoiceMaterial, materialCostID: 0, materialCost: 0 })
      return
    }

    setSelectedMaterialCost(value)
    if (materialCostQuery.isSuccess && materialCostQuery.data) {
      const materialCost = materialCostQuery.data!.find((cost) => cost.id == value.value)!
      setInvoiceMaterial({ ...invoiceMaterial, materialCostID: materialCost.id, materialCost: materialCost.costM19 })
    }
  }
  //Serial number add modal logic
  const [showSerialNumberAddModal, setShowSerialNumberAddModal] = useState(false)
  const addSerialNumbersToInvoice = (serialNumbers: string[]) => {
    setShowSerialNumberAddModal(false)
    setInvoiceMaterial({
      ...invoiceMaterial,
      serialNumbers: serialNumbers,
    })
  }

  //ADDING MATERIAL TO LIST LOGIC
  const onAddClick = () => {

    if (invoiceMaterial.materialID == 0) {
      toast.error("Не выбран материал")
      return
    }

    const index = invoiceMaterials.findIndex((value) => value.materialID == invoiceMaterial.materialID)
    if (index != -1) {
      if (invoiceMaterial.materialCost == invoiceMaterials[index].materialCost) {
        toast.error("Такой материал с такой ценой уже был выбран. Выберите другой ценник или же другой материл")
        return
      }
    }

    if (invoiceMaterial.materialCostID == 0) {
      toast.error("Не выбрана цена материала")
      return
    }

    if (invoiceMaterial.amount <= 0) {
      toast.error("Неправильно указано количество")
      return
    }

    if (invoiceMaterial.hasSerialNumber && invoiceMaterial.serialNumbers.length !== invoiceMaterial.amount) {
      toast.error("Количство материала не совпадает с количеством серийных намеров")
      return
    }

    setInvoiceMaterials([invoiceMaterial, ...invoiceMaterials])
    setInvoiceMaterial({
      amount: 0,
      materialCostID: 0,
      materialName: "",
      materialCost: 0,
      materialID: 0,
      notes: "",
      unit: "",
      hasSerialNumber: false,
      serialNumbers: [],
    })
    setSelectedMaterial({ label: "", value: 0 })
    setSelectedMaterialCost({ label: "", value: 0 })
  }

  // DELETE MATERIAL LOGIC
  const onDeleteClick = (index: number) => {
    setInvoiceMaterials(invoiceMaterials.filter((_, i) => i != index))
  }

  // SUBMIT
  const queryClient = useQueryClient()
  const mutation = useMutation<InvoiceInputMutation, Error, InvoiceInputMutation>({
    mutationFn: isEdit ? updateInvoiceInput : createInvoiceInput,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-input"])
      props.setShowModal(false)
    }
  })

  const onMutationSubmit = () => {

    if (invoiceData.warehouseManagerWorkerID == 0) {
      toast.error("Заведующий складом не выбран")
      return
    }

    if (!invoiceData.dateOfInvoice) {
      toast.error("Дата не выбрана")
      return
    }

    if (invoiceMaterials.length == 0) {
      toast.error("Накладная не имеет материалов")
      return
    }

    mutation.mutate({
      details: invoiceData,
      items: [
        ...invoiceMaterials.map<InvoiceInputMaterial>((value) => ({
          materialData: {
            id: 0,
            amount: value.amount,
            invoiceID: 0,
            invoiceType: "input",
            materialCostID: value.materialCostID,
            notes: value.notes,
          },
          serialNumbers: value.serialNumbers,
        }))
      ],
    })

  }

  return (
    <Modal setShowModal={props.setShowModal} bigModal>
      <div className="mb-2">
        <h3 className="text-2xl font-medium text-gray-800">
          {isEdit ? `Изменение накладной ${invoiceData.deliveryCode}` : "Добавление накладной"}
        </h3>
      </div>
      <div className="flex flex-col w-full max-h-[85vh] ">
        <div className="flex flex-col">
          <p className="text-xl font-semibold text-gray-800">Детали накладной</p>
          <div className="flex space-x-2 items-center w-full">
            {warehouseManagerQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {warehouseManagerQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="warehouse-manager">Зав. Склад</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="warehouse-manager"
                    placeholder={""}
                    value={selectedWarehouseManager}
                    options={allWarehouseManagers}
                    onChange={(value) => {
                      setSelectedWarehouseManager(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        warehouseManagerWorkerID: value?.value ?? 0,
                      })
                    }}
                  />
                </div>
              </div>
            }
            <div className="flex flex-col space-y-1">
              <label htmlFor="dateOfInvoice">Дата накладной</label>
              <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
                <DatePicker
                  name="dateOfInvoice"
                  className="outline-none w-full"
                  dateFormat={"dd-MM-yyyy"}
                  selected={invoiceData.dateOfInvoice}
                  onChange={(date) => setInvoiceData({ ...invoiceData, dateOfInvoice: date ?? new Date(+0) })}
                />
              </div>
            </div>
          </div>
          <div className="mt-4 flex">
            <div
              onClick={() => onMutationSubmit()}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
            >
              {mutation.isLoading ? <LoadingDots height={30} /> : isEdit ? "Опубликовать Изменение" : "Опубликовать"}
            </div>
          </div>
        </div>
        <div>
          <div className="flex space-x-2 items-center justify-between">
            <p className="text-xl font-semibold text-gray-800">Материалы наклданой</p>
            <div>
              <Button text="Добавить новые данные" onClick={() => setShowAddNewMaterialDetailsModal(true)} />
            </div>
          </div>
          <div className="grid grid-cols-6 text-sm font-bold shadow-md text-left mt-2 w-full border-box">
            {/* table head START */}
            <div className="px-4 py-3">
              <span>Наименование</span>
            </div>
            <div className="px-4 py-3">
              <span>Ед.Изм.</span>
            </div>
            <div className="px-4 py-3">
              <span>Количество</span>
            </div>
            <div className="px-4 py-3">
              <span>Цена</span>
            </div>
            <div className="px-4 py-3">
              <span>Примичание</span>
            </div>
            <div className="px-4 py-3"></div>
            {/* table head END */}
          </div>
          <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box items-center">
            {materialQuery.isLoading &&
              <div className="px-4 py-3">
                <LoadingDots height={36} />
              </div>
            }
            {materialQuery.isSuccess &&
              <div className="px-4 py-3">
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  menuPosition="fixed"
                  name={"materials"}
                  placeholder={""}
                  value={selectedMaterial}
                  options={allMaterialData}
                  onChange={(value) => onMaterialSelect(value)}
                />
              </div>
            }
            <div className="px-4 py-3 flex items-center">{invoiceMaterial.unit}</div>
            <div className="px-4 py-3">
              <Input
                name="amount"
                value={invoiceMaterial.amount}
                type="number"
                onChange={(e) => setInvoiceMaterial((prev) => ({ ...prev, amount: e.target.valueAsNumber }))}
              />
            </div>
            {materialCostQuery.isLoading &&
              <div className="px-4 py-3">
                <LoadingDots height={36} />
              </div>
            }
            {materialCostQuery.isSuccess &&
              <div className="px-4 py-3">
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  menuPosition="fixed"
                  name={"materials-costs"}
                  placeholder={""}
                  value={selectedMaterialCost}
                  options={allMaterialCostData}
                  onChange={(value) => onMaterialCostSelect(value)}
                />
              </div>
            }
            <div className="px-4 py-3">
              <Input
                name="notes"
                value={invoiceMaterial.notes}
                type="text"
                onChange={(e) => setInvoiceMaterial((prev) => ({ ...prev, notes: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-2 text-center justify-items-center">
              {invoiceMaterial.hasSerialNumber &&
                <div>
                  <IconButton
                    icon={<FaBarcode
                      size="25px"
                      title={`Привязать серийные номера`} />}
                    onClick={() => setShowSerialNumberAddModal(true)}
                  />
                </div>
              }
              <div className="text-center">
                <IconButton
                  icon={<IoIosAddCircleOutline
                    size="25px"
                    title={`Привязать серийные номера`} />}
                  onClick={() => onAddClick()}
                />
              </div>
            </div>
          </div>
          {isEdit && invoiceMaterialsForEditQuery.isLoading &&
            <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box overflow-y-auto max-h-[35vh]">
              <div className="px-4 py-3 col-span-6">
                <LoadingDots height={30} />
              </div>
            </div>
          }
          {invoiceMaterials.length > 0 &&
            <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box overflow-y-auto max-h-[35vh]">
              {invoiceMaterials.map((value, index) =>
                <Fragment key={index}>
                  <div className="px-4 py-3">{value.materialName}</div>
                  <div className="px-4 py-3">{value.unit}</div>
                  <div className="px-4 py-3">{value.amount}</div>
                  <div className="px-4 py-3">{value.materialCost}</div>
                  <div className="px-4 py-3">{value.notes}</div>
                  <div className="px-4 py-3 flex items-center">
                    <Button buttonType="delete" onClick={() => onDeleteClick(index)} text="Удалить" />
                  </div>
                </Fragment>
              )}
            </div>
          }
        </div>
        {showAddNewMaterialDetaisModal && <AddNewMaterialModal setShowModal={setShowAddNewMaterialDetailsModal} />}
        {showSerialNumberAddModal &&
          <SerialNumberAddModal
            setShowModal={setShowSerialNumberAddModal}
            availableSerialNumber={invoiceMaterial.serialNumbers}
            addSerialNumbersToInvoice={addSerialNumbersToInvoice}
          />}
      </div>
    </Modal>
  )
}
