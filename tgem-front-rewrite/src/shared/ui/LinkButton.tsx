import { Link } from "react-router-dom"

interface Props{
  to: string
  text: string
  state?: unknown
}

export default function LinkBox({to, text, state}: Props) {
  return (
    <Link 
      to={to} 
      state={state}
      className="text-white py-2.5 px-5 rounded-lg bg-gray-700 hover:bg-gray-800"
    >
      {text}
    </Link>
  )
}