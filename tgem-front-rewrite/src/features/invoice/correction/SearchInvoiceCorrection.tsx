import { useEffect, useState } from "react"
import Modal from "@shared/components/Modal"
import Select from "react-select"
import { useQuery } from "@tanstack/react-query";
import IReactSelectOptions from "@shared/types/ReactSelectOptions";
import { InvoiceCorrectionSearchParameters, InvoiceCorrectionSearchParametersData, getInvoiceCorrectionSearchParameters } from "./api";

interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  searchParameters: InvoiceCorrectionSearchParameters,
  setSearchParameters: React.Dispatch<React.SetStateAction<InvoiceCorrectionSearchParameters>>,
}

export default function SearchInvoiceCorrection({ setShowModal, searchParameters, setSearchParameters }: Props) {

  const searchParametersDataQuery = useQuery<InvoiceCorrectionSearchParametersData, Error, InvoiceCorrectionSearchParametersData>({
    queryKey: ["invoice-correction-search-parameters-data"],
    queryFn: () => getInvoiceCorrectionSearchParameters(),
  })
  useEffect(() => {
    if (searchParametersDataQuery.isSuccess && searchParametersDataQuery.data) {
      setAllObject(searchParametersDataQuery.data.objects)
      setAllTeams(searchParametersDataQuery.data.teams)

      if (searchParameters.objectID > 0) {
        const team = searchParametersDataQuery.data.teams.find((val) => val.value == searchParameters.teamID)!
        setSelectedTeam(team)
      }

      if (searchParameters.teamID > 0) {
        const object = searchParametersDataQuery.data.objects.find((val) => val.value == searchParameters.objectID)!
        setSelectedObject(object)
      }
    }
  }, [searchParametersDataQuery.data, searchParameters.objectID, searchParameters.teamID, searchParametersDataQuery.isSuccess])

  const [selectedTeam, setSelectedTeam] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allTeams, setAllTeams] = useState<IReactSelectOptions<number>[]>([])

  const [selectedObject, setSelectedObject] = useState<IReactSelectOptions<number>>({ label: "", value: 0 })
  const [allObject, setAllObject] = useState<IReactSelectOptions<number>[]>([])

  return (
    <Modal setShowModal={setShowModal}>
      <div>
        <h3 className="text-2xl font-medium text-gray-800">Поиск по накладной Приход</h3>
      </div>
      <div className="felx flex-col space-y-2">
        <div className="flex flex-col space-y-1">
          <label className="font-semibold">Бригада</label>
          <Select
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            name={"teams"}
            placeholder={""}
            value={selectedTeam}
            options={allTeams}
            onChange={value => {
              setSelectedTeam(value ?? { label: "", value: 0 })
              setSearchParameters({ ...searchParameters, teamID: value?.value ?? 0 })
            }}
          />
        </div>
        <div className="flex flex-col space-y-1">
          <label className="font-semibold">Объект</label>
          <Select
            className="basic-single"
            classNamePrefix="select"
            isSearchable={true}
            isClearable={true}
            name={"objects"}
            placeholder={""}
            value={selectedObject}
            options={allObject}
            onChange={value => {
              setSelectedObject(value ?? { label: "", value: 0 })
              setSearchParameters({ ...searchParameters, objectID: value?.value ?? 0 })
            }}
          />
        </div>
      </div>
    </Modal>
  )
}
