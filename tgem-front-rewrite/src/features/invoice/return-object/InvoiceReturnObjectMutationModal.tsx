import Modal from "@shared/components/Modal";
import Select from 'react-select'
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { Fragment, useEffect, useState } from "react";
import { IInvoiceReturn, IInvoiceReturnMaterials, IInvoiceReturnView } from "@entities/invoice-return/types";
import Button from "@shared/ui/Button";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  InvoiceReturnItem,
  InvoiceReturnMaterialsForSelect,
  InvoiceReturnMutation,
  createInvoiceReturn,
  getInvoiceReturnMaterialsForEdit,
  getUniqueMaterialsInLocation,
  updateInvoiceReturn,
} from "@entities/invoice-return/api";
import Input from "@shared/ui/Input";
import toast from "react-hot-toast";
import SerialNumberSelectReturnModal from "@entities/invoice-return/ui/SerialNumberSelectReturn";
import IconButton from "@shared/components/IconButtons";
import { FaBarcode } from "react-icons/fa";
import { IoIosAddCircleOutline } from "react-icons/io";
import { getAllTeamsForSelect } from "@entities/team/api";
import { getAllWorkers } from "@entities/worker/api";
import IWorker from "@entities/worker/types";
import { TeamDataForSelect } from "@entities/team/types";
import { getAllDistricts } from "@entities/district/api";
import { IDistrict } from "@entities/district/types";
import LoadingDots from "@shared/ui/LoadingDots";
import { IObject } from "@entities/object/types";
import { getAllObjects } from "@entities/object/api";
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses";

type Props =
  | {
      mode: "create"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
    }
  | {
      mode: "edit"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
      invoiceReturnObject: IInvoiceReturnView
    }

