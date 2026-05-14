package utils

func IsCorrectInvoiceType(invoiceType string) bool {
	return invoiceType == "input" || invoiceType == "output" || invoiceType == "return" || invoiceType == "writeoff"
}
