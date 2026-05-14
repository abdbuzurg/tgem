package utils

import "backend-v2/model"

func InvoiceMaterailsDevider(newInvoiceMaterails, oldInvoiceMaterails []model.InvoiceMaterials) ([]int, []model.InvoiceMaterials, []model.InvoiceMaterials, []model.InvoiceMaterials) {
	commonInvoiceMaterials := []model.InvoiceMaterials{}
	addInvoiceMaterails := []model.InvoiceMaterials{}
	deleteInvoiceMaterials := []model.InvoiceMaterials{}
	indexesOfOldInvoiceMaterials := []int{}

	for _, newInvoiceMaterial := range newInvoiceMaterails {
		exist := false
		for index, oldInvoiceMaterial := range oldInvoiceMaterails {
			if newInvoiceMaterial.ID == oldInvoiceMaterial.ID {
				commonInvoiceMaterials = append(commonInvoiceMaterials, newInvoiceMaterial)
				indexesOfOldInvoiceMaterials = append(indexesOfOldInvoiceMaterials, index)
				exist = true
			}
		}

		if !exist {
			addInvoiceMaterails = append(addInvoiceMaterails, newInvoiceMaterial)
		}
	}

	for _, oldInvoiceMaterial := range oldInvoiceMaterails {
		exist := false
		for _, newInvoiceMaterial := range newInvoiceMaterails {
			if oldInvoiceMaterial.ID == newInvoiceMaterial.ID {
				exist = true
				break
			}
		}

		if !exist {
			deleteInvoiceMaterials = append(deleteInvoiceMaterials, oldInvoiceMaterial)
		}
	}

	return indexesOfOldInvoiceMaterials, commonInvoiceMaterials, addInvoiceMaterails, deleteInvoiceMaterials
}