export default function InvoiceReturnObjectMutationModal(props: Props) {
  const isEdit = props.mode === "edit"
  const queryClient = useQueryClient()

  const [invoiceData, setInvoiceData] = useState<IInvoiceReturn>(
    isEdit
      ? {
          dateOfInvoice: new Date(props.invoiceReturnObject.dateOfInvoice),
          deliveryCode: props.invoiceReturnObject.deliveryCode,
          districtID: 0,
          id: props.invoiceReturnObject.id,
          notes: "",
          projectID: 0,
          returnerID: 0,
          returnerType: "object",
          acceptorID: 0,
          acceptorType: "team",
          acceptedByWorkerID: 0,
          confirmation: false,
        }
      : {
          dateOfInvoice: new Date(),
          deliveryCode: "",
          districtID: 0,
          id: 0,
          notes: "",
          projectID: 0,
          returnerID: 0,
          returnerType: "object",
          acceptorID: 0,
          acceptorType: "team",
          acceptedByWorkerID: 0,
          confirmation: false,
        }
  )

  // DISTRICT
  const [selectedDistrict, setSelectedDistrict] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allDistricts, setAllDistricts] = useState<IReactSelectOptions<number>[]>([])
  const allDistrictsQuery = useQuery<IDistrict[], Error, IDistrict[]>({
    queryKey: [`all-districts`],
    queryFn: getAllDistricts,
  })
  useEffect(() => {
    if (allDistrictsQuery.isSuccess && allDistrictsQuery.data) {
      setAllDistricts(allDistrictsQuery.data.map<IReactSelectOptions<number>>(val => ({
        label: val.name,
        value: val.id,
      })))

      if (isEdit) {
        const district = allDistrictsQuery.data.find(val => val.name == props.invoiceReturnObject.districtName)!
        setSelectedDistrict({
          label: district.name,
          value: district.id,
        })
        setInvoiceData(prev => ({
          ...prev,
          districtID: district.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allDistrictsQuery.data, allDistrictsQuery.isSuccess, isEdit])

  // OBJECTS
  const [selectedObject, setSelectedObject] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allObjects, setAllObjects] = useState<IReactSelectOptions<number>[]>([])
  const allObjectsQuery = useQuery<IObject[], Error, IObject[]>({
    queryKey: ["all-objects"],
    queryFn: getAllObjects,
  })
  useEffect(() => {
    if (allObjectsQuery.isSuccess && allObjectsQuery.data) {
      setAllObjects(allObjectsQuery.data.map<IReactSelectOptions<number>>(val => ({
        label: val.name + " (" + objectTypeIntoRus(val.type) + ")",
        value: val.id,
      })))

      if (isEdit) {
        const object = allObjectsQuery.data.find(val => val.name == props.invoiceReturnObject.objectName && val.type == props.invoiceReturnObject.objectType)!
        setSelectedObject({
          label: object.name + " (" + objectTypeIntoRus(object.type) + ")",
          value: object.id,
        })
        setInvoiceData(prev => ({
          ...prev,
          returnerID: object.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allObjectsQuery.data, allObjectsQuery.isSuccess, isEdit])

  // TEAM
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allTeams, setAllTeams] = useState<IReactSelectOptions<number>[]>([])
  const allTeamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["all-teams-for-select"],
    queryFn: getAllTeamsForSelect,
  })
  useEffect(() => {
    if (allTeamsQuery.isSuccess && allTeamsQuery.data) {
      setAllTeams(allTeamsQuery.data.map<IReactSelectOptions<number>>(val => ({
        label: val.teamNumber + " (" + val.teamLeaderName + ")",
        value: val.id,
      })))

      if (isEdit) {
        const team = allTeamsQuery.data.find(val => val.teamNumber == props.invoiceReturnObject.teamNumber && val.teamLeaderName == props.invoiceReturnObject.teamLeaderName)!
        setSelectedTeam({
          label: team.teamNumber + " (" + team.teamLeaderName + ")",
          value: team.id,
        })
        setInvoiceData(prev => ({
          ...prev,
          acceptorID: team.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allTeamsQuery.data, allTeamsQuery.isSuccess, isEdit])

  // ACCEPTOR (worker)
  const [selectedAcceptor, setSelectedAcceptor] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allWorkers, setAllWorkers] = useState<IReactSelectOptions<number>[]>([])
  const allWorkersQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["all-workers"],
    queryFn: getAllWorkers,
  })
  useEffect(() => {
    if (allWorkersQuery.isSuccess && allWorkersQuery.data) {
      setAllWorkers(allWorkersQuery.data.map<IReactSelectOptions<number>>(val => ({
        label: val.name,
        value: val.id,
      })))

      if (isEdit) {
        const worker = allWorkersQuery.data.find(val => val.name == props.invoiceReturnObject.acceptorName)!
        setSelectedAcceptor({
          label: worker.name,
          value: worker.id,
        })
        setInvoiceData(prev => ({
          ...prev,
          acceptedByWorkerID: worker.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allWorkersQuery.data, allWorkersQuery.isSuccess, isEdit])

  // INVOICE MATERIALS
  const [invoiceMaterials, setInvoiceMaterials] = useState<IInvoiceReturnMaterials[]>([])
  const invoiceMaterialsForEditQuery = useQuery<IInvoiceReturnMaterials[], Error, IInvoiceReturnMaterials[]>({
    queryKey: ["invoice-return-materials-for-edit", isEdit ? props.invoiceReturnObject.id : 0, invoiceData.returnerType, invoiceData.returnerID],
    queryFn: () => getInvoiceReturnMaterialsForEdit(isEdit ? props.invoiceReturnObject.id : 0, invoiceData.returnerType, invoiceData.returnerID),
    enabled: isEdit && invoiceData.returnerID != 0,
  })
  useEffect(() => {
    if (isEdit && invoiceMaterialsForEditQuery.isSuccess && invoiceMaterialsForEditQuery.data) {
      setInvoiceMaterials(invoiceMaterialsForEditQuery.data)
    }
  }, [invoiceMaterialsForEditQuery.data, invoiceMaterialsForEditQuery.isSuccess, isEdit])

  const [invoiceMaterial, setInvoiceMaterial] = useState<IInvoiceReturnMaterials>({
    materialID: 0,
    amount: 0,
    materialName: "",
    unit: "",
    holderAmount: 0,
    hasSerialNumber: false,
    serialNumbers: [],
    isDefective: false,
    notes: "",
  })

  // MATERIAL SELECT
  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>>({ value: 0, label: "" })
  const [allAvaialableMaterials, setAllAvailableMaterails] = useState<IReactSelectOptions<number>[]>([])
  const allMaterialInALocation = useQuery<InvoiceReturnMaterialsForSelect[], Error, InvoiceReturnMaterialsForSelect[]>({
    queryKey: ["available-materials", invoiceData.returnerType, invoiceData.returnerID],
    queryFn: () => getUniqueMaterialsInLocation(invoiceData.returnerType, invoiceData.returnerID),
  })
  useEffect(() => {
    if (allMaterialInALocation.isSuccess) {
      if (allMaterialInALocation.data)
        setAllAvailableMaterails([
          ...allMaterialInALocation.data.map<IReactSelectOptions<number>>((value) => ({ label: value.materialName, value: value.materialID }))
        ])
      else
        setAllAvailableMaterails([]);
    }
  }, [allMaterialInALocation.data, allMaterialInALocation.isSuccess])

  const onAllAvailableMaterialSelect = (value: IReactSelectOptions<number> | null) => {
    if (!value) {
      setSelectedMaterial({ label: "", value: 0 })
      setInvoiceMaterial({
        ...invoiceMaterial,
        unit: "",
        holderAmount: 0,
        materialName: "",
        hasSerialNumber: false,
        serialNumbers: [],
        materialID: 0,
      })
      return
    }
    if (allMaterialInALocation.data) {
      setSelectedMaterial(value)
      const materialInfo = allMaterialInALocation.data.find((material) => material.materialID == value.value)!
      setInvoiceMaterial({
        ...invoiceMaterial,
        materialID: materialInfo.materialID,
        unit: materialInfo.materialUnit,
        holderAmount: materialInfo.amount,
        materialName: materialInfo.materialName,
        hasSerialNumber: materialInfo.hasSerialNumber,
        serialNumbers: [],
      })
    }
  }

  // SERIAL NUMBERS
  const [showSerialNumberSelectModal, setShowSerialNumberSelectModal] = useState(false)
  const addSerialNumbersToInvoice = (serialNumbers: string[]) => {
    setShowSerialNumberSelectModal(false)
    setInvoiceMaterial({
      ...invoiceMaterial,
      serialNumbers: serialNumbers,
    })
  }

  // ADD MATERIAL
  const onAddClick = () => {
    if (invoiceMaterial.amount <= 0) {
      toast.error("Не указано количество материала")
      return
    }

    if (invoiceMaterial.amount > invoiceMaterial.holderAmount) {
      toast.error("Выбранное количество привышает доступное")
      return
    }

    if (invoiceMaterial.hasSerialNumber && invoiceMaterial.serialNumbers.length !== invoiceMaterial.amount) {
      toast.error("Количество материала не совпадает с количеством сирийных номеров")
      return
    }

    if (invoiceMaterial.materialID == 0) {
      toast.error("Не выбран материал")
      return
    }

    const index = invoiceMaterials.findIndex((value) => value.materialID == invoiceMaterial.materialID)
    if (index != -1) {
      if (invoiceMaterials[index].materialID == invoiceMaterial.materialID) {
        if (invoiceMaterials[index].isDefective == invoiceMaterial.isDefective) {
          toast.error("Материал с такой ценой и с такими статусом браковоности был указан")
          return
        }
        if (invoiceMaterial.isDefective != !invoiceMaterials[index].isDefective) {
          toast.error("Данный материал уже указан с такой ценой. Либо укажаите что он бракованный либо помяйте цену материла")
          return
        }
        if (invoiceMaterial.amount + invoiceMaterials[index].amount > invoiceMaterial.holderAmount) {
          toast.error("Сумма даннаго материала с выбранной ценой и (не-)браковынным вариантом превышают имееющееся количество")
          return
        }
      }
    }

    setInvoiceMaterials([invoiceMaterial, ...invoiceMaterials])
    setInvoiceMaterial({
      materialID: 0,
      amount: 0,
      holderAmount: 0,
      materialName: "",
      unit: "",
      notes: "",
      hasSerialNumber: false,
      serialNumbers: [],
      isDefective: false,
    })
    setSelectedMaterial({ label: "", value: 0 })
  }

  // DELETE MATERIAL
  const onDeleteClick = (index: number) => {
    setInvoiceMaterials(invoiceMaterials.filter((_, i) => i != index))
  }

  // SUBMIT
  const mutation = useMutation<InvoiceReturnMutation, Error, InvoiceReturnMutation>({
    mutationFn: isEdit ? updateInvoiceReturn : createInvoiceReturn,
    ...(isEdit
      ? {
          onSuccess: () => {
            queryClient.invalidateQueries(["invoice-return-object"])
            props.setShowModal(false)
          },
        }
      : {
          onSettled: () => {
            queryClient.invalidateQueries(["invoice-return-object"])
            props.setShowModal(false)
          },
        }),
  })

  const onMutationSubmit = () => {
    if (invoiceData.districtID == 0) {
      toast.error("Не выбран район")
    }

    if (invoiceData.acceptorID == 0) {
      toast.error("Не выбрана бригада")
      return
    }

    if (invoiceData.returnerID == 0) {
      toast.error("Не выбран Объект")
      return
    }

    if (invoiceData.acceptedByWorkerID == 0) {
      toast.error("Не выбран принимающий")
      return
    }

    if (invoiceMaterials.length == 0) {
      toast.error("Накладная не имеет материалов")
      return
    }

    mutation.mutate({
      details: invoiceData,
      items: invoiceMaterials.map<InvoiceReturnItem>((value) => ({
        amount: value.amount,
        materialID: value.materialID,
        isDefected: value.isDefective,
        serialNumbers: value.serialNumbers,
        notes: value.notes,
      })),
    })
  }

  return (
    <Modal setShowModal={props.setShowModal} bigModal>
      <div className="mb-2">
        <h3 className="text-2xl font-medium text-gray-800">
          {isEdit ? `Изменение накладной ${invoiceData.deliveryCode}` : "Добавление накладной возврат из объекта"}
        </h3>
      </div>
      <div className="flex flex-col w-full max-h-[80vh] space-y-2">
        <p className="text-xl font-semibold text-gray-800">Детали накладной</p>
        <div className="flex flex-col space-y-2 w-full">
          <div className="flex space-x-2">
            {allDistrictsQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allDistrictsQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="district">Район</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="district"
                    placeholder={""}
                    value={selectedDistrict}
                    options={allDistricts}
                    onChange={(value) => {
                      setSelectedDistrict(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        districtID: value?.value ?? 0,
                      })
                    }}
                  />
                </div>
              </div>
            }
            {allObjectsQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allObjectsQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="object">Объект</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="object"
                    placeholder={""}
                    value={selectedObject}
                    options={allObjects}
                    onChange={(value) => {
                      setSelectedObject(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        returnerID: value?.value ?? 0,
                      })
                    }}
                  />
                </div>
              </div>
            }
          </div>
          <div className="flex space-x-2 items-center">
            {allTeamsQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allTeamsQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="team">Бригада</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="team"
                    placeholder={""}
                    value={selectedTeam}
                    options={allTeams}
                    onChange={(value) => {
                      setSelectedTeam(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        acceptorID: value?.value ?? 0,
                      })
                    }}
                  />
                </div>
              </div>
            }
            {allWorkersQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allWorkersQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="acceptor">Принял</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="acceptor"
                    placeholder={""}
                    value={selectedAcceptor}
                    options={allWorkers}
                    onChange={(value) => {
                      setSelectedAcceptor(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        acceptedByWorkerID: value?.value ?? 0,
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
            {isEdit ? (
              <div
                onClick={() => onMutationSubmit()}
                className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
              >
                {mutation.isLoading ? <LoadingDots height={30} /> : "Опубликовать"}
              </div>
            ) : (
              <Button text="Опубликовать" onClick={() => onMutationSubmit()} />
            )}
          </div>
        </div>
        <div>
          <div className="flex space-x-2 items-center justify-between">
            <p className="text-xl font-semibold text-gray-800">Материалы наклданой</p>
          </div>
          <div className="grid grid-cols-7 text-sm font-bold shadow-md text-left mt-2 w-full border-box">
            <div className="px-4 py-3"><span>Наименование</span></div>
            <div className="px-4 py-3"><span>Ед.Изм.</span></div>
            <div className="px-4 py-3"><span>Доступно</span></div>
            <div className="px-4 py-3"><span>Количество</span></div>
            <div className="px-4 py-3"><span>Брак</span></div>
            <div className="px-4 py-3"><span>Примичание</span></div>
            <div className="px-4 py-3"></div>
          </div>
          <div className="grid grid-cols-7 text-sm text-left mt-2 w-full border-box">
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
                options={allAvaialableMaterials}
                onChange={(value) => onAllAvailableMaterialSelect(value)}
              />
            </div>
            <div className="px-4 py-3 flex items-center">{invoiceMaterial.unit}</div>
            <div className="px-4 py-3 flex items-center">{invoiceMaterial.holderAmount}</div>
            <div className="px-4 py-3">
              <Input
                name="amount"
                value={invoiceMaterial.amount}
                type="number"
                onChange={(e) => setInvoiceMaterial((prev) => ({ ...prev, amount: e.target.valueAsNumber }))}
              />
            </div>
            <div className="px-4 py-3 flex items-center">
              <input
                type="checkbox"
                checked={invoiceMaterial.isDefective}
                onChange={(e) => setInvoiceMaterial({ ...invoiceMaterial, isDefective: e.currentTarget.checked })}
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
            <div className="grid grid-cols-2 gap-2 items-center">
              {invoiceMaterial.hasSerialNumber &&
                <div>
                  <IconButton
                    icon={<FaBarcode size="25px" title={`Привязать серийные номера`} />}
                    onClick={() => setShowSerialNumberSelectModal(true)}
                  />
                </div>
              }
              <div className="text-center">
                <IconButton
                  icon={<IoIosAddCircleOutline size="25px" title={`Привязать серийные номера`} />}
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
            <div className="grid grid-cols-7 text-sm text-left mt-2 w-full border-box overflow-y-auto max-h-[50vh]">
              {invoiceMaterials.map((value, index) =>
                <Fragment key={index}>
                  <div className="px-4 py-3">{value.materialName}</div>
                  <div className="px-4 py-3">{value.unit}</div>
                  <div className="px-4 py-3">{value.holderAmount}</div>
                  <div className="px-4 py-3">{value.amount}</div>
                  <div className="px-4 py-3">{value.isDefective ? "ДА" : "НЕТ"}</div>
                  <div className="px-4 py-3">{value.notes}</div>
                  <div className="px-4 py-3 flex items-center">
                    <Button buttonType="delete" onClick={() => onDeleteClick(index)} text="Удалить" />
                  </div>
                </Fragment>
              )}
            </div>
          }
        </div>
      </div>
      {showSerialNumberSelectModal &&
        <SerialNumberSelectReturnModal
          setShowModal={setShowSerialNumberSelectModal}
          alreadySelectedSerialNumers={invoiceMaterial.serialNumbers}
          locationType={invoiceData.returnerType}
          locationID={invoiceData.returnerID}
          addSerialNumbersToInvoice={addSerialNumbersToInvoice}
          materialID={invoiceMaterial.materialID}
        />
      }
    </Modal>
  )
}
