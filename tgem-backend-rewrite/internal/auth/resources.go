package auth

// ResourceType is a stable, code-keyed domain concept (e.g. "invoice.output").
// String values must match resource_types.code rows seeded by migration
// 00005_permissions_v2_foundation.sql. See docs/permissions-spec.md §3.
type ResourceType string

const (
	// invoice
	ResInvoiceInput              ResourceType = "invoice.input"
	ResInvoiceOutput             ResourceType = "invoice.output"
	ResInvoiceOutputOutOfProject ResourceType = "invoice.output_out_of_project"
	ResInvoiceReturnTeam         ResourceType = "invoice.return_team"
	ResInvoiceReturnObject       ResourceType = "invoice.return_object"
	ResInvoiceWriteoff           ResourceType = "invoice.writeoff"
	ResInvoiceObject             ResourceType = "invoice.object"
	ResInvoiceCorrection         ResourceType = "invoice.correction"

	// reference
	ResRefMaterial            ResourceType = "reference.material"
	ResRefMaterialCost        ResourceType = "reference.material_cost"
	ResRefMaterialDefect      ResourceType = "reference.material_defect"
	ResRefMaterialLocation    ResourceType = "reference.material_location"
	ResRefSerialNumber        ResourceType = "reference.serial_number"
	ResRefWorker              ResourceType = "reference.worker"
	ResRefTeam                ResourceType = "reference.team"
	ResRefDistrict            ResourceType = "reference.district"
	ResRefOperation           ResourceType = "reference.operation"
	ResRefProject             ResourceType = "reference.project"
	ResRefObjectKL04KV        ResourceType = "reference.object.kl04kv"
	ResRefObjectMJD           ResourceType = "reference.object.mjd"
	ResRefObjectSIP           ResourceType = "reference.object.sip"
	ResRefObjectSTVT          ResourceType = "reference.object.stvt"
	ResRefObjectTP            ResourceType = "reference.object.tp"
	ResRefObjectSubstation    ResourceType = "reference.object.substation"
	ResRefObjectSubstationCell ResourceType = "reference.object.substation_cell"

	// report
	ResReportBalance    ResourceType = "report.balance"
	ResReportInvoice    ResourceType = "report.invoice"
	ResReportAttendance ResourceType = "report.attendance"
	ResReportStatistics ResourceType = "report.statistics"

	// admin
	ResAdminUser          ResourceType = "admin.user"
	ResAdminUserAction    ResourceType = "admin.user_action"
	ResAdminUserInProject ResourceType = "admin.user_in_project"
	ResAdminRole          ResourceType = "admin.role"
	ResAdminRoleGrant     ResourceType = "admin.role_grant"
	ResAdminResourceType  ResourceType = "admin.resource_type"
	ResAdminProject       ResourceType = "admin.project"

	// auction
	ResAuctionBidPublic  ResourceType = "auction.bid_public"
	ResAuctionBidPrivate ResourceType = "auction.bid_private"
	ResAuctionManage     ResourceType = "auction.manage"

	// hr
	ResHRAttendance ResourceType = "hr.attendance"

	// system
	ResSystemImport             ResourceType = "system.import"
	ResSystemMaterialLocLive    ResourceType = "system.material_location_live"
)
