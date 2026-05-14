import { Fragment, useEffect, useState } from "react";
import Modal from "@shared/components/Modal";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Select from 'react-select'
import Button from "@shared/ui/Button";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Input from '@shared/ui/Input'
import SerialNumberSelectModal from "./SerialNumberSelectModal";
import toast from "react-hot-toast";
import IconButton from "@shared/components/IconButtons";
import { FaBarcode } from "react-icons/fa";
import { IoIosAddCircleOutline } from "react-icons/io";
import { IInvoiceOutputInProject, IInvoiceOutputInProjectView, IInvoiceOutputMaterials } from "./types";
import {
  AvailableMaterial,
  InvoiceOutputInProjectMutation,
  InvoiceOutputItem,
  createInvoiceOutputInProject,
  getAvailableMaterialsInWarehouse,
  getInvoiceOutputInProjectMaterialsForEdit,
  updateInvoiceOutputInProject,
} from "./api";
import { TeamDataForSelect } from "@entities/team/types";
import { getAllTeamsForSelect } from "@entities/team/api";
import LoadingDots from "@shared/ui/LoadingDots";
import { IDistrict } from "@entities/district/types";
import { getAllDistricts } from "@entities/district/api";
import { getAllWorkers, getWorkerByJobTitle } from "@entities/worker/api";
import IWorker from "@entities/worker/types";

type Props =
  | {
      mode: "create"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
    }
  | {
      mode: "edit"
      setShowModal: React.Dispatch<React.SetStateAction<boolean>>
      invoiceOutputInProject: IInvoiceOutputInProjectView
    }

