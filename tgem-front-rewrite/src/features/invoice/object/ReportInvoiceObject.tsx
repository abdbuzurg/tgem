import Select from 'react-select'
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import Modal from '@shared/components/Modal';
import IReactSelectOptions from '@shared/types/ReactSelectOptions';
import { useEffect, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { InvoiceCorrectionReportFilter, buildInvoiceCorrectionReport, getInvoiceCorrectionUniqueObjects, getInvoiceCorrectionUniqueTeams } from '@features/invoice/correction/api';
import toast from 'react-hot-toast';
import Button from '@shared/ui/Button';
import LoadingDots from '@shared/ui/LoadingDots';
import { ObjectDataForSelect } from '@entities/object/types';
import { objectTypeIntoRus } from '@shared/lib/data/objectStatuses';

interface Props {
  setShowReportModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportInvoiceObject({ setShowReportModal }: Props) {

  //LOGIC FOR SELECT OBJECTS
  const [selectedObject, setSelectedObject] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allObjects, setAllObjects] = useState<IReactSelectOptions<number>[]>([])
  const allObjectQuery = useQuery<ObjectDataForSelect[], Error, ObjectDataForSelect[]>({
    queryKey: ["invoice-correction-objects"],
    queryFn: getInvoiceCorrectionUniqueObjects
  })
  useEffect(() => {
    if (allObjectQuery.isSuccess && allObjectQuery.data) {
      setAllObjects(allObjectQuery.data.map(val => ({
        label: val.objectName + " (" + objectTypeIntoRus(val.objectType) + ")",
        value: val.id,
      })))
    }
  }, [allObjectQuery.data, allObjectQuery.isSuccess])

  //LOGIC FOR SELECT TEAM
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allTeams, setAllTeams] = useState<IReactSelectOptions<number>[]>([])
  const allTeamsQuery = useQuery<IReactSelectOptions<number>[], Error, IReactSelectOptions<number>[]>({
    queryKey: ["invoice-correction-teams"],
    queryFn: getInvoiceCorrectionUniqueTeams,
  })
  useEffect(() => {
    if (allTeamsQuery.isSuccess && allTeamsQuery.data) {
      setAllTeams(allTeamsQuery.data)
    }
  }, [allTeamsQuery.data, allTeamsQuery.isSuccess])

  //Filter data 
  const [filter, setFilter] = useState<InvoiceCorrectionReportFilter>({
    dateFrom: null,
    dateTo: null,
    teamID: 0,
    objectID: 0,
  })

  //Submit filter
  const buildInvoiceCorrectionReportMutation = useMutation({
    mutationFn: () => buildInvoiceCorrectionReport(filter)
  })

  const onCreateReportClick = () => {

    if (filter.dateFrom && filter.dateTo && filter.dateFrom > filter.dateTo) {
      toast.error("Неправильно указанный диапазон для дат. ")
      return
    }

    buildInvoiceCorrectionReportMutation.mutate()
  }

  return (
    <Modal setShowModal={setShowReportModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отчет Расход материалов на объект</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <span className="text-xl font-semibold">Параметры</span>
        <div className="flex flex-col space-y-1">
          <label htmlFor="object">Объект</label>
          <Select
            id="object"
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            menuPosition="fixed"
            name={"object"}
            placeholder={""}
            value={selectedObject}
            options={allObjects}
            onChange={value => {
              setSelectedObject(value ?? { label: "", value: 0 })
              setFilter({ ...filter, objectID: value?.value ?? 0 })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label htmlFor="team">Бригада</label>
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
            options={allTeams}
            onChange={value => {
              setSelectedTeam(value ?? { label: "", value: 0 })
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
            {buildInvoiceCorrectionReportMutation.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
