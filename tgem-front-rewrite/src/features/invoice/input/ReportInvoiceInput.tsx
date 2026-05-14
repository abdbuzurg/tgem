import Modal from "@shared/components/Modal";
import Select from 'react-select'
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Button from "@shared/ui/Button";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { InvoiceInputReportFilter, buildReport, getAllUniqueCode, getAllUniqueReleased, getAllUniqueWarehouseManager } from "./api";
import toast from "react-hot-toast";
import LoadingDots from "@shared/ui/LoadingDots";

interface Props {
  setShowReportModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportInvoiceInput({ setShowReportModal }: Props) {

  //Logic for All the codes(serial codes)
  const [selectedCode, setSelectedCode] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [codes, setCodes] = useState<IReactSelectOptions<string>[]>([])
  const codesQuery = useQuery<IReactSelectOptions<string>[], Error, IReactSelectOptions<string>[]>({
    queryKey: ["invoice-input-codes"],
    queryFn: getAllUniqueCode
  })
  useEffect(() => {
    if (codesQuery.isSuccess && codesQuery.data) {
      setCodes(codesQuery.data)
    }
  }, [codesQuery.data, codesQuery.isSuccess])

  //Logic for all the warehouse managers
  const [selectedWarehouseManager, setSelectedWarehouseManager] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [warehouseManagers, setWarehouseManagers] = useState<IReactSelectOptions<number>[]>([])
  const warehouseManagersQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: ["invoice-input-warehouse-managers"],
    queryFn: getAllUniqueWarehouseManager
  })
  useEffect(() => {
    if (warehouseManagersQuery.isSuccess && warehouseManagersQuery.data) {
      setWarehouseManagers(warehouseManagersQuery.data)
    }
  }, [warehouseManagersQuery.data, warehouseManagersQuery.isSuccess])

  //Logic for all Released
  const [selectedReleased, setSelectedReleased] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [releaseds, setReleaseds] = useState<IReactSelectOptions<number>[]>([])
  const releasedsQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: ["invoice-input-releaseds"],
    queryFn: getAllUniqueReleased
  })
  useEffect(() => {
    if (releasedsQuery.isSuccess && releasedsQuery.data) {
      setReleaseds(releasedsQuery.data)
    }
  }, [releasedsQuery.data, releasedsQuery.isSuccess])

  //Filter data 
  const [filter, setFilter] = useState<InvoiceInputReportFilter>({
    code: "",
    dateFrom: null,
    dateTo: null,
    releasedID: 0,
    warehouseManagerID: 0,
  })

  //Submit filter

  const buildInvoiceInputReport = useMutation({
    mutationFn: () => buildReport(filter)
  })

  const onCreateReportClick = () => {

    if (filter.dateFrom && filter.dateTo) {
      toast.error("Неправильно указан диапазон дат")
      return
    }

    buildInvoiceInputReport.mutate()
  }

  return (
    <Modal setShowModal={setShowReportModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отчет для накладной приход</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <span className="text-xl font-semibold">Параметры</span>
        <div className="flex flex-col space-y-1">
          <label htmlFor="code">Номер накладной</label>
          <Select
            id="code"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"code"}
            placeholder={""}
            value={selectedCode}
            options={codes}
            onChange={(value: null | IReactSelectOptions<string>) => {
              setSelectedCode(value ?? { label: "", value: "" })
              setFilter({ ...filter, code: value?.value ?? "" })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label htmlFor="warehouseManager">Зав. Склад</label>
          <Select
            id="warehouseManager"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"warehouseManager"}
            placeholder={""}
            value={selectedWarehouseManager}
            options={warehouseManagers}
            onChange={(value) => { 
              setSelectedWarehouseManager(value ?? {label: "", value: 0})
              setFilter({ ...filter, warehouseManagerID: value?.value ?? 0 })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label htmlFor="released">Составил</label>
          <Select
            id="released"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"released"}
            placeholder={""}
            value={selectedReleased}
            options={releaseds}
            onChange={(value) => {
              setSelectedReleased(value ?? {label: "", value: 0})
              setFilter({ ...filter, releasedID: value?.value ?? 0 })
            }}
          />
        </div>
        <div className="felx flex-col space-y-1">
          <label htmlFor="rangeDate">Диапозон Дат</label>
          <div className="flex space-x-2">
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
              <DatePicker
                name="dateOfInvoice"
                className="outline-none w-full"
                dateFormat={"dd-MM-yyyy"}
                selected={filter.dateFrom}
                onChange={(date: Date | null) =>
                  setFilter({ ...filter, dateFrom: date ?? new Date() })
                }
              />
            </div>
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
              <DatePicker
                name="dateOfInvoice"
                className="outline-none w-full"
                dateFormat={"dd-MM-yyyy"}
                selected={filter.dateTo}
                onChange={(date: Date | null) =>
                  setFilter({ ...filter, dateTo: date ?? new Date() })
                }
              />
            </div>
            <div>
              <Button
                onClick={() =>
                  setFilter({
                    ...filter,
                    dateFrom: null,
                    dateTo: null
                  })
                }
                text="X"
                buttonType="default"
              />
            </div>
          </div>
        </div>
        <div className="flex">
          <div
            onClick={() => onCreateReportClick()}
            className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            {buildInvoiceInputReport.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
