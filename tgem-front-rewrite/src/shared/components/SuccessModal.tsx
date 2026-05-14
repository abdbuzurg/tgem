interface Props {
  setShowModal: React.Dispatch<React.SetStateAction<boolean>>
  onNextClick: () => void
  children: string | React.ReactNode
}

export default function SuccessModal({setShowModal, onNextClick, children}: Props) {
  return (
    <>
      <div className="fixed inset-0 z-10 overflow-y-auto">
          <div
              className="fixed inset-0 w-full h-full bg-black opacity-40"
              onClick={() => setShowModal(false)}
          ></div>
          <div className="flex items-center min-h-screen px-4 py-8">
              <div className="relative w-full max-w-lg p-4 mx-auto bg-white rounded-md shadow-lg">
                  <div className="mt-3 flex justify-start">
                      <div className="flex items-center justify-center flex-none w-12 h-12 mx-auto bg-red-100 rounded-full">
                        <svg style={{color: "green"}} viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg">
                          <path fill="#50f26b" d="M512 64a448 448 0 1 1 0 896 448 448 0 0 1 0-896zm-55.808 536.384-99.52-99.584a38.4 38.4 0 1 0-54.336 54.336l126.72 126.72a38.272 38.272 0 0 0 54.336 0l262.4-262.464a38.4 38.4 0 1 0-54.272-54.336L456.192 600.384z">
                          </path>
                        </svg>
                      </div>
                      <div className="mt-2 text-center sm:ml-4 sm:text-left">
                          <h4 className="text-lg font-medium text-gray-800">
                              Успех
                          </h4>
                          <p className="mt-2 text-[15px] leading-relaxed text-gray-500">
                              {children}
                          </p>
                          <div className="items-center gap-2 mt-3 sm:flex">
                              <button
                                  className="w-full mt-2 p-2.5 flex-1 text-gray-800 rounded-md outline-none border ring-offset-2 ring-indigo-600 focus:ring-2"
                                  onClick={() =>{
                                      setShowModal(false)
                                      onNextClick()
                                    }
                                  }
                              >
                                  Далее
                              </button>
                          </div>
                      </div>
                  </div>
              </div>
          </div>
      </div>
    </>
  )
}