export default function InvoiceOutputInProjectMutationModal(props: Props) {
  const isEdit = props.mode === "edit"
  const queryClient = useQueryClient()

  // MAIN INVOICE INFORMATION
  const [invoiceData, setInvoiceData] = useState<IInvoiceOutputInProject>(
    isEdit
      ? {
          id: props.invoiceOutputInProject.id,
          projectID: 0,
          warehouseManagerWorkerID: 0,
          releasedWorkerID: 0,
          districtID: 0,
          teamID: 0,
          recipientWorkerID: 0,
          deliveryCode: props.invoiceOutputInProject.deliveryCode,
          notes: props.invoiceOutputInProject.notes,
          dateOfInvoice: new Date(props.invoiceOutputInProject.dateOfInvoice),
          confirmation: false,
        }
      : {
          dateOfInvoice: new Date(),
          deliveryCode: "",
          districtID: 0,
          id: 0,
          notes: "",
          projectID: 0,
          recipientWorkerID: 0,
          releasedWorkerID: 0,
          teamID: 0,
          warehouseManagerWorkerID: 0,
          confirmation: false,
        }
  )

  // DISTRICT SELECT LOGIC
  const [selectedDistrict, setSelectedDistrict] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allDistricts, setAllDistricts] = useState<IReactSelectOptions<number>[]>([])
  const allDistrictsQuery = useQuery<IDistrict[], Error, IDistrict[]>({
    queryKey: [`all-districts`],
    queryFn: getAllDistricts,
  })
  useEffect(() => {
    if (allDistrictsQuery.isSuccess && allDistrictsQuery.data) {
      setAllDistricts(allDistrictsQuery.data.map<IReactSelectOptions<number>>((value) => ({
        label: value.name,
        value: value.id,
      })))

      if (isEdit) {
        const alreadySelectedDistrict = allDistrictsQuery.data.find(val => val.name == props.invoiceOutputInProject.districtName)!
        setInvoiceData(prev => ({
          ...prev,
          districtID: alreadySelectedDistrict.id,
        }))
        setSelectedDistrict({
          label: alreadySelectedDistrict.name,
          value: alreadySelectedDistrict.id,
        })
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allDistrictsQuery.data, allDistrictsQuery.isSuccess, isEdit])

  // SELECT LOGIC FOR WAREHOUSE MANAGER
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
        const alreadyWarehouseManager = warehouseManagerQuery.data.find((val) => val.name == props.invoiceOutputInProject.warehouseManagerName)!
        setInvoiceData(prev => ({
          ...prev,
          warehouseManagerWorkerID: alreadyWarehouseManager.id,
        }))
        setSelectedWarehouseManager({
          label: alreadyWarehouseManager.name,
          value: alreadyWarehouseManager.id,
        })
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [warehouseManagerQuery.data, isEdit, warehouseManagerQuery.isSuccess])

  // SELECT LOGIC FOR RECIPIENT
  const [selectedRecipient, setSelectedRecipient] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allRecipients, setAllRecipients] = useState<IReactSelectOptions<number>[]>([])
  const allRecipientsQuery = useQuery<IWorker[], Error, IWorker[]>({
    queryKey: ["all-worker"],
    queryFn: getAllWorkers,
  })
  useEffect(() => {
    if (allRecipientsQuery.data && allRecipientsQuery.isSuccess) {
      setAllRecipients(allRecipientsQuery.data.map<IReactSelectOptions<number>>(val => ({
        label: val.name,
        value: val.id,
      })))

      if (isEdit) {
        const alreadyRecipient = allRecipientsQuery.data.find((val) => val.name == props.invoiceOutputInProject.recipientName)!
        setSelectedRecipient({
          label: alreadyRecipient.name,
          value: alreadyRecipient.id,
        })
        setInvoiceData(prev => ({
          ...prev,
          recipientWorkerID: alreadyRecipient.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allRecipientsQuery.data, allRecipientsQuery.isSuccess, isEdit])

  // SELECT LOGIC FOR TEAMS
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allTeams, setAllTeams] = useState<IReactSelectOptions<number>[]>([])
  const allTeamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["all-teams-for-select"],
    queryFn: getAllTeamsForSelect,
  })
  useEffect(() => {
    if (allTeamsQuery.data && allTeamsQuery.isSuccess) {
      setAllTeams(allTeamsQuery.data.map<IReactSelectOptions<number>>((val) => ({
        label: val.teamNumber + " (" + val.teamLeaderName + ")",
        value: val.id,
      })))

      if (isEdit) {
        const alreadyTeam = allTeamsQuery.data.find((val) => val.id == props.invoiceOutputInProject.teamID)!
        setSelectedTeam({
          label: alreadyTeam.teamNumber + " (" + alreadyTeam.teamLeaderName + ")",
          value: alreadyTeam.id,
        })
        setInvoiceData(prev => ({
          ...prev,
          teamID: alreadyTeam.id,
        }))
      }
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allTeamsQuery.data, allTeamsQuery.isSuccess, isEdit])

  //INVOICE MATERIAL DATA
  const [invoiceMaterials, setInvoiceMaterials] = useState<IInvoiceOutputMaterials[]>([])
  const invoiceMaterialsQuery = useQuery<IInvoiceOutputMaterials[], Error, IInvoiceOutputMaterials[]>({
    queryKey: ["invoice-output-materials", isEdit ? props.invoiceOutputInProject.id : 0],
    queryFn: () => getInvoiceOutputInProjectMaterialsForEdit(isEdit ? props.invoiceOutputInProject.id : 0),
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

  //MATERIAL SELECT LOGIC
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

  //Serial Number modal logic
  const [showSerialNumberSelectModal, setShowSerialNumberSelectModal] = useState(false)
  const addSerialNumbersToInvoice = (serialNumbers: string[]) => {
    setShowSerialNumberSelectModal(false)
    setInvoiceMaterial({
      ...invoiceMaterial,
      serialNumbers: serialNumbers,
    })
  }

  //ADD MATERIAL LOGIC
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

  //DELETE MATERIAL LOGIC
  const onDeleteClick = (index: number) => {
    setInvoiceMaterials(invoiceMaterials.filter((_, i) => i != index))
  }

  // SUBMIT
  const mutation = useMutation<InvoiceOutputInProjectMutation, Error, InvoiceOutputInProjectMutation>({
    mutationFn: isEdit ? updateInvoiceOutputInProject : createInvoiceOutputInProject,
    onSuccess: () => {
      queryClient.invalidateQueries(["invoice-output-in-project"])
      if (isEdit) {
        queryClient.invalidateQueries(["invoice-output-materials"])
        queryClient.invalidateQueries(["invoice-output-materials-without-serial-numbers"])
        queryClient.invalidateQueries(["invoice-output-materials-with-serial-numbers"])
      }
      props.setShowModal(false)
    },
    onError: (err) => {
      toast.error(`${isEdit ? "Ошибка при изменение накладной" : "Ошибка при создании накладной"}: ${err.message}`)
    },
  })

  const onMutationSubmit = () => {

    if (invoiceData.warehouseManagerWorkerID == 0) {
      toast.error("Не указан заведующий складом")
      return
    }

    if (invoiceData.districtID == 0) {
      toast.error("Не указан район")
      return
    }

    if (invoiceData.recipientWorkerID == 0) {
      toast.error("Не указан получатель")
      return
    }

    if (invoiceData.teamID == 0) {
      toast.error("Не указана бригада")
      return
    }

    if (invoiceMaterials.length == 0) {
      toast.error("Накладная должна иметь хотя бы 1 материал")
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
            {allDistrictsQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allDistrictsQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="districts">Район</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name={"districts"}
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

            {allRecipientsQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allRecipientsQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor="recipient">Получатель</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    name="recipient"
                    placeholder={""}
                    value={selectedRecipient}
                    options={allRecipients}
                    onChange={(value) => {
                      setSelectedRecipient(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        recipientWorkerID: value?.value ?? 0,
                      })
                    }}
                  />
                </div>
              </div>
            }
          </div>
          <div className="flex space-x-2 items-center w-full">
            {allTeamsQuery.isLoading &&
              <div className="flex h-full w-[200px] items-center">
                <LoadingDots height={40} />
              </div>
            }
            {allTeamsQuery.isSuccess &&
              <div className="flex flex-col space-y-1">
                <label htmlFor={"teams"}>Бригады</label>
                <div className="w-[200px]">
                  <Select
                    className="basic-single"
                    classNamePrefix="select"
                    isSearchable={true}
                    isClearable={true}
                    menuPosition="fixed"
                    name={"teams"}
                    placeholder={""}
                    value={selectedTeam}
                    options={allTeams}
                    onChange={(value) => {
                      setSelectedTeam(value ?? { label: "", value: 0 })
                      setInvoiceData({
                        ...invoiceData,
                        teamID: value?.value ?? 0,
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
              {mutation.isLoading ? <LoadingDots height={30} /> : isEdit ? "Опубликовать изменения" : "Опубликовать"}
            </div>
          </div>
        </div>
        <div>
          <div className="grid grid-cols-6 text-sm font-bold shadow-md text-left mt-2 w-full border-box">
            {/* table head START */}
            <div className="px-4 py-3"><span>Наименование</span></div>
            <div className="px-4 py-3"><span>Ед.Изм.</span></div>
            <div className="px-4 py-3"><span>На складе</span></div>
            <div className="px-4 py-3"><span>Количество</span></div>
            <div className="px-4 py-3"><span>Примичание</span></div>
            <div className="px-4 py-3"></div>
            {/* table head END */}
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
          <div className="grid grid-cols-6 text-sm text-left mt-2 w-full border-box overflow-y-scroll max-h-[30vh]">
            {isEdit && invoiceMaterialsQuery.isLoading &&
              <div className="col-span-6">
                <LoadingDots height={30} />
              </div>
            }
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
      {showSerialNumberSelectModal &&
        <SerialNumberSelectModal
          addSerialNumbersToInvoice={addSerialNumbersToInvoice}
          materialID={invoiceMaterial.materialID}
          setShowModal={setShowSerialNumberSelectModal}
          alreadySelectedSerialNumers={invoiceMaterial.serialNumbers}
        />
      }
    </Modal>
  )
}
