import { Fragment, useEffect, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import Select from "react-select"
import DatePicker from "react-datepicker"
import "react-datepicker/dist/react-datepicker.css"
import toast from "react-hot-toast"
import { IoIosAddCircleOutline } from "react-icons/io"

import Modal from "@shared/components/Modal"
import IconButton from "@shared/components/IconButtons"
import Button from "@shared/ui/Button"
import Input from "@shared/ui/Input"
import LoadingDots from "@shared/ui/LoadingDots"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import {
  IInvoiceWriteOff,
  IInvoiceWriteOffMaterials,
  IInvoiceWriteOffView,
} from "@entities/invoice-writeoff/types"
import {
  InvoiceWriteOffItem,
  InvoiceWriteOffMaterialsForSelect,
  InvoiceWriteOffMutation,
  createInvoiceWriteOff,
  getInvoiceWriteOffMaterialsForEdit,
  getUniqueMaterialsInLocation,
  updateInvoiceWriteOff,
} from "@entities/invoice-writeoff/api"
import useTeamOptions from "@entities/team/hooks/useTeamOptions"
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses"
import { IObject } from "@entities/object/types"
import { getAllObjects } from "@entities/object/api"

export type LocationKind = "warehouse" | "team" | "object"

export type WriteOffType =
  | "loss-warehouse"
  | "writeoff-warehouse"
  | "loss-team"
  | "loss-object"
  | "writeoff-object"

type Props =
  | {
      mode: "create"
      locationKind: LocationKind
      writeOffType: WriteOffType
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
    }
  | {
      mode: "edit"
      locationKind: LocationKind
      writeOffType: WriteOffType
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
      invoiceWriteOff: IInvoiceWriteOffView
    }

const LOCATION_LABELS: Record<LocationKind, string> = {
  warehouse: "склад",
  team: "Бригада",
  object: "Объект",
}

export default function WriteoffMutationModal(props: Props) {
  const isEdit = props.mode === "edit"
  const queryClient = useQueryClient()
  const { locationKind, writeOffType } = props

  const [invoiceData, setInvoiceData] = useState<IInvoiceWriteOff>(
    isEdit
      ? {
          id: props.invoiceWriteOff.id,
          projectID: props.invoiceWriteOff.projectID,
          releasedWorkerID: props.invoiceWriteOff.releasedWorkerID,
          writeOffType,
          writeOffLocationID: props.invoiceWriteOff.writeOffLocationID,
          dateOfInvoice: new Date(props.invoiceWriteOff.dateOfInvoice),
          confirmation: false,
          dateOfConfirmation: new Date(),
          deliveryCode: props.invoiceWriteOff.deliveryCode,
        }
      : {
          id: 0,
          projectID: 0,
          releasedWorkerID: 0,
          writeOffType,
          writeOffLocationID: 0,
          dateOfInvoice: new Date(),
          confirmation: false,
          dateOfConfirmation: new Date(),
          deliveryCode: "",
        },
  )

  // LOCATION PICKER (only for team / object — warehouse is implicit)
  const [writeOffLocation, setWriteOffLocation] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })

  const teamOptions = useTeamOptions()
  const allObjectsQuery = useQuery<IObject[], Error, IObject[]>({
    queryKey: ["all-objects"],
    queryFn: () => getAllObjects(),
    enabled: locationKind === "object",
  })
  const objectOptions: IReactSelectOptions<number>[] = allObjectsQuery.data?.map((val) => ({
    label: `${val.name} (${objectTypeIntoRus(val.type)})`,
    value: val.id,
  })) ?? []

  const availableLocations: IReactSelectOptions<number>[] =
    locationKind === "team" ? teamOptions
    : locationKind === "object" ? objectOptions
    : []

  // Pre-select the current location on edit once the lookup data is available.
  useEffect(() => {
    if (!isEdit || locationKind === "warehouse") return
    const match = availableLocations.find((opt) => opt.value === props.invoiceWriteOff.writeOffLocationID)
    if (match) setWriteOffLocation(match)
  // availableLocations is recomputed each render but the deps that matter are
  // the underlying query data — listed via length & the target id.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isEdit, locationKind, availableLocations.length])

  // The effective location ID we use for both materials lookup and submission.
  // For warehouse, this is always 0 (server distinguishes by writeOffType).
  const effectiveLocationID = locationKind === "warehouse" ? 0 : writeOffLocation.value

  // MATERIALS
  const [invoiceMaterials, setInvoiceMaterials] = useState<IInvoiceWriteOffMaterials[]>([])
  const invoiceMaterialsForEditQuery = useQuery<IInvoiceWriteOffMaterials[], Error, IInvoiceWriteOffMaterials[]>({
    queryKey: ["invoice-writeoff-materials", isEdit ? props.invoiceWriteOff.id : 0],
    queryFn: () => getInvoiceWriteOffMaterialsForEdit(
      isEdit ? props.invoiceWriteOff.id : 0,
      locationKind,
      effectiveLocationID,
    ),
    enabled: isEdit && (locationKind === "warehouse" || effectiveLocationID !== 0),
  })
  useEffect(() => {
    if (isEdit && invoiceMaterialsForEditQuery.isSuccess && invoiceMaterialsForEditQuery.data) {
      setInvoiceMaterials(invoiceMaterialsForEditQuery.data)
    }
  }, [invoiceMaterialsForEditQuery.data, invoiceMaterialsForEditQuery.isSuccess, isEdit])

  const [invoiceMaterial, setInvoiceMaterial] = useState<IInvoiceWriteOffMaterials>({
    materialID: 0,
    materialName: "",
    unit: "",
    amount: 0,
    notes: "",
    hasSerialNumber: false,
    serialNumbers: [],
    locationAmount: 0,
  })

  const materialQuery = useQuery<InvoiceWriteOffMaterialsForSelect[], Error, InvoiceWriteOffMaterialsForSelect[]>({
    queryKey: ["material-location", locationKind, effectiveLocationID],
    queryFn: () => getUniqueMaterialsInLocation(locationKind, effectiveLocationID),
    enabled: locationKind === "warehouse" || effectiveLocationID !== 0,
  })

  const allMaterialData: IReactSelectOptions<number>[] = materialQuery.data?.map((value) => ({
    value: value.materialID,
    label: value.materialName,
  })) ?? []

  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ value: 0, label: "" })

  const onMaterialSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedMaterial({ label: "", value: 0 })
      setInvoiceMaterial({ ...invoiceMaterial, unit: "", materialID: 0, materialName: "", hasSerialNumber: false })
      return
    }
    setSelectedMaterial(value)
    if (materialQuery.data) {
      const material = materialQuery.data.find((m) => m.materialID === value.value)!
      setInvoiceMaterial({
        ...invoiceMaterial,
        unit: material.materialUnit,
        materialID: material.materialID,
        materialName: material.materialName,
        hasSerialNumber: material.hasSerialNumber,
        locationAmount: material.amount,
      })
    }
  }

  const onAddClick = () => {
    if (invoiceMaterial.materialID === 0) {
      toast.error("Не выбран материал")
      return
    }
    if (invoiceMaterials.some((v) => v.materialID === invoiceMaterial.materialID)) {
      toast.error("Такой материал уже был выбран. Выберите другой материл")
      return
    }
    if (invoiceMaterial.amount <= 0) {
      toast.error("Неправильно указано количество")
      return
    }
    if (invoiceMaterial.amount > invoiceMaterial.locationAmount) {
      toast.error("Выбранное количество привышает доступное")
      return
    }
    if (invoiceMaterial.hasSerialNumber && invoiceMaterial.serialNumbers.length !== invoiceMaterial.amount) {
      toast.error("Количство материала не совпадает с количеством серийных намеров")
      return
    }
    setInvoiceMaterials([invoiceMaterial, ...invoiceMaterials])
    setInvoiceMaterial({
      amount: 0,
      materialName: "",
      materialID: 0,
      notes: "",
      unit: "",
      hasSerialNumber: false,
      serialNumbers: [],
      locationAmount: 0,
    })
    setSelectedMaterial({ label: "", value: 0 })
  }

  const onDeleteClick = (index: number) => {
    setInvoiceMaterials(invoiceMaterials.filter((_, i) => i !== index))
  }

  // SUBMIT
  const mutation = useMutation<InvoiceWriteOffMutation, Error, InvoiceWriteOffMutation>({
    mutationFn: isEdit ? updateInvoiceWriteOff : createInvoiceWriteOff,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-writeoff", writeOffType])
      props.setShowModal(false)
    },
  })

  const onMutationSubmit = () => {
    if (locationKind !== "warehouse" && writeOffLocation.value === 0) {
      toast.error(`Не выбран ${LOCATION_LABELS[locationKind].toLowerCase()}`)
      return
    }
    if (!invoiceData.dateOfInvoice) {
      toast.error("Дата не выбрана")
      return
    }
    if (invoiceMaterials.length === 0) {
      toast.error("Накладная не имеет материалов")
      return
    }
    mutation.mutate({
      details: { ...invoiceData, writeOffLocationID: effectiveLocationID },
      items: invoiceMaterials.map<InvoiceWriteOffItem>((val) => ({
        materialID: val.materialID,
        amount: val.amount,
        notes: val.notes,
      })),
    })
  }

  const showLocationPicker = locationKind !== "warehouse"
  const stockColumnHeader = locationKind === "warehouse" ? "На складе" : "Доступно"

  return (
    <Modal setShowModal={props.setShowModal} bigModal>
      <div className="mb-2">
        <h3 className="text-2xl font-medium text-gray-800">
          {isEdit ? `Изменение накладной ${invoiceData.deliveryCode}` : "Добавление накладной"}
        </h3>
      </div>
      <div className="flex flex-col w-full max-h-[85vh]">
        <div className="flex flex-col">
          <p className="text-xl font-semibold text-gray-800">Детали накладной</p>
          <div className="flex space-x-2 items-center w-full">
            {showLocationPicker && (
              <div className="flex flex-col space-y-1">
                <label htmlFor="location">{LOCATION_LABELS[locationKind]}</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="location"
                    placeholder={""}
                    value={writeOffLocation}
                    options={availableLocations}
                    onChange={(value) => setWriteOffLocation(value ?? { label: "", value: 0 })}
                  />
                </div>
              </div>
            )}
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
              onClick={onMutationSubmit}
              className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
            >
              {mutation.isLoading ? <LoadingDots height={30} /> : "Опубликовать"}
            </div>
          </div>
        </div>
        <div>
          <p className="text-xl font-semibold text-gray-800">Материалы наклданой</p>
          <div className="grid grid-cols-6 text-sm font-bold shadow-md text-left mt-2 w-full border-box">
            <div className="px-4 py-3"><span>Наименование</span></div>
            <div className="px-4 py-3"><span>Ед.Изм.</span></div>
            <div className="px-4 py-3"><span>{stockColumnHeader}</span></div>
            <div className="px-4 py-3"><span>Количество</span></div>
            <div className="px-4 py-3"><span>Примичание</span></div>
            <div className="px-4 py-3"></div>
          </div>
          <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box items-center">
            {materialQuery.isLoading && (
              <div className="px-4 py-3"><LoadingDots height={36} /></div>
            )}
            {(materialQuery.isSuccess || (showLocationPicker && effectiveLocationID === 0)) && (
              <div className="px-4 py-3">
                <Select
                  className="basic-single"
                  classNamePrefix="select"
                  isSearchable={true}
                  isClearable={true}
                  menuPosition="fixed"
                  name="materials"
                  placeholder={""}
                  value={selectedMaterial}
                  options={allMaterialData}
                  onChange={(value) => onMaterialSelect(value)}
                />
              </div>
            )}
            <div className="px-4 py-3 flex items-center">{invoiceMaterial.unit}</div>
            <div className="px-4 py-3">{invoiceMaterial.locationAmount}</div>
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
            <div className="grid grid-cols-2 gap-2 text-center justify-items-center">
              <div className="text-center">
                <IconButton
                  icon={<IoIosAddCircleOutline size="25px" title="Добавить материал" />}
                  onClick={() => onAddClick()}
                />
              </div>
            </div>
          </div>
          {invoiceMaterials.length > 0 && (
            <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box overflow-y-auto max-h-[35vh]">
              {invoiceMaterials.map((value, index) => (
                <Fragment key={index}>
                  <div className="px-4 py-3">{value.materialName}</div>
                  <div className="px-4 py-3">{value.unit}</div>
                  <div className="px-4 py-3">{value.locationAmount}</div>
                  <div className="px-4 py-3">{value.amount}</div>
                  <div className="px-4 py-3">{value.notes}</div>
                  <div className="px-4 py-3 flex items-center">
                    <Button buttonType="delete" onClick={() => onDeleteClick(index)} text="Удалить" />
                  </div>
                </Fragment>
              ))}
            </div>
          )}
        </div>
      </div>
    </Modal>
  )
}
