import Select from 'react-select'
import IReactSelectOptions from '@shared/types/ReactSelectOptions'
import useDistrictOptions from '../hooks/useDistrictOptions'

interface Props {
  selectedDistrictID: IReactSelectOptions<number>
  setSelectedDistrictID: React.Dispatch<React.SetStateAction<IReactSelectOptions<number>>>
}

export default function DistrictSelect({ selectedDistrictID, setSelectedDistrictID }: Props) {
  const options = useDistrictOptions()

  return (
    <div className="flex flex-col space-y-1">
      <label htmlFor="teams">Район</label>
      <div className="w-[200px]">
        <Select
          className="basic-single"
          classNamePrefix="select"
          isSearchable={true}
          isClearable={true}
          name={"teams"}
          placeholder={""}
          value={selectedDistrictID}
          options={options}
          onChange={(value) => setSelectedDistrictID(value ?? { label: "", value: 0 })}
        />
      </div>
    </div>
  )
}
