import Modal from "@shared/components/Modal"
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Select from 'react-select'
import Input from "@shared/ui/Input";
import IconButton from "@shared/components/IconButtons";
import { IoIosAddCircleOutline } from "react-icons/io";
import { Fragment, useEffect, useState } from "react";
import Button from "@shared/ui/Button";
import LoadingDots from "@shared/ui/LoadingDots";
import { InvoiceOutputOutOfProject } from "./types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { IInvoiceOutputMaterials } from "@features/invoice/output-in-project/types";
import { AvailableMaterial, InvoiceOutputItem, getAvailableMaterialsInWarehouse } from "@features/invoice/output-in-project/api";
import toast from "react-hot-toast";
import {
  InvoiceOutputOutOfProjectMutation,
  InvoiceOutputOutOfProjectView,
  createInvoiceOutputOfOutProject,
  getInvoiceOutputOutOfProjectMaterialsForEdit,
  updateInvoiceOutputOfOutProject,
} from "./api";

type Props =
  | {
      mode: "create"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
    }
  | {
      mode: "edit"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
      invoiceOutputOutOfProject: InvoiceOutputOutOfProjectView
    }

export default function InvoiceOutputOutOfProjectMutationModal(props: Props) {
  const isEdit = props.mode === "edit"
  const queryClient = useQueryClient()

  const [invoiceData, setInvoiceData] = useState<InvoiceOutputOutOfProject>(
    isEdit
      ? {
          id: props.invoiceOutputOutOfProject.id,
          projectID: 0,
          nameOfProject: props.invoiceOutputOutOfProject.nameOfProject,
          dateOfInvoice: new Date(props.invoiceOutputOutOfProject.dateOfInvoice),
          releasedWorkerID: 0,
          confirmation: false,
          deliveryCode: props.invoiceOutputOutOfProject.deliveryCode,
          notes: "",
        }
      : {
          id: 0,
          projectID: 0,
          nameOfProject: "",
          dateOfInvoice: new Date(),
          releasedWorkerID: 0,
          confirmation: false,
          deliveryCode: "",
          notes: "",
        }
  )

  const [invoiceMaterials, setInvoiceMaterials] = useState<IInvoiceOutputMaterials[]>([])
  const invoiceMaterialsQuery = useQuery<IInvoiceOutputMaterials[], Error, IInvoiceOutputMaterials[]>({
    queryKey: ["invoice-output-materials", isEdit ? props.invoiceOutputOutOfProject.id : 0],
    queryFn: () => getInvoiceOutputOutOfProjectMaterialsForEdit(isEdit ? props.invoiceOutputOutOfProject.id : 0),
    enabled: isEdit,
  })
  useEffect(() => {
    if (isEdit && invoiceMaterialsQuery.isSuccess && invoiceMaterialsQuery.data) {
      const realData: IInvoiceOutputMaterials[] = []
      invoiceMaterialsQuery.data.forEach((val) => {
        const index = realData.findIndex(value => value.materialName == val.materialName)
        if (index >= 0) {
          realData[index].amount += val.amount
        } else {
          realData.push(val)
        }
      })
      setInvoiceMaterials(realData)
    }
  }, [invoiceMaterialsQuery.data, invoiceMaterialsQuery.isSuccess, isEdit])

  const [invoiceMaterial, setInvoiceMaterial] = useState<IInvoiceOutputMaterials>({
    amount: 0,
    materialName: "",
    unit: "",
    warehouseAmount: 0,
    materialID: 0,
    notes: "",
    hasSerialNumber: false,
    serialNumbers: [],
  })

  // MATERIAL SELECT LOGIC
  const materialQuery = useQuery<AvailableMaterial[], Error, AvailableMaterial[]>({
    queryKey: ["available-materials"],
    queryFn: getAvailableMaterialsInWarehouse,
  })
  const [allMaterialData, setAllMaterialData] = useState<IReactSelectOptions<number>[]>([])
  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ value: 0, label: "" })
  useEffect(() => {
    if (materialQuery.isSuccess && materialQuery.data) {
      setAllMaterialData([
        ...materialQuery.data.map<IReactSelectOptions<number>>((value) => ({
          value: value.id,
          label: value.name,
        })),
      ])
    }
  }, [materialQuery.data, materialQuery.isSuccess])

  const onMaterialSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedMaterial({ label: "", value: 0 })
      setInvoiceMaterial({
        ...invoiceMaterial,
        unit: "",
        materialID: 0,
        materialName: "",
        warehouseAmount: 0,
        hasSerialNumber: false,
        serialNumbers: [],
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
        warehouseAmount: material.amount,
        hasSerialNumber: material.hasSerialNumber,
        serialNumbers: [],
      })
    }
  }

  // ADD MATERIAL LOGIC
  const onAddClick = () => {
    const materialExistIndex = invoiceMaterials.findIndex((value) =>
      value.materialID == invoiceMaterial.materialID
    )
    if (materialExistIndex !== -1) {
      toast.error("Данный материал уже в списке. Используйте другой")
      return
    }

    if (invoiceMaterial.materialID == 0) {
      toast.error("Не выбран материал")
      return
    }

    if (invoiceMaterial.amount <= 0) {
      toast.error("Неправильно указано количество материала")
      return
    }

    if (invoiceMaterial.amount > invoiceMaterial.warehouseAmount) {
      toast.error("Указаное количество привышает доступное количество на складе")
      return
    }

    if (invoiceMaterial.hasSerialNumber && invoiceMaterial.amount != invoiceMaterial.serialNumbers.length) {
      toast.error("Указанное количество материалов и количество добавленных серийных номеров не совпадают")
      return
    }

    setInvoiceMaterials([invoiceMaterial, ...invoiceMaterials])
    setSelectedMaterial({ label: "", value: 0 })
    setInvoiceMaterial({
      amount: 0,
      materialName: "",
      unit: "",
      warehouseAmount: 0,
      materialID: 0,
      notes: "",
      hasSerialNumber: false,
      serialNumbers: [],
    })
  }

  // DELETE MATERIAL LOGIC
  const onDeleteClick = (index: number) => {
    setInvoiceMaterials(invoiceMaterials.filter((_, i) => i != index))
  }

  // SUBMIT
  const mutation = useMutation<InvoiceOutputOutOfProjectMutation, Error, InvoiceOutputOutOfProjectMutation>({
    mutationFn: isEdit ? updateInvoiceOutputOfOutProject : createInvoiceOutputOfOutProject,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-output-out-of-project"])
      props.setShowModal(false)
    },
  })

  const onMutationSubmit = () => {
    if (invoiceData.nameOfProject == "") {
      toast.error("Не указан проект")
      return
    }

    mutation.mutate({
      details: invoiceData,
      items: [
        ...invoiceMaterials.map<InvoiceOutputItem>((value) => ({
          materialID: value.materialID,
          amount: value.amount,
          serialNumbers: value.serialNumbers,
          notes: value.notes,
        })),
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
      <div className="flex flex-col w-full max-h-[80vh]">
        <div className="flex flex-col space-y-2">
          <p className="text-xl font-semibold text-gray-800">Детали накладной</p>
          <div className="flex space-x-2 items-center w-full">
            <div className="flex flex-col space-y-1">
              <label htmlFor="nameOfProject">Имя проекта</label>
              <input
                type="text"
                name="nameOfProject"
                onChange={(e) => setInvoiceData({ ...invoiceData, nameOfProject: e.target.value })}
                value={invoiceData.nameOfProject}
              />
            </div>
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
              {mutation.isLoading ? <LoadingDots height={30} /> : "Опубликовать"}
            </div>
          </div>
        </div>
        <div>
          <div className="grid grid-cols-6 text-sm font-bold shadow-md text-left mt-2 w-full border-box">
            <div className="px-4 py-3"><span>Наименование</span></div>
            <div className="px-4 py-3"><span>Ед.Изм.</span></div>
            <div className="px-4 py-3"><span>На складе</span></div>
            <div className="px-4 py-3"><span>Количество</span></div>
            <div className="px-4 py-3"><span>Примичание</span></div>
            <div className="px-4 py-3"></div>
          </div>
          <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box ">
            {materialQuery.isLoading &&
              <div className="flex h-full items-center px-4 py-3">
                <LoadingDots height={40} />
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
            <div className="px-4 py-3 flex items-center">{invoiceMaterial.warehouseAmount}</div>
            <div className="px-4 py-3">
              <Input
                name="amount"
                value={invoiceMaterial.amount}
                type="number"
                onChange={(e) => setInvoiceMaterial((prev) => ({ ...prev, amount: e.target.valueAsNumber }))}
              />
            </div>
            <div className="px-4 py-3">
              <Input
                name="notes"
                value={invoiceMaterial.notes}
                type="text"
                onChange={(e) => setInvoiceMaterial((prev) => ({ ...prev, notes: e.target.value }))}
              />
            </div>
            <div className="grid grid-cols-2 gap-2 text-center items-center">
              <div className="text-center">
                <IconButton
                  icon={<IoIosAddCircleOutline size="25px" title={`Привязать серийные номера`} />}
                  onClick={() => onAddClick()}
                />
              </div>
            </div>
          </div>
          <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box overflow-y-scroll max-h-[30vh]">
            {invoiceMaterials.map((value, index) =>
              <Fragment key={index}>
                <div className="px-4 py-3">{value.materialName}</div>
                <div className="px-4 py-3">{value.unit}</div>
                <div className="px-4 py-3">{value.warehouseAmount}</div>
                <div className="px-4 py-3">{value.amount}</div>
                <div className="px-4 py-3">{value.notes}</div>
                <div className="px-4 py-3 flex items-center">
                  <Button buttonType="delete" onClick={() => onDeleteClick(index)} text="Удалить" />
                </div>
              </Fragment>
            )}
          </div>
        </div>
      </div>
    </Modal>
  )
}
