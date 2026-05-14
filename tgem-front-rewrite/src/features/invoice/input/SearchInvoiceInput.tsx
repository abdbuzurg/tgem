import { useEffect, useState } from "react"
import Modal from "@shared/components/Modal"
import Select from "react-select"
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { InvoiceInputSearchParameters, InvoiceInputSearchParametersData, getInvoiceInputSearchParametersData } from "./api";
import { useQuery } from "@tanstack/react-query";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  searchParameters: InvoiceInputSearchParameters,
  setSearchParameters: React.Dispatch<React.SetStateAction<InvoiceInputSearchParameters>>,
}

export default function SearchInvoiceInput({ setShowModal, searchParameters, setSearchParameters }: Props) {

  const [searchType, setSearchType] = useState<"parametric" | "materialistic">("parametric")
  const searchParametersDataQuery = useQuery<InvoiceInputSearchParametersData, Error, InvoiceInputSearchParametersData>({
    queryKey: ["invoice-input-search-parameters-data"],
    queryFn: () => getInvoiceInputSearchParametersData(),
  })
  useEffect(() => {
    if (searchParametersDataQuery.isSuccess && searchParametersDataQuery.data) {
      setAllDeliveryCodes(searchParametersDataQuery.data.deliveryCodes.map(value => ({
        label: value,
        value: value,
      })))

      setAllWarehouseManager(searchParametersDataQuery.data.warehouseManagers)
      setAllReleased(searchParametersDataQuery.data.releaseds)
      setAllMaterials(searchParametersDataQuery.data.materials)

      if (searchParameters.deliveryCode) {
        setSelectedDelivery({ label: searchParameters.deliveryCode, value: searchParameters.deliveryCode })
      }

      if (searchParameters.warehouseManagerWorkerID >= 0) {
        const warehouseManager = searchParametersDataQuery.data.warehouseManagers.find((val) => val.value == searchParameters.warehouseManagerWorkerID)!
        setSelectedWarehouseManager(warehouseManager)
      }

      if (searchParameters.releasedWorkerID >= 0) {
        const releasedWorker = searchParametersDataQuery.data.releaseds.find((val) => val.value == searchParameters.releasedWorkerID)!
        setSelectedReleased(releasedWorker)
      }

      if (searchParameters.materials.length > 0) {
        const materials = searchParametersDataQuery.data.materials.filter((val) => searchParameters.materials.includes(val.value))
        setSelectedMaterial(materials)
      }

    }
  }, [searchParametersDataQuery.data, searchParameters.deliveryCode, searchParameters.materials, searchParameters.releasedWorkerID, searchParameters.warehouseManagerWorkerID, searchParametersDataQuery.isSuccess])

  const [selectedDelivery, setSelectedDelivery] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [allDeliveryCodes, setAllDeliveryCodes] = useState<IReactSelectOptions<string>[]>([])

  const [selectedWarehouseManager, setSelectedWarehouseManager] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allWarehouseManager, setAllWarehouseManager] = useState<IReactSelectOptions<number>[]>([])

  const [selectedReleased, setSelectedReleased] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allRealesed, setAllReleased] = useState<IReactSelectOptions<number>[]>([])

  const [selectedMaterial, setSelectedMaterial] = useState<IReactSelectOptions<number>[]>([])
  const [allMaterials, setAllMaterials] = useState<IReactSelectOptions<number>[]>([])

  return (

    <Modal setShowModal={setShowModal}>
      <div>
        <h3 className="text-2xl font-medium text-gray-800">Поиск по накладной Приход</h3>
      </div>
      <div className="flex space-x-3 py-2">
        <div className="flex space-x-2 items-center">
          <input
            id="searchByParameters"
            type="radio"
            checked={searchType == "parametric"}
            onClick={() => setSearchType("parametric")}
            onChange={() => { }}
          />
          <label htmlFor="searchByParameters" className="font-semibold">Поиск по параметрам накладной</label>
        </div>
        <div className="flex space-x-2 items-center">
          <input
            id="searchByMaterials"
            type="radio"
            checked={searchType == "materialistic"}
            onClick={() => setSearchType("materialistic")}
            onChange={() => { }}
          />
          <label htmlFor="searchByMaterials" className="font-semibold">Поиск по материалам накладной</label>
        </div>
      </div>
      {searchType == "parametric" &&
        <div className="felx flex-col space-y-2">
          <div className="flex flex-col space-y-1">
            <label className="font-semibold">Код Накладной</label>
            <Select
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              name={"delivery-codes"}
              placeholder={""}
              value={selectedDelivery}
              options={allDeliveryCodes}
              onChange={value => {
                setSelectedDelivery(value ?? { label: "", value: "" })
                setSearchParameters({ ...searchParameters, deliveryCode: value?.value ?? "" })
              }}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label className="font-semibold">Заведующий складом</label>
            <Select
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              name={"warehouse-manager"}
              placeholder={""}
              value={selectedWarehouseManager}
              options={allWarehouseManager}
              onChange={value => {
                setSelectedWarehouseManager(value ?? { label: "", value: 0 })
                setSearchParameters({ ...searchParameters, warehouseManagerWorkerID: value?.value ?? 0 })
              }}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label className="font-semibold">Составитель</label>
            <Select
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              name={"released-person"}
              placeholder={""}
              value={selectedReleased}
              options={allRealesed}
              onChange={value => {
                setSelectedReleased(value ?? { label: "", value: 0 })
                setSearchParameters({ ...searchParameters, releasedWorkerID: value?.value ?? 0 })
              }}
            />
          </div>
          <div className="flex flex-col space-y-1">
            <label className="font-semibold">Дата</label>
            <div className="flex space-x-2">
              <div className="flex space-x-1 items-center">
                <span>От: </span>
                <DatePicker
                  name="dateOfInvoice"
                  className="outline-none w-full"
                  dateFormat={"dd-MM-yyyy"}
                  selected={searchParameters.dateFrom}
                  onChange={(date: Date | null) =>
                    setSearchParameters({ ...searchParameters, dateFrom: date ?? new Date() })
                  }
                />
              </div>
              <div className="flex space-x-1 items-center">
                <span>До: </span>
                <DatePicker
                  name="dateOfInvoice"
                  className="outline-none w-full"
                  dateFormat={"dd-MM-yyyy"}
                  selected={searchParameters.dateTo}
                  onChange={(date: Date | null) =>
                    setSearchParameters({ ...searchParameters, dateTo: date ?? new Date() })
                  }
                />
              </div>
            </div>
          </div>
        </div>
      }
      {searchType == "materialistic" &&
        <div className="felx flex-col space-y-2">
          <div className="flex flex-col space-y-1">
            <label className="font-semibold">Материалы</label>
            <Select
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              name={"search-materials"}
              placeholder={""}
              value={selectedMaterial}
              options={allMaterials}
              isMulti
              onChange={value => {
                setSelectedMaterial(value.map(val => ({ value: val.value, label: val.label })))
                setSearchParameters({ ...searchParameters, materials: value.map(val => val.value) })
              }}
            />
          </div>
        </div>
      }
    </Modal>
  )
}
