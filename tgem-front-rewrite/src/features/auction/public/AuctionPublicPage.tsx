import { useNavigate } from "react-router-dom";
import Button from "@shared/ui/Button";
import { LOGIN } from "@routes/paths";
import { HiArrowSmDown, HiArrowSmUp } from "react-icons/hi";
import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AuctionPublicData, getDataForPublicAuction } from "@features/auction/api";
import React from "react";
import LoadingDots from "@shared/ui/LoadingDots";

interface AuctionPublic extends AuctionPublicData {
  expand: boolean
}

export default function AuctionPublic() {
  const navigate = useNavigate();

  const [auctionData, setAuctionData] = useState<AuctionPublic[]>([])
  const auctionPublicDataQuery = useQuery<AuctionPublicData[], Error>({
    queryKey: ["auction-public-data"],
    queryFn: () => getDataForPublicAuction(1),
  })
  useEffect(() => {
    if (auctionPublicDataQuery.isSuccess && auctionPublicDataQuery.data) {
      setAuctionData(auctionPublicDataQuery.data.map((val) => ({
        ...val,
        participantTotalPriceForPacakge: val.participantTotalPriceForPackage.filter((v) => v.totalPrice != 0 && v.participantTitle != ""),
        expand: false,
      })))
    }
  }, [auctionPublicDataQuery.data, auctionPublicDataQuery.isSuccess])

  return (
    <>
      <nav className="relative flex md:flex-row w-full justify-normal md:justify-between md:items-center bg-gray-800 px-3 py-2 shadow-lg text-gray-400">
        <p className="text-4xl font-bold">ТГЭМ</p>
        <Button onClick={() => navigate(LOGIN)} text="LOGIN" />
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
            {auctionPublicDataQuery.isLoading &&
              <tr>
                <td colSpan={3}><LoadingDots /></td>
              </tr>
            }
            {auctionPublicDataQuery.isError &&
              <tr>
                <td colSpan={3}>Unexpected Error Occured</td>
              </tr>
            }
            {auctionData.map((auctionPackage, index) =>
              <React.Fragment key={index}>
                <tr className={"hover:bg-gray-200 w-full" + (!auctionPackage.expand ? " border-b" : " bg-gray-200")}>
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
                  <tr className={"w-full" + (auctionPackage.expand ? " border-b bg-gray-200" : "")} >
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
                            </tr>
                          )}
                        </tbody>
                      </table>
                      {auctionPackage.participantTotalPriceForPackage.length != 0 &&
                        <div className="flex-col space-y-2 mt-2">
                          <span className="font-semibold">Participants General Prices:</span>
                          <div className="flex w-full space-x-3">
                            {auctionPackage.participantTotalPriceForPackage.map((participant, participantIndex) =>
                              <div className="flex-col space-y-1 text-white py-2.5 px-5 rounded-lg bg-gray-700 text-center" key={participantIndex}>
                                <div className="font-semibold text-sm">{participant.participantTitle}</div>
                                <div className="font-bold text-xl">{participant.totalPrice}</div>
                              </div>
                            )}
                          </div>
                        </div>
                      }
                    </td>
                  </tr>
                }
              </React.Fragment>
            )}
          </tbody>
        </table>
      </main >
    </>
  )
}
