type ButtonType = "default" | "delete"

interface Props {
  text: string
  onClick: () => void
  buttonType?: ButtonType
  disabled?: boolean
  className?: string
}

function buttonBgColor(buttonType: ButtonType) {
  switch(buttonType) {
    case "default":
      return "bg-gray-700 hover:bg-gray-800 border-gray-700 hover:border-gray-800"
    case "delete":
      return "border-red-700 bg-red-800 hover:bg-red-700 hover:border-red-700"
  }
}

export default function Button({text, onClick, buttonType = "default", disabled = false, className = ""}: Props) {
  return (
    <button 
      disabled={disabled}
      onClick={onClick}
      className={"text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800 " + className + " " + buttonBgColor(buttonType)}
    >
      {text}
    </button>
  )
}

