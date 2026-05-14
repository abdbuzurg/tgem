import { useEffect, useState } from "react"
import { InvoiceReturnReportFilter, buildInvoiceReturnReport, getUniqueCode, getUniqueObject, getUniqueTeam } from "../api"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import Modal from "@shared/components/Modal"
import Select from 'react-select'
import { useMutation, useQuery } from "@tanstack/react-query"
import Button from "@shared/ui/Button"
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import toast from "react-hot-toast"
import LoadingDots from "@shared/ui/LoadingDots"

interface Props {
  setShowReportModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportInvoiceReturn({ setShowReportModal }: Props) {

  // Filter Data
  const [filter, setFilter] = useState<InvoiceReturnReportFilter>({
    code: "",
    dateFrom: null,
    dateTo: null,
    returner: "",
    returnerType: "all"
  })

  //Code logic
  const [codes, setCodes] = useState<IReactSelectOptions<string>[]>([])
  const codeQuery = useQuery<string[], Error, string[]>({
    queryKey: ["invoice-return-codes"],
    queryFn: getUniqueCode,
  })
  useEffect(() => {
    if (codeQuery.isSuccess && codeQuery.data) {
      setCodes(codeQuery.data.map<IReactSelectOptions<string>>(value => ({ label: value, value: value })))
    }
  }, [codeQuery.data, codeQuery.isSuccess])

  //Team logic
  const [teams, setTeams] = useState<IReactSelectOptions<string>[]>([])
  const teamsQuery = useQuery<string[], Error, string[]>({
    queryKey: ["invoice-return-teams"],
    queryFn: getUniqueTeam,
    enabled: filter.returnerType == "teams"
  })
  useEffect(() => {
    if (teamsQuery.isSuccess && teamsQuery.data) {
      setTeams(teamsQuery.data.map<IReactSelectOptions<string>>(value => ({ label: value, value: value })))
    }
  }, [teamsQuery.data, teamsQuery.isSuccess])

  //Object logic
  const [objects, setObjects] = useState<IReactSelectOptions<string>[]>([])
  const objectsQuery = useQuery<string[], Error, string[]>({
    queryKey: ["invoice-retrun-objects"],
    queryFn: getUniqueObject,
    enabled: filter.returnerType == "objects"
  })
  useEffect(() => {
    if (objectsQuery.isSuccess && objectsQuery.data) {
      setObjects(objectsQuery.data.map<IReactSelectOptions<string>>(value => ({ label: value, value: value })))
    }
  }, [objectsQuery.data, objectsQuery.isSuccess])

  const buildInvoiceReturnReportMutation = useMutation({
    mutationFn: () => buildInvoiceReturnReport(filter)
  })

  const onCreateReportClick = () => {
    if (filter.dateTo && filter.dateFrom && filter.dateFrom > filter.dateTo) {
      toast.error("Неправильно указаны даты")
      return
    }

    buildInvoiceReturnReportMutation.mutate()
  }

  return (
    <Modal setShowModal={setShowReportModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отчет для накладной возврат</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <span className="text-xl font-semibold">Параметры</span>
        <div className="flex flex-col space-y-1">
          <label htmlFor="returnType">Тип возврата</label>
          <div id="returnType" className="flex space-x-3">
            <div className="flex space-x-1">
              <input
                type="radio"
                name="returnType"
                id="all"
                value="all"
                defaultChecked={true}
                onChange={() => setFilter({ ...filter, returnerType: "all" })}
              />
              <label htmlFor="all">Все</label>
            </div>
            <div className="flex space-x-1">
              <input
                type="radio"
                name="returnType"
                id="teams"
                value="teams"
                onChange={() => setFilter({ ...filter, returnerType: "teams" })}
              />
              <label htmlFor="teams">Бригады</label>
            </div>
            <div className="flex space-x-1">
              <input
                type="radio"
                name="returnType"
                id="objects"
                value="objects"
                onChange={() => setFilter({ ...filter, returnerType: "objects" })}
              />
              <label htmlFor="objects">Объекты</label>
            </div>
          </div>
        </div>
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
            value={{ value: filter.code, label: filter.code }}
            options={codes}
            onChange={(value: null | IReactSelectOptions<string>) => setFilter({ ...filter, code: value?.value ?? "" })}
          />
        </div>
        {filter.returnerType == "teams" && <div className="flex flex-col space-y-1">
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
            value={{ value: filter.returner, label: filter.returner }}
            options={teams}
            onChange={(value: null | IReactSelectOptions<string>) => setFilter({ ...filter, returner: value?.value ?? "" })}
          />
        </div>
        }
        {filter.returnerType == "objects" && <div className="flex flex-col space-y-1">
          <label htmlFor="object">Объект</label>
          <Select
            id="obecjt"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"object"}
            placeholder={""}
            value={{ value: filter.returner, label: filter.returner }}
            options={objects}
            onChange={(value: null | IReactSelectOptions<string>) => setFilter({ ...filter, returner: value?.value ?? "" })}
          />
        </div>
        }
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
            {buildInvoiceReturnReportMutation.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
