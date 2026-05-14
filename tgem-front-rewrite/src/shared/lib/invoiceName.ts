export default function invoiceName(invoiceFromURL: string): string {
  switch(invoiceFromURL) {
    case "input":
      return "Приход"
    
    case "output":
      return "Уход"

    case "return":
      return "Возврат"

    case "writeoff":
      return "Списание"
    
    default:
      throw new Error("Incorrect Invoice type")
  }
}