package utils

func MaterialDirection(invoiceType string) string {
	switch invoiceType {
	case "input":
		return "warehouse"

	case "output":
		return "teams"

	case "return":
		return "warehouse"

	case "writeoff":
		return "warehouse"

	default:
		return "none"
	}
}
