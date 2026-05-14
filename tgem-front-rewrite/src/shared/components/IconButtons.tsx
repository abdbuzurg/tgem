import { ReactElement } from "react"
import { IconType } from "react-icons"

type iconButtonType = "default" | "delete"

interface Props {
  type?: iconButtonType
  icon: ReactElement<{ size: string, title: string }, IconType>
  onClick: React.MouseEventHandler<HTMLDivElement>
}

function iconButtonColor(buttonType: iconButtonType) {
  switch(buttonType) {
    case "default":
      return "bg-gray-700 hover:bg-gray-800 border-gray-700 hover:border-gray-800"
    case "delete":
      return "border-red-700 bg-red-800 hover:bg-red-700 hover:border-red-700"
  }
}

export default function IconButton({ icon, onClick, type = "default" }: Props) {
  return (
    <div
      className={`px-4 py-2 flex items-center text-white rounded-lg text-center cursor-pointer w-fit ${iconButtonColor(type)}`}
      onClick={onClick}
    >
      {icon}
    </div>
  )
}
