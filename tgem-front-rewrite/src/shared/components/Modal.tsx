interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  children: React.ReactNode
  bigModal?: boolean
}

export default function Modal({setShowModal, children, bigModal = false}: Props) {
  const size = bigModal ? "" : "max-w-lg"
  return (
    <>
      <div className="fixed inset-0 z-10 overflow-y-auto">
          <div className="fixed inset-0 w-full h-full bg-black opacity-40"></div>
          <div className="flex items-center h-screen px-2">
              <div className={`relative w-full px-2 py-2 mx-auto bg-white rounded-md shadow-lg ${size} max-h-screen`}>
                  <div className="flex justify-end w-full">
                    <span 
                      className="px-3 py-0.5 rounded text-white bg-red-700 hover:bg-red-600 hover:cursor-pointer"
                      onClick={() => setShowModal(false)}
                    >
                      X
                    </span>
                  </div>
                  <div className="flex w-full">
                      <div className="text-center sm:text-left w-full">
                        {children}
                      </div>
                  </div>
              </div>
          </div>
      </div>
    </>
  )
}
