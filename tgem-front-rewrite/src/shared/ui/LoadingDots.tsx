import { ThreeDots } from "react-loader-spinner";

interface Props {
  height?: number
  width?: number
}

export default function LoadingDots({height = 80, width = 60}: Props) {
  return (
    <div className="flex items-center justify-center w-full h-full">
      <ThreeDots 
        height={height.toString()} 
        width={width.toString()} 
        radius="9"
        color="#4fa94d" 
        ariaLabel="three-dots-loading"
        wrapperStyle={{}}
        wrapperClass="text-gray-800"
        visible={true}
        />
    </div>
  )
}