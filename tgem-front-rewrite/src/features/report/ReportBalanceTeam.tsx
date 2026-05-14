import { useEffect, useState } from "react"
import { ReportBalanceFilter, buildReportBalance } from "@features/report/api"
import { getAllUniqueTeams } from "@entities/team/api"
import IReactSelectOptions from "@shared/types/ReactSelectOptions"
import { TeamDataForSelect } from "@entities/team/types"
import { useMutation, useQuery } from "@tanstack/react-query"
import Modal from "@shared/components/Modal"
import Select from 'react-select'
import LoadingDots from "@shared/ui/LoadingDots"

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
}

export default function ReportBalanceTeam({ setShowModal }: Props) {
  const [filter, setFilter] = useState<ReportBalanceFilter>({
    objectID: 0,
    teamID: 0,
    type: "team",
  })

  //Teams Logic
  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [teams, setTeams] = useState<IReactSelectOptions<number>[]>([])
  const teamsQuery = useQuery<TeamDataForSelect[], Error, TeamDataForSelect[]>({
    queryKey: ["balance-report-unique-teams"],
    queryFn: getAllUniqueTeams,
  })
  useEffect(() => {
    if (teamsQuery.data && teamsQuery.isSuccess) {
      setTeams([...teamsQuery.data.map<IReactSelectOptions<number>>((value) => ({
        label: `${value.teamNumber} (${value.teamLeaderName})`,
        value: value.id
      }))])
    }
  }, [teamsQuery.data, teamsQuery.isSuccess])

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
        <p className="text-2xl font-bold">Отчет остатка бригад(-ы)</p>
      </div>
      <div className="px-2 flex flex-col space-y-2 pb-2">
        <span className="text-xl font-semibold">Параметры</span>
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
              setSelectedTeam(value ?? { label: "", value: 0 })
              setFilter({
                ...filter,
                teamID: value?.value ?? 0,
              })
            }}
          />
        </div>
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
