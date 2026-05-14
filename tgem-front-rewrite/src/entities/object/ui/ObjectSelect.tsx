import Select from 'react-select'
import IReactSelectOptions from '@shared/types/ReactSelectOptions'
import useObjectOptions from '../hooks/useObjectOptions'

interface Props {
  selectedObjectID: IReactSelectOptions<number>
  setSelectedObjectID: React.Dispatch<React.SetStateAction<IReactSelectOptions<number>>>
}

export default function ObjectSelect({ selectedObjectID, setSelectedObjectID }: Props) {
  const options = useObjectOptions()

  return (
    <div className="flex flex-col space-y-1">
      <label htmlFor={"objects"}>Объект</label>
      <div className="w-[200px]">
        <Select
          className="basic-single"
          classNamePrefix="select"
          isSearchable={true}
          isClearable={true}
          name={"objects"}
          placeholder={""}
          value={selectedObjectID}
          options={options}
          onChange={(value) => setSelectedObjectID(value ?? { label: "", value: 0 })}
        />
      </div>
    </div>
  )
}
