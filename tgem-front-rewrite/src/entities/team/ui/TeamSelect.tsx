import Select from 'react-select'
import IReactSelectOptions from '@shared/types/ReactSelectOptions'
import useTeamOptions from '../hooks/useTeamOptions'

interface Props {
  selectedTeamID: IReactSelectOptions<number>
  setSelectedTeamID: React.Dispatch<React.SetStateAction<IReactSelectOptions<number>>>
}

export default function TeamSelect({ selectedTeamID, setSelectedTeamID }: Props) {
  const options = useTeamOptions()

  return (
    <div className="flex flex-col space-y-1">
      <label htmlFor="teams">Бригада</label>
      <div className="w-[200px]">
        <Select
          className="basic-single"
          classNamePrefix="select"
          isSearchable={true}
          isClearable={true}
          name={"teams"}
          placeholder={""}
          value={selectedTeamID}
          options={options}
          onChange={(value) => setSelectedTeamID(value ?? { label: "", value: 0 })}
        />
      </div>
    </div>
  )
}
