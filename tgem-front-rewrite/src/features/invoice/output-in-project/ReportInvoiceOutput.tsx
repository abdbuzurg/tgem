import Modal from "@shared/components/Modal";
import Select from 'react-select'
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Button from "@shared/ui/Button";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { useEffect, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { InvoiceOutputReportFilter, buildReport, getAllUniqueCode, getAllUniqueRecieved, getAllUniqueDistrict, getAllUniqueTeam, getAllUniqueWarehouseManager } from "./api";
import LoadingDots from "@shared/ui/LoadingDots";
import toast from "react-hot-toast";

interface Props {
  setShowReportModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportInvoiceOutput({ setShowReportModal }: Props) {

  //Logic for all the districts
  const [selectedDistrict, setSelectedDistrict] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [districts, setDistricts] = useState<IReactSelectOptions<number>[]>([])
  const districtQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: ["invoice-output-districts"],
    queryFn: getAllUniqueDistrict
  })
  useEffect(() => {
    if (districtQuery.isSuccess && districtQuery.data) {
      setDistricts(districtQuery.data)
    }
  }, [districtQuery.data, districtQuery.isSuccess])

  //Logic for All the codes(serial codes)
  const [selectedCode, setSelectedCode] = useState<IReactSelectOptions<string>>({ label: "", value: "" })
  const [codes, setCodes] = useState<IReactSelectOptions<string>[]>([])
  const codesQuery = useQuery<IReactSelectOptions<string>[], Error, IReactSelectOptions<string>[]>({
    queryKey: ["invoice-output-codes"],
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
    queryKey: ["invoice-output-warehouse-managers"],
    queryFn: getAllUniqueWarehouseManager
  })
  useEffect(() => {
    if (warehouseManagersQuery.isSuccess && warehouseManagersQuery.data) {
      setWarehouseManagers(warehouseManagersQuery.data)
    }
  }, [warehouseManagersQuery.data, warehouseManagersQuery.isSuccess])

  //Logic for all Released
  const [selectedRecieved, setSelectedRecieved] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [recieveds, setReceiveds] = useState<IReactSelectOptions<number>[]>([])
  const recievedsQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: ["invoice-output-recieveds"],
    queryFn: getAllUniqueRecieved
  })
  useEffect(() => {
    if (recievedsQuery.isSuccess && recievedsQuery.data) {
      setReceiveds(recievedsQuery.data)
    }
  }, [recievedsQuery.data, recievedsQuery.isSuccess])

  //Logic for all Teams
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [teams, setTeams] = useState<IReactSelectOptions<number>[]>([])
  const teamQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: ["invoice-output-teams"],
    queryFn: getAllUniqueTeam
  })
  useEffect(() => {
    if (teamQuery.isSuccess && teamQuery.data) {
      setTeams(teamQuery.data)
    }
  }, [teamQuery.data, teamQuery.isSuccess])


  //Filter data 
  const [filter, setFilter] = useState<InvoiceOutputReportFilter>({
    code: "",
    dateFrom: null,
    dateTo: null,
    recievedID: 0,
    districtID: 0,
    teamID: 0,
    warehouseManagerID: 0,
  })

  //Submit filter
  const buildInvoiceOutputReport = useMutation({
    mutationFn: () => buildReport(filter)
  })

  const onCreateReportClick = () => {

    if (filter.dateFrom && filter.dateTo && filter.dateFrom > filter.dateTo) {
      toast.error("Неправильно указанный диапазон для дат. ")
      return
    }

    buildInvoiceOutputReport.mutate()
  }

  return (
    <Modal setShowModal={setShowReportModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отчет для накладной отпуск</p>
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
            onChange={(value) => {
              setSelectedCode(value ?? { label: "", value: "" })
              setFilter({ ...filter, code: value?.value ?? "" })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label htmlFor="district">Район</label>
          <Select
            id="district"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"district"}
            placeholder={""}
            value={selectedDistrict}
            options={districts}
            onChange={value => {
              setSelectedDistrict(value ?? { label: "", value: 0 })
              setFilter({ ...filter, districtID: value?.value ?? 0 })
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
            onChange={value => {
              setSelectedWarehouseManager(value ?? { label: "", value: 0 })
              setFilter({ ...filter, warehouseManagerID: value?.value ?? 0 })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label htmlFor="recieved">Получатель</label>
          <Select
            id="recieved"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"recieved"}
            placeholder={""}
            value={selectedRecieved}
            options={recieveds}
            onChange={value => {
              setSelectedRecieved(value ?? { label: "", value: 0 })
              setFilter({ ...filter, recievedID: value?.value ?? 0 })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label htmlFor="team">Бригады</label>
          <Select
            id="team"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"team"}
            placeholder={""}
            value={selectedTeam}
            options={teams}
            onChange={value => {
              setSelectedTeam(value ?? {label: "", value: 0})
              setFilter({ ...filter, teamID: value?.value ?? 0 })
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
                onChange={(date: Date | null) => setFilter({ ...filter, dateFrom: date ?? new Date() })}
              />
            </div>
            <div className="py-[4px] px-[8px] border-[#cccccc] border rounded-[4px]">
              <DatePicker
                name="dateOfInvoice"
                className="outline-none w-full"
                dateFormat={"dd-MM-yyyy"}
                selected={filter.dateTo}
                onChange={(date: Date | null) => setFilter({ ...filter, dateTo: date ?? new Date() })}
              />
            </div>
            <div>
              <Button onClick={() => setFilter({ ...filter, dateFrom: null, dateTo: null })} text="X" buttonType="default" />
            </div>
          </div>
        </div>
        <div className="flex">
          <div
            onClick={() => onCreateReportClick()}
            className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            {buildInvoiceOutputReport.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
