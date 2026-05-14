import { useEffect, useState } from "react";
import Modal from "@shared/components/Modal";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { useMutation, useQuery } from "@tanstack/react-query";
import { ReportBalanceFilter, buildReportBalance } from "@features/report/api";
import { getAllUniqueTeams } from "@entities/team/api";
import { getAllUniqueObjects } from "@entities/object/api";
import Select from 'react-select'
import LoadingDots from "@shared/ui/LoadingDots";
import { TeamDataForSelect } from "@entities/team/types";
import { ObjectDataForSelect } from "@entities/object/types";
import { objectTypeIntoRus } from "@shared/lib/data/objectStatuses";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportBalance({ setShowModal }: Props) {
  const [reportBalanceType, setReportBalanceType] = useState<"warehouse" | "teams" | "objects" | string>("warehouse")
  const [filter, setFilter] = useState<ReportBalanceFilter>({
    objectID: 0,
    teamID: 0,
    type: "warehouse",
  })

  const onRadioSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    setReportBalanceType(e.target.value)
    setFilter({ ...filter, type: e.target.value })
  }

  //Teams Logic
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({label: "", value: 0})
  const [teams, setTeams] = useState<IReactSelectOptions<number>[]>([])
  const teamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["balance-report-unique-teams"],
    queryFn: getAllUniqueTeams,
    enabled: reportBalanceType == "team"
  })
  useEffect(() => {
    if (teamsQuery.data && teamsQuery.isSuccess) {
      setTeams([...teamsQuery.data.map<IReactSelectOptions<number>>((value) => ({ 
        label: `${value.teamNumber} (${value.teamLeaderName})`, 
        value: value.id 
      }))])
    }
  }, [teamsQuery.data, teamsQuery.isSuccess])

  //Object Logic
  const [selectedObject, setSelectedObject] = useState<IReactSelectOptions<number>>({label: "", value: 0})
  const [objects, setObjects] = useState<IReactSelectOptions<number>[]>([])
  const objectsQuery = useQuery<ObjectDataForSelect[], Error, ObjectDataForSelect[]>({
    queryKey: ["balance-report-unique-objects"],
    queryFn: getAllUniqueObjects,
    enabled: reportBalanceType == "object"
  })
  useEffect(() => {
    if (objectsQuery.data && objectsQuery.isSuccess) {
      setObjects([...objectsQuery.data.map<IReactSelectOptions<number>>((value) => ({ 
        label: `${value.objectName} (${objectTypeIntoRus(value.objectType)})`, 
        value: value.id, 
      }))])
    }
  }, [objectsQuery.data, objectsQuery.isSuccess])

  //Submit filter
  const buildReportBalanceMutation = useMutation({
    mutationFn: () => buildReportBalance(filter)
  })

  const onCreateReportClick = () => {
    buildReportBalanceMutation.mutate()
  }

  return (
    <Modal setShowModal={setShowModal}>
      <div className="py-2">
        <p className="text-2xl font-bold">Отчет остатка</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <span className="text-xl font-semibold">Параметры</span>
        <div className="flex flex-col space-y-1">
          <p>Тип баланса</p>
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-2">
              <input id="warehouse" name="reportBalanceType" type="radio" value="warehouse" onChange={(e) => onRadioSelect(e)} defaultChecked />
              <label htmlFor="warehouse">Склад</label>
            </div>
            <div className="flex items-center space-x-2">
              <input id="teams" name="reportBalanceType" type="radio" value="team" onChange={(e) => onRadioSelect(e)} />
              <label htmlFor="teams">Бригады</label>
            </div>
            <div className="flex items-center space-x-2">
              <input id="objects" name="reportBalanceType" type="radio" value="object" onChange={(e) => onRadioSelect(e)} />
              <label htmlFor="objects">Объект</label>
            </div>
          </div>
        </div>
        {reportBalanceType == "team" &&
          <div className="flex flex-col space-y-1">
            <label htmlFor="teams">Номер Бригады</label>
            <Select
              id="teams"
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              menuPosition="fixed"
              name={"teams"}
              placeholder={""}
              value={selectedTeam}
              options={teams}
              onChange={(value) => {
                setSelectedTeam(value ?? {label: "", value: 0})
                setFilter({
                  ...filter,
                  teamID: value?.value ?? 0,
                })
              }}
            />
          </div>
        }
        {reportBalanceType == "object" &&
          <div className="flex flex-col space-y-1">
            <label htmlFor="teams">Имя объекта</label>
            <Select
              id="objects"
              className="basic-single"
              classNamePrefix="select"
              isSearchable={true}
              isClearable={true}
              menuPosition="fixed"
              name={"objects"}
              placeholder={""}
              value={selectedObject}
              options={objects}
              onChange={(value) => {
                setSelectedObject(value ?? {label: "", value: 0})
                setFilter({
                  ...filter,
                  objectID: value?.value ?? 0,
                })
              }}
            />
          </div>
        }
        <div className="flex">
          <div
            onClick={() => onCreateReportClick()}
            className="text-center text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 hover:cursor-pointer"
          >
            {buildReportBalanceMutation.isLoading ? <LoadingDots height={30} /> : "Создать Отчет"}
          </div>
        </div>
      </div>
    </Modal>
  )
}
