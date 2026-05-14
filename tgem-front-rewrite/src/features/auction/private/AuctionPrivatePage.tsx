import { HiArrowSmDown, HiArrowSmUp } from "react-icons/hi";
import Button from "@shared/ui/Button";
import { useEffect, useState } from "react";
import toast, { Toaster } from "react-hot-toast";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AuctionPrivateData, AuctionPrivateParticipantDataForSave, getDataForPrivateAuction, savePartisipantChanges } from "@features/auction/api";
import LoadingDots from "@shared/ui/LoadingDots";
import useLogout from "@app/hooks/useLogout";
import React from "react";

interface AuctionPrivate extends AuctionPrivateData {
  totalPackagePrice: number
  expand: boolean
}

export default function AuctionPrivate() {

  const queryClient = useQueryClient()
  const logout = useLogout()

  const [auctionData, setAuctionData] = useState<AuctionPrivate[]>([])
  const auctionPrivateDataQuery = useQuery<AuctionPrivateData[], Error>({
    queryKey: ["auction-private-data"],
    queryFn: () => getDataForPrivateAuction(1),
  })
  useEffect(() => {
    if (auctionPrivateDataQuery.isSuccess && auctionPrivateDataQuery.data) {
      setAuctionData([...auctionPrivateDataQuery.data.map((val) => {
        let total = 0
        val.packageItems.map((val) => {
          total += val.quantity * val.userUnitPrice
        })
        return ({
          ...val,
          expand: false,
          totalPackagePrice: total,
        })
      }
      )])
    }
  }, [auctionPrivateDataQuery.data, auctionPrivateDataQuery.isSuccess])

  const convertNumberToReadableFormat = (value: number) => {
    const valueStr = value.toString()
    const reverseValueStr = valueStr.split("").reverse().join("")
    let reverseValueStrlength = reverseValueStr.length
    let readableReverseStr = ""
    let count = 0
    while (reverseValueStrlength > 0) {
      readableReverseStr += reverseValueStr.substring(count * 3, (count + 1) * 3) + " "
      count++
      reverseValueStrlength -= 3
    }

    return readableReverseStr.split("").reverse().join("")
  }

  useEffect(() => {
    auctionData.map(val => {
      let total = 0
      val.packageItems.map(v => {
        total += v.userUnitPrice * v.quantity
      })
      return ({
        ...val,
        totalPackagePrice: total,
      })
    })

  }, [auctionData])

  const saveParticipantEntriesMutation = useMutation<boolean, Error, AuctionPrivateParticipantDataForSave[]>({
    mutationFn: savePartisipantChanges,
  })
  const onSaveChangesClick = (packageIndex: number) => {
    const filterData = auctionData[packageIndex].packageItems.filter(v => v.userUnitPrice != 0)
    if (filterData.length != auctionData[packageIndex].packageItems.length) {
      toast.error("Please set price for all the items in the package")
      return
    }
    const participantInfoToSave: AuctionPrivateParticipantDataForSave[] = auctionData[packageIndex].packageItems.map<AuctionPrivateParticipantDataForSave>(v => ({
      itemID: v.itemID,
      unitPrice: v.userUnitPrice,
      comment: v.comment,
    }))

    saveParticipantEntriesMutation.mutate(participantInfoToSave, {
      onSuccess: () => {
        toast.success("Your changes are saved")
        queryClient.invalidateQueries(["auction-private-data"])
      },
      onError: () => {
        toast.error("Unexpected error occurred while saving changes")
      }
    })
  }

  return (
    <>
      <nav className="relative flex md:flex-row w-full justify-normal md:justify-between md:items-center bg-gray-800 px-3 py-2 shadow-lg text-gray-400">
        <p className="text-4xl font-bold">ТГЭМ</p>
        <Button onClick={() => logout()} text="LOGOUT" />
      </nav>
      <main>
        <div className="mt-2 px-2 flex justify-between">
          <span className="text-3xl font-bold">Auction</span>
        </div>
        <table className="table-auto text-sm text-left mt-2 w-full border-box">
          <thead className="shadow-md border-t-2">
            <tr>
              <th className="px-4 py-3 w-[50px] invisible">
                <HiArrowSmDown size="20px" />
              </th>
              <th className="px-4 py-3">
                <span>Package</span>
              </th>
              <th className="px-4 py-3">
                <span>Number of items</span>
              </th>
            </tr>
          </thead>
          <tbody>
            {auctionPrivateDataQuery.isLoading &&
              <tr>
                <td colSpan={3}><LoadingDots /></td>
              </tr>
            }
            {auctionPrivateDataQuery.isError &&
              <tr>
                <td colSpan={3}>Unexpected Error Occured</td>
              </tr>
            }
            {auctionData.map((auctionPackage, index) =>
              <React.Fragment key={index}>
                <tr className={`hover:bg-gray-200 w-full border-black ${!auctionPackage.expand ? 'border-b' : 'border-b-none bg-gray-200'}`}>
                  <td
                    className="px-4 py-3 cursor-pointer"
                    onClick={() => {
                      auctionData[index].expand = !auctionData[index].expand
                      setAuctionData([...auctionData])
                    }}
                  >
                    {auctionPackage.expand ? <HiArrowSmUp size="20px" /> : <HiArrowSmDown size="20px" />}
                  </td>
                  <td className="px-4 py-3">
                    <span>{auctionPackage.packageName}</span>
                  </td>
                  <td className="px-4 py-3">
                    <span>{auctionPackage.packageItems.length}</span>
                  </td>
                </tr>
                {auctionPackage.expand &&
                  <tr className={`w-full border-black ${auctionPackage.expand ? "border-b bg-gray-200" : ""}`} >
                    <td colSpan={3} className="px-4 py-3">
                      <div>
                        <span className="font-semibold">Package Items</span>
                      </div>
                      <table className="table-auto text-sm text-left mt-2 w-full border-box border-collapse">
                        <thead>
                          <tr>
                            <th className="border border-slate-600 px-4 py-3">
                              <span>#</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Name</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Description</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Unit</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Quantity</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Note</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Unit Price</span>
                            </th>
                            <th className="border border-slate-600 px-4 py-2">
                              <span>Comment</span>
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          {auctionPackage.packageItems.map((auctionPackageItems, itemIndex) =>
                            <tr key={itemIndex}>
                              <td className="border border-slate-600 px-4 py-2">
                                <span>{itemIndex + 1}</span>
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <span>{auctionPackageItems.name}</span>
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <span>{auctionPackageItems.description}</span>
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <span>{auctionPackageItems.unit}</span>
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <span>{auctionPackageItems.quantity}</span>
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <span>{auctionPackageItems.note}</span>
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <input type="number" value={auctionPackageItems.userUnitPrice} className="w-[100px]" onChange={(e) => {
                                  if (isNaN(e.target.valueAsNumber)) {
                                    return
                                  }

                                  if (e.target.valueAsNumber < 0) {
                                    return
                                  }

                                  const quantity = auctionData[index].packageItems[itemIndex].quantity
                                  const substractValue = Math.round(quantity * auctionData[index].packageItems[itemIndex].userUnitPrice * 100) / 100
                                  auctionData[index].totalPackagePrice -= substractValue
                                  auctionData[index].packageItems[itemIndex].userUnitPrice = e.target.valueAsNumber
                                  auctionData[index].totalPackagePrice += Math.round(e.target.valueAsNumber * quantity * 100) / 100
                                  setAuctionData([...auctionData])
                                }} />
                              </td>
                              <td className="border border-slate-600 px-4 py-2">
                                <textarea value={auctionPackageItems.comment} className="text-sm" onChange={(e) => {
                                  auctionData[index].packageItems[itemIndex].comment = e.target.value
                                  setAuctionData([...auctionData])
                                }}></textarea>
                              </td>
                            </tr>
                          )}
                          <tr>
                            <td className="px-4 py-2">
                            </td>
                            <td className="px-4 py-2">
                            </td>
                            <td className="px-4 py-2">
                            </td>
                            <td className="px-4 py-2">
                            </td>
                            <td className="px-4 py-2">
                            </td>
                            <td className="px-4 py-2">
                              <span className="font-semibold">Current lowest price for the package: <span className="font-bold text-lg">{auctionData[index].minimumPackagePrice == 0 ? "Not set yet" : convertNumberToReadableFormat(auctionData[index].minimumPackagePrice)}</span></span>
                            </td>
                            <td className="border border-slate-600 px-4 py-2 flex-col">
                              <div className="font-semibold">Package Price: </div>
                              <div className="font-bold text-lg">{convertNumberToReadableFormat(auctionPackage.totalPackagePrice)}</div>
                            </td>
                            <td className="border border-slate-600 px-4 py-2">
                              <div className="bg-green-400 px-2 py-4 cursor-pointer text-center rounded-lg hover:bg-green-300" onClick={() => onSaveChangesClick(index)}>Save Changes</div>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                      <div>
                      </div>
                    </td>
                  </tr>
                }
              </React.Fragment>
            )}
          </tbody >
        </table >
      </main >
      <Toaster />
    </>
  )
}
