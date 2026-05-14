// Package http hosts the HTTP transport layer (router, route registration,
// and middleware wiring). The package literal "http" intentionally shadows
// the stdlib package within this file's scope; if you need to reference
// net/http here, alias it: import nethttp "net/http".
package http

import (
	"backend-v2/internal/auth"
	dbq "backend-v2/internal/db"
	"backend-v2/internal/http/handlers"
	"backend-v2/internal/http/middleware"
	"backend-v2/internal/usecase"
	"backend-v2/pkg/tempfiles"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, pool *pgxpool.Pool) *gin.Engine {

	queries := dbq.New(pool)
	resolver := auth.NewResolver(queries)
	userActionUsecase := usecase.NewUserActionUsecase(queries)

	mainRouter := gin.Default()
	mainRouter.MaxMultipartMemory = 400 << 20

	mainRouter.Use(gin.Recovery())

	mainRouter.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Cache-Control"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Content-Disposition"},
		AllowCredentials: true,
		AllowFiles:       true,
		MaxAge:           12 * time.Hour,
	}))

	router := mainRouter.Group("/api")
	// Tempfile cleanup: handlers that write to storage/import_excel/temp/
	// register the path via tempfiles.Track(c, path); this middleware deletes
	// every tracked path in a defer block when the request ends, so the temp
	// directory stays empty even if a handler panics or returns early.
	router.Use(tempfiles.Cleanup())
	// Audit log: wraps every /api request so that any handler running under
	// Authentication() ends up emitting a user_actions row. The middleware
	// itself skips GETs, the login/is-authenticated paths, and any request
	// whose context userID is zero (i.e. unauthenticated routes).
	router.Use(middleware.RecordUserAction(userActionUsecase))

	auctionUsecase := usecase.NewAuctionUsecase(pool)
	invoiceInputUsecase := usecase.NewInvoiceInputUsecase(pool)
	invoiceOutputUsecase := usecase.NewInvoiceOutputUsecase(pool)
	invoiceOutputOutOfProjectUsecase := usecase.NewInvoiceOutputOutOfProjectUsecase(pool)
	invoiceReturnUsecase := usecase.NewInvoiceReturnUsecase(pool)
	invoiceObjectUsecase := usecase.NewInvoiceObjectUsecase(pool)
	invoiceCorrectionUsecase := usecase.NewInvoiceCorrectionUsecase(pool)
	kl04kvObjectUsecase := usecase.NewKL04KVObjectUsecase(pool)
	materialCostUsecase := usecase.NewMaterialCostUsecase(queries)
	materialLocationUsecase := usecase.NewMaterialLocationUsecase(pool)
	materialUsecase := usecase.NewMaterialUsecase(queries)
	mjdObjectUsecase := usecase.NewMJDObjectUsecase(pool)
	objectUsecase := usecase.NewObjectUsecase(queries)
	operationUsecase := usecase.NewOperationUsecase(pool)
	projectUsecase := usecase.NewProjectUsecase(pool)
	sipObjectUsecase := usecase.NewSIPObjectUsecase(pool)
	stvtObjectUsecase := usecase.NewSTVTObjectUsecase(pool)
	teamUsecase := usecase.NewTeamUsecase(pool)
	tpObjectUsecase := usecase.NewTPObjectUsecase(pool)
	userUsecase := usecase.NewUserUsecase(pool)
	workerUsecase := usecase.NewWorkerUsecase(queries)
	districtUsecase := usecase.NewDistrictUsecase(queries)
	permissionUsecase := usecase.NewPermissionUsecase(queries)
	roleUsecase := usecase.NewRoleUsecase(queries)
	resourceUsecase := usecase.NewResourceUsecase(queries)
	substationObjectUsecase := usecase.NewSubstationObjectUsecase(pool)
	invoiceWriteOffUsecase := usecase.NewInvoiceWriteOffUsecase(pool)
	workerAttendanceUsecase := usecase.NewWorkerAttendanceUsecase(queries)
	mainReportUsecase := usecase.NewMainReportUsecase(queries)
	substationCellObjectUsecase := usecase.NewSubstationCellObjectUsecase(pool)
	statisticsUsecase := usecase.NewStatisticsUsecase(queries)

	auctionHandler := handlers.NewAuctionHandler(auctionUsecase)
	invoiceInputHandler := handlers.NewInvoiceInputHandler(invoiceInputUsecase, userActionUsecase)
	invoiceOutputHandler := handlers.NewInvoiceOutputHandler(invoiceOutputUsecase)
	invoiceReturnHandler := handlers.NewInvoiceReturnHandler(invoiceReturnUsecase)
	materialHandler := handlers.NewMaterialHandler(materialUsecase)
	materialCostHandler := handlers.NewMaterialCostHandler(materialCostUsecase)
	materialLocationHandler := handlers.NewMaterialLocationHandler(materialLocationUsecase)
	objectHandler := handlers.NewObjectHandler(objectUsecase)
	operationHandler := handlers.NewOperationHandler(operationUsecase)
	projectHandler := handlers.NewProjectHandler(projectUsecase)
	teamHandler := handlers.NewTeamHandler(teamUsecase)
	userHandler := handlers.NewUserHandler(userUsecase, userActionUsecase)
	userActionHandler := handlers.NewUserActionHandler(userActionUsecase)
	workerHandler := handlers.NewWorkerHandler(workerUsecase)
	districtHandler := handlers.NewDistrictHandler(districtUsecase, userActionUsecase)
	permissionHandler := handlers.NewPermissionHandler(permissionUsecase)
	roleHandler := handlers.NewRoleHandler(roleUsecase)
	resourceHandler := handlers.NewResourceHandler(resourceUsecase)
	invoiceObjectHandler := handlers.NewInvoiceObjectHandler(invoiceObjectUsecase)
	invoiceCorrectionHandler := handlers.NewInvoiceCorrectionHandler(invoiceCorrectionUsecase)
	kl04kvObjectHandler := handlers.NewKl04KVObjectHandler(kl04kvObjectUsecase)
	mjdObjectHandler := handlers.NewMJDObjectHandler(mjdObjectUsecase)
	sipObjectHandler := handlers.NewSIPObjectHandler(sipObjectUsecase)
	stvtObjectHandler := handlers.NewSTVTObjectHandler(stvtObjectUsecase)
	tpObjectHandler := handlers.NewTPObjectHandler(tpObjectUsecase)
	substationObjectHandler := handlers.NewSubstationObjectHandler(substationObjectUsecase)
	invoiceOutputOutOfProjectHandler := handlers.NewInvoiceOutputOutOfProjectHandler(invoiceOutputOutOfProjectUsecase)
	invoiceWriteOffHandler := handlers.NewInvoiceWriteOffHandler(invoiceWriteOffUsecase)
	workerAttendanceHandler := handlers.NewWorkerAttendanceHandler(workerAttendanceUsecase)
	mainReportHandler := handlers.NewMainReportHandler(mainReportUsecase)
	substationCellObjectHandler := handlers.NewSubstationCellObjectHandler(substationCellObjectUsecase)
	statisticsHandler := handlers.NewStatisticsHandler(statisticsUsecase)

	InitAuctionRoutes(router, auctionHandler, resolver)
	InitInvoiceInputRoutes(router, invoiceInputHandler, resolver)
	InitInvoiceOutputRoutes(router, invoiceOutputHandler, resolver)
	InitInvoiceReturnRoutes(router, invoiceReturnHandler, resolver)
	InitProjectRoutes(router, projectHandler, resolver)
	InitMaterialRoutes(router, materialHandler, resolver)
	InitMaterialLocationRoutes(router, materialLocationHandler, resolver)
	InitTeamRoutes(router, teamHandler, resolver)
	InitObjectRoutes(router, objectHandler, resolver)
	InitWorkerRoutes(router, workerHandler, resolver)
	InitUserRoutes(router, userHandler, permissionHandler, resolver)
	InitDistrictRoutes(router, districtHandler, resolver)
	InitMaterialCostRoutes(router, materialCostHandler, resolver)
	InitPermissionRoutes(router, permissionHandler, resolver)
	InitRoleRoutes(router, roleHandler, resolver)
	InitResourceRoutes(router, resourceHandler, resolver)
	InitInvoiceObjectRoutes(router, invoiceObjectHandler, resolver)
	InitInvoiceCorrectionRoutes(router, invoiceCorrectionHandler, resolver)
	InitKL04KVObjectRoutes(router, kl04kvObjectHandler, resolver)
	InitMJDObjectRoutes(router, mjdObjectHandler, resolver)
	InitSIPObjectRoutes(router, sipObjectHandler, resolver)
	InitSTVTObjectRoutes(router, stvtObjectHandler, resolver)
	InitTPObjectRoutes(router, tpObjectHandler, resolver)
	InitSubstationObjectRoutes(router, substationObjectHandler, resolver)
	InitInvoiceOutputOutOfProjectRoutes(router, invoiceOutputOutOfProjectHandler, resolver)
	InitOperationRoutes(router, operationHandler, resolver)
	InitInvoiceWriteOffRoutes(router, invoiceWriteOffHandler, resolver)
	InitWorkerAttendanceRoutes(router, workerAttendanceHandler, resolver)
	InitMainReports(router, mainReportHandler, resolver)
	InitSubstationCellRoutes(router, substationCellObjectHandler, resolver)
	InitStatisticsRoutes(router, statisticsHandler, resolver)
	InitUserActionRoutes(router, userActionHandler, resolver)

	_ = db
	return mainRouter
}

func InitUserActionRoutes(router *gin.RouterGroup, handler handlers.IUserActionHandler, resolver auth.Resolver) {
	g := router.Group("/user-action")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResAdminUserAction)
	g.GET("/paginated",       gate(auth.ActionView), handler.GetPaginated)
	g.GET("/user/:userID",    gate(auth.ActionView), handler.GetAllByUserID)
	g.GET("/filter-users",    gate(auth.ActionView), handler.GetFilterUsers)
}

func InitStatisticsRoutes(router *gin.RouterGroup, handler handlers.IStatisticsHandler, resolver auth.Resolver) {
	g := router.Group("/statistics")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResReportStatistics)
	g.GET("/invoice-count",                gate(auth.ActionView), handler.InvoiceCountStat)
	g.GET("/invoice-input-creator",        gate(auth.ActionView), handler.InvoiceInputCreatorStat)
	g.GET("/invoice-output-creator",       gate(auth.ActionView), handler.InvoiceOutputCreatorStat)
	g.GET("/material/invoice/:materialID", gate(auth.ActionView), handler.MaterialInInvoice)
	g.GET("/material/location/:materialID", gate(auth.ActionView), handler.MaterialInLocations)
}

func InitAuctionRoutes(router *gin.RouterGroup, handler handlers.IAuctionHandler, resolver auth.Resolver) {
	g := router.Group("/auction")
	// Public — anyone can view a public auction.
	g.GET("/:auctionID", handler.GetAuctionDataForPublic)

	gatePriv := middleware.GroupGate(resolver, auth.ResAuctionBidPrivate)
	g.GET("/private/:auctionID", middleware.Authentication(), gatePriv(auth.ActionView),   handler.GetAuctionDataForPrivate)
	g.POST("/private",           middleware.Authentication(), gatePriv(auth.ActionCreate), handler.SaveParticipantChanges)
}

func InitMainReports(router *gin.RouterGroup, handler handlers.IMainReportHandler, resolver auth.Resolver) {
	g := router.Group("/main-reports")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResReportStatistics)
	g.POST("/project-progress",              gate(auth.ActionReport), handler.ProjectProgress)
	g.GET("/analysis-of-remaining-materials", gate(auth.ActionReport), handler.RemainingMaterialAnalysis)
}

func InitWorkerAttendanceRoutes(router *gin.RouterGroup, handler handlers.IWorkerAttendanceHandler, resolver auth.Resolver) {
	g := router.Group("/worker-attendance")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResHRAttendance)
	g.GET("/paginated", gate(auth.ActionView),   handler.GetPaginated)
	g.POST("/",         gate(auth.ActionImport), handler.Import)
}

func InitInvoiceWriteOffRoutes(router *gin.RouterGroup, handler handlers.IInvoiceWriteOffHandler, resolver auth.Resolver) {
	g := router.Group("/invoice-writeoff")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResInvoiceWriteoff)
	g.GET("/paginated",                                   gate(auth.ActionView),    handler.GetPaginated)
	g.GET("/:id/materials/without-serial-number",         gate(auth.ActionView),    handler.GetInvoiceMaterialsWithoutSerialNumber)
	g.GET("/invoice-materials/:id/:locationType/:locationID", gate(auth.ActionView), handler.GetMaterialsForEdit)
	g.GET("/document/:deliveryCode",                      gate(auth.ActionExport),  handler.GetDocument)
	g.GET("/material/:locationType/:locationID",          gate(auth.ActionView),    handler.GetMaterialsInLocation)
	g.POST("/",                                           gate(auth.ActionCreate),  handler.Create)
	g.POST("/confirm/:id",                                gate(auth.ActionConfirm), handler.Confirmation)
	g.POST("/report",                                     gate(auth.ActionReport),  handler.Report)
	g.PATCH("/",                                          gate(auth.ActionEdit),    handler.Update)
	g.DELETE("/:id",                                      gate(auth.ActionDelete),  handler.Delete)
}

func InitInvoiceOutputOutOfProjectRoutes(router *gin.RouterGroup, handler handlers.IInvoiceOutputOutOfProjectHandler, resolver auth.Resolver) {
	g := router.Group("/invoice-output-out-of-project")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResInvoiceOutputOutOfProject)
	g.GET("/paginated",                                gate(auth.ActionView),    handler.GetPaginated)
	g.GET("/:id/materials/without-serial-number",      gate(auth.ActionView),    handler.GetInvoiceMaterialsWithoutSerialNumbers)
	g.GET("/:id/materials/with-serial-number",         gate(auth.ActionView),    handler.GetInvoiceMaterialsWithSerialNumbers)
	g.GET("/invoice-materials/:id",                    gate(auth.ActionView),    handler.GetMaterialsForEdit)
	g.GET("/unique/name-of-project",                   gate(auth.ActionView),    handler.UniqueNameOfProjects)
	g.GET("/document/:deliveryCode",                   gate(auth.ActionExport),  handler.GetDocument)
	g.POST("/",                                        gate(auth.ActionCreate),  handler.Create)
	g.POST("/report",                                  gate(auth.ActionReport),  handler.Report)
	g.PATCH("/",                                       gate(auth.ActionEdit),    handler.Update)
	g.POST("/confirm/:id",                             gate(auth.ActionConfirm), handler.Confirmation)
	g.DELETE("/:id",                                   gate(auth.ActionDelete),  handler.Delete)
}

func InitInvoiceCorrectionRoutes(router *gin.RouterGroup, handler handlers.IInvoiceCorrectionHandler, resolver auth.Resolver) {
	g := router.Group("/invoice-correction")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResInvoiceCorrection)
	g.GET("/paginated",                                                  gate(auth.ActionView),    handler.GetPaginated)
	g.GET("/",                                                           gate(auth.ActionView),    handler.GetAll)
	g.GET("/materials/:id",                                              gate(auth.ActionView),    handler.GetInvoiceMaterialsByInvoiceObjectID)
	g.GET("/operations/:id",                                             gate(auth.ActionView),    handler.GetOperationsByInvoiceObjectID)
	g.GET("/total-amount/:materialID/team/:teamNumber",                  gate(auth.ActionView),    handler.GetTotalMaterialInTeamByTeamNumber)
	g.GET("/serial-number/material/:materialID/teams/:teamNumber",       gate(auth.ActionView),    handler.GetSerialNumbersOfMaterial)
	g.GET("unique/team",                                                 gate(auth.ActionView),    handler.UniqueTeam)
	g.GET("unique/object",                                               gate(auth.ActionView),    handler.UniqueObject)
	g.POST("/report",                                                    gate(auth.ActionReport),  handler.Report)
	g.POST("/",                                                          gate(auth.ActionCorrect), handler.Create)
	g.GET("/search-parameters",                                          gate(auth.ActionView),    handler.GetParametersForSearch)
}

func InitInvoiceObjectRoutes(router *gin.RouterGroup, handler handlers.IInvoiceObjectHandler, resolver auth.Resolver) {
	g := router.Group("/invoice-object")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResInvoiceObject)
	g.GET("/:id",                                                         gate(auth.ActionView),   handler.GetInvoiceObjectDescriptiveDataByID)
	g.GET("/paginated",                                                   gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/team-materials/:teamID",                                      gate(auth.ActionView),   handler.GetTeamsMaterials)
	g.GET("/available-operations-for-team/:teamID",                       gate(auth.ActionView),   handler.GetOperationsBasedOnMaterialsInTeamID)
	g.GET("/serial-number/material/:materialID/teams/:teamID",            gate(auth.ActionView),   handler.GetSerialNumbersOfMaterial)
	g.GET("/material/:materialID/team/:teamID",                           gate(auth.ActionView),   handler.GetMaterialAmountInTeam)
	g.GET("/object/:objectID",                                            gate(auth.ActionView),   handler.GetTeamsFromObjectID)
	g.POST("/",                                                           gate(auth.ActionCreate), handler.Create)
}

func InitInvoiceReturnRoutes(router *gin.RouterGroup, handler handlers.IInvoiceReturnHandler, resolver auth.Resolver) {
	g := router.Group("/return")
	g.Use(middleware.Authentication())
	// /return covers both team-return and object-return invoice flavors. Gate with
	// the team variant by default; object-return-specific endpoints get their own
	// gate. The non-shared paths below are object-specific by URL convention.
	gateTeam := middleware.GroupGate(resolver, auth.ResInvoiceReturnTeam)
	g.GET("/paginated",                                          gateTeam(auth.ActionView),    handler.GetPaginated)
	g.GET("/unique/code",                                        gateTeam(auth.ActionView),    handler.UniqueCode)
	g.GET("/unique/team",                                        gateTeam(auth.ActionView),    handler.UniqueTeam)
	g.GET("/unique/object",                                      gateTeam(auth.ActionView),    handler.UniqueObject)
	g.GET("/document/:deliveryCode",                             gateTeam(auth.ActionExport),  handler.GetDocument)
	g.GET("/material/:locationType/:locationID",                 gateTeam(auth.ActionView),    handler.GetMaterialsInLocation)
	g.GET("/material-cost/:materialID/:locationType/:locationID", gateTeam(auth.ActionView),   handler.GetUniqueMaterialCostsFromLocation)
	g.GET("/material-amount/:materialCostID/:locationType/:locationID", gateTeam(auth.ActionView), handler.GetMaterialAmountInLocation)
	g.GET("/serial-number/:locationType/:locationID/:materialID", gateTeam(auth.ActionView),   handler.GetSerialNumberCodesInLocation)
	g.GET("/:id/materials/without-serial-number",                gateTeam(auth.ActionView),    handler.GetInvoiceMaterialsWithoutSerialNumbers)
	g.GET("/:id/materials/with-serial-number",                   gateTeam(auth.ActionView),    handler.GetInvoiceMaterialsWithSerialNumbers)
	g.GET("/invoice-materials/:id/:locationType/:locationID",    gateTeam(auth.ActionView),    handler.GetMaterialsForEdit)
	g.GET("/amount/:locationType/:locationID/:materialID",       gateTeam(auth.ActionView),    handler.GetMaterialAmountByMaterialID)
	g.POST("/confirm/:id",                                       gateTeam(auth.ActionConfirm), handler.Confirmation)
	g.POST("/",                                                  gateTeam(auth.ActionCreate),  handler.Create)
	g.POST("/report",                                            gateTeam(auth.ActionReport),  handler.Report)
	g.PATCH("/",                                                 gateTeam(auth.ActionEdit),    handler.Update)
	g.DELETE("/:id",                                             gateTeam(auth.ActionDelete),  handler.Delete)
}

func InitInvoiceOutputRoutes(router *gin.RouterGroup, handler handlers.IInvoiceOutputHandler, resolver auth.Resolver) {
	g := router.Group("/output")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResInvoiceOutput)
	g.GET("/paginated",                                  gate(auth.ActionView),    handler.GetPaginated)
	g.GET("/unique/district",                            gate(auth.ActionView),    handler.UniqueDistrict)
	g.GET("/unique/code",                                gate(auth.ActionView),    handler.UniqueCode)
	g.GET("/unique/recieved",                            gate(auth.ActionView),    handler.UniqueRecieved)
	g.GET("/unique/warehouse-manager",                   gate(auth.ActionView),    handler.UniqueWarehouseManager)
	g.GET("/unique/team",                                gate(auth.ActionView),    handler.UniqueTeam)
	g.GET("/document/:deliveryCode",                     gate(auth.ActionExport),  handler.GetDocument)
	g.GET("/material/available-in-warehouse",            gate(auth.ActionView),    handler.GetAvailableMaterialsInWarehouse)
	g.GET("/material/:materialID/total-amount",          gate(auth.ActionView),    handler.GetTotalAmountInWarehouse)
	g.GET("/serial-number/material/:materialID",         gate(auth.ActionView),    handler.GetCodesByMaterialID)
	g.GET("/:id/materials/without-serial-number",        gate(auth.ActionView),    handler.GetInvoiceMaterialsWithoutSerialNumbers)
	g.GET("/:id/materials/with-serial-number",           gate(auth.ActionView),    handler.GetInvoiceMaterialsWithSerialNumbers)
	g.GET("/invoice-materials/:id",                      gate(auth.ActionView),    handler.GetMaterialsForEdit)
	g.POST("/report",                                    gate(auth.ActionReport),  handler.Report)
	g.POST("/confirm/:id",                               gate(auth.ActionConfirm), handler.Confirmation)
	g.POST("/",                                          gate(auth.ActionCreate),  handler.Create)
	g.POST("/import",                                    gate(auth.ActionImport),  handler.Import)
	g.PATCH("/",                                         gate(auth.ActionEdit),    handler.Update)
	g.DELETE("/:id",                                     gate(auth.ActionDelete),  handler.Delete)
}

func InitInvoiceInputRoutes(router *gin.RouterGroup, handler handlers.IInvoiceInputHandler, resolver auth.Resolver) {
	g := router.Group("/input")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResInvoiceInput)
	g.GET("/paginated",                            gate(auth.ActionView),    handler.GetPaginated)
	g.GET("/unique/code",                          gate(auth.ActionView),    handler.UniqueCode)
	g.GET("/unique/warehouse-manager",             gate(auth.ActionView),    handler.UniqueWarehouseManager)
	g.GET("/unique/released",                      gate(auth.ActionView),    handler.UniqueReleased)
	g.GET("/document/:deliveryCode",               gate(auth.ActionExport),  handler.GetDocument)
	g.GET("/:id/materials/without-serial-number",  gate(auth.ActionView),    handler.GetInvoiceMaterialsWithoutSerialNumbers)
	g.GET("/:id/materials/with-serial-number",     gate(auth.ActionView),    handler.GetInvoiceMaterialsWithSerialNumbers)
	g.GET("/invoice-materials/:id",                gate(auth.ActionView),    handler.GetMaterialsForEdit)
	g.GET("/search-parameters",                    gate(auth.ActionView),    handler.GetParametersForSearch)
	g.POST("/",                                    gate(auth.ActionCreate),  handler.Create)
	g.POST("/report",                              gate(auth.ActionReport),  handler.Report)
	g.POST("/confirm/:id",                         gate(auth.ActionConfirm), handler.Confirmation)
	g.POST("/material/new",                        gate(auth.ActionCreate),  handler.NewMaterial)
	g.POST("/material-cost/new",                   gate(auth.ActionCreate),  handler.NewMaterialCost)
	g.POST("/import",                              gate(auth.ActionImport),  handler.Import)
	g.PATCH("/",                                   gate(auth.ActionEdit),    handler.Update)
	g.DELETE("/:id",                               gate(auth.ActionDelete),  handler.Delete)
}

func InitProjectRoutes(router *gin.RouterGroup, handler handlers.IProjectHandler, resolver auth.Resolver) {
	g := router.Group("/project")
	// /project/all is intentionally public — used by the login flow to populate
	// the "select project" dropdown before the user has a token.
	g.GET("/all", handler.GetAll)

	g.Use(middleware.Authentication())
	gateRef := middleware.GroupGate(resolver, auth.ResRefProject)
	gateAdm := middleware.GroupGate(resolver, auth.ResAdminProject)
	g.GET("/paginated", gateRef(auth.ActionView),   handler.GetPaginated)
	g.GET("/name",      gateRef(auth.ActionView),   handler.GetProjectName)
	g.POST("/",         gateAdm(auth.ActionCreate), handler.Create)
	g.PATCH("/",        gateAdm(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",    gateAdm(auth.ActionDelete), handler.Delete)
}

func InitMaterialLocationRoutes(router *gin.RouterGroup, handler handlers.IMaterialLocationHandler, resolver auth.Resolver) {
	g := router.Group("/material-location")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefMaterialLocation)
	gateLive := middleware.GroupGate(resolver, auth.ResSystemMaterialLocLive)
	gateRpt := middleware.GroupGate(resolver, auth.ResReportBalance)
	g.GET("/available/:locationType/:locationID",                       gate(auth.ActionView),    handler.GetMaterialInLocation)
	g.GET("/costs/:materialID/:locationType/:locationID",               gate(auth.ActionView),    handler.GetMaterialCostsInLocation)
	g.GET("/amount/:materialCostID/:locationType/:locationID",          gate(auth.ActionView),    handler.GetMaterialAmountBasedOnCost)
	g.GET("/unique/team",                                               gate(auth.ActionView),    handler.UniqueTeams)
	g.GET("/unique/object",                                             gate(auth.ActionView),    handler.UniqueObjects)
	g.GET("/live",                                                      gateLive(auth.ActionView), handler.Live)
	g.POST("/report/balance",                                           gateRpt(auth.ActionReport), handler.ReportBalance)
	g.POST("/report/balance/writeoff",                                  gateRpt(auth.ActionReport), handler.ReportBalanceWriteOff)
	g.POST("/report/balance/out-of-project",                            gateRpt(auth.ActionReport), handler.ReportBalanceOutOfProject)
}

func InitMaterialCostRoutes(router *gin.RouterGroup, handler handlers.IMaterialCostHandler, resolver auth.Resolver) {
	g := router.Group("/material-cost")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefMaterialCost)
	g.GET("/paginated",                  gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/material-id/:materialID",    gate(auth.ActionView),   handler.GetAllMaterialCostByMaterialID)
	g.GET("/document/template",          gate(auth.ActionExport), handler.ImportTemplate)
	g.GET("/document/export",            gate(auth.ActionExport), handler.Export)
	g.POST("/document/import",           gate(auth.ActionImport), handler.Import)
	g.POST("/",                          gate(auth.ActionCreate), handler.Create)
	g.PATCH("/",                         gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",                     gate(auth.ActionDelete), handler.Delete)
}

func InitMaterialRoutes(router *gin.RouterGroup, handler handlers.IMaterialHandler, resolver auth.Resolver) {
	g := router.Group("/material")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefMaterial)
	g.GET("/all",                gate(auth.ActionView),   handler.GetAll)
	g.GET("/paginated",          gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/document/template",  gate(auth.ActionExport), handler.GetTemplateFile)
	g.GET("/document/export",    gate(auth.ActionExport), handler.Export)
	g.POST("/",                  gate(auth.ActionCreate), handler.Create)
	g.POST("/document/import",   gate(auth.ActionImport), handler.Import)
	g.PATCH("/",                 gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",             gate(auth.ActionDelete), handler.Delete)
}

func InitDistrictRoutes(router *gin.RouterGroup, handler handlers.IDistictHandler, resolver auth.Resolver) {
	g := router.Group("/district")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefDistrict)
	g.GET("/all",        gate(auth.ActionView),   handler.GetAll)
	g.GET("/paginated",  gate(auth.ActionView),   handler.GetPaginated)
	g.POST("/",          gate(auth.ActionCreate), handler.Create)
	g.PATCH("/",         gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",     gate(auth.ActionDelete), handler.Delete)
}

func InitTeamRoutes(router *gin.RouterGroup, handler handlers.ITeamHandler, resolver auth.Resolver) {
	g := router.Group("/team")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefTeam)
	g.GET("/all",                       gate(auth.ActionView),   handler.GetAll)
	g.GET("/all/for-select",            gate(auth.ActionView),   handler.GetAllForSelect)
	g.GET("/paginated",                 gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/:id",                       gate(auth.ActionView),   handler.GetByID)
	g.GET("/document/template",         gate(auth.ActionExport), handler.GetTemplateFile)
	g.GET("/unique/team-number",        gate(auth.ActionView),   handler.GetAllUniqueTeamNumbers)
	g.GET("/unique/mobile-number",      gate(auth.ActionView),   handler.GetAllUniqueMobileNumber)
	g.GET("/unique/team-company",       gate(auth.ActionView),   handler.GetAllUniqueCompanies)
	g.GET("/document/export",           gate(auth.ActionExport), handler.Export)
	g.POST("/",                         gate(auth.ActionCreate), handler.Create)
	g.POST("/document/import",          gate(auth.ActionImport), handler.Import)
	g.PATCH("/",                        gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",                    gate(auth.ActionDelete), handler.Delete)
}

func InitObjectRoutes(router *gin.RouterGroup, handler handlers.IObjectHandler, resolver auth.Resolver) {
	g := router.Group("/object")
	g.Use(middleware.Authentication())
	// Generic object endpoints — gate with invoice.object since these power the
	// invoice-object UX. Per-type CRUD lives in /kl04kv, /mjd, etc. with their
	// specific resource gates.
	gate := middleware.GroupGate(resolver, auth.ResInvoiceObject)
	g.GET("/all",                     gate(auth.ActionView),   handler.GetAll)
	g.GET("/paginated",               gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/:id",                     gate(auth.ActionView),   handler.GetByID)
	g.GET("/teams/:objectID",         gate(auth.ActionView),   handler.GetTeamsByObject)
	g.POST("/",                       gate(auth.ActionCreate), handler.Create)
	g.PATCH("/",                      gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",                  gate(auth.ActionDelete), handler.Delete)
}

func initObjectFamilyRoutes(g *gin.RouterGroup, gate func(auth.Action) gin.HandlerFunc, h interface {
	GetPaginated(c *gin.Context)
	GetTemplateFile(c *gin.Context)
	Export(c *gin.Context)
	Create(c *gin.Context)
	Import(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetObjectNamesForSearch(c *gin.Context)
}) {
	g.GET("/paginated",          gate(auth.ActionView),   h.GetPaginated)
	g.GET("/document/template",  gate(auth.ActionExport), h.GetTemplateFile)
	g.GET("/document/export",    gate(auth.ActionExport), h.Export)
	g.GET("/search/object-names", gate(auth.ActionView),  h.GetObjectNamesForSearch)
	g.POST("/",                  gate(auth.ActionCreate), h.Create)
	g.POST("/document/import",   gate(auth.ActionImport), h.Import)
	g.PATCH("/",                 gate(auth.ActionEdit),   h.Update)
	g.DELETE("/:id",             gate(auth.ActionDelete), h.Delete)
}

func InitTPObjectRoutes(router *gin.RouterGroup, handler handlers.ITPObjectHandler, resolver auth.Resolver) {
	g := router.Group("/tp")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectTP)
	g.GET("/all", gate(auth.ActionView), handler.GetAll)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitSubstationCellRoutes(router *gin.RouterGroup, handler handlers.ISubstationCellObjectHandler, resolver auth.Resolver) {
	g := router.Group("/cell-substation")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectSubstationCell)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitSubstationObjectRoutes(router *gin.RouterGroup, handler handlers.ISubstationObjectHandler, resolver auth.Resolver) {
	g := router.Group("/substation")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectSubstation)
	g.GET("/all", gate(auth.ActionView), handler.GetAll)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitSTVTObjectRoutes(router *gin.RouterGroup, handler handlers.ISTVTObjectHandler, resolver auth.Resolver) {
	g := router.Group("/stvt")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectSTVT)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitSIPObjectRoutes(router *gin.RouterGroup, handler handlers.ISIPObjectHandler, resolver auth.Resolver) {
	g := router.Group("/sip")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectSIP)
	g.GET("/tp-names", gate(auth.ActionView), handler.GetTPNames)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitMJDObjectRoutes(router *gin.RouterGroup, handler handlers.IMJDObjectHandler, resolver auth.Resolver) {
	g := router.Group("/mjd")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectMJD)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitKL04KVObjectRoutes(router *gin.RouterGroup, handler handlers.IKL04KVObjectHandler, resolver auth.Resolver) {
	g := router.Group("/kl04kv")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefObjectKL04KV)
	initObjectFamilyRoutes(g, gate, handler)
}

func InitWorkerRoutes(router *gin.RouterGroup, handler handlers.IWorkerHandler, resolver auth.Resolver) {
	g := router.Group("/worker")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefWorker)
	g.GET("/all",                              gate(auth.ActionView),   handler.GetAll)
	g.GET("/paginated",                        gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/:id",                              gate(auth.ActionView),   handler.GetByID)
	g.GET("/job-title/:jobTitleInProject",     gate(auth.ActionView),   handler.GetByJobTitleInProject)
	g.GET("/document/template",                gate(auth.ActionExport), handler.GetTemplateFile)
	g.GET("/unique/worker-information",        gate(auth.ActionView),   handler.GetWorkerInformationForSearch)
	g.GET("/document/export",                  gate(auth.ActionExport), handler.Export)
	g.POST("/",                                gate(auth.ActionCreate), handler.Create)
	g.POST("/document/import",                 gate(auth.ActionImport), handler.Import)
	g.PATCH("/",                               gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",                           gate(auth.ActionDelete), handler.Delete)
}

func InitUserRoutes(router *gin.RouterGroup, handler handlers.IUserHandler, permissionHandler handlers.IPermissionHandler, resolver auth.Resolver) {
	g := router.Group("/user")

	// Public — login is the entry point; is-authenticated verifies its own token.
	g.POST("/login",            handler.Login)
	g.GET("/is-authenticated",  handler.IsAuthenticated)

	authed := g.Group("")
	authed.Use(middleware.Authentication())

	// Self-query — any authenticated user reads their own permission set.
	authed.GET("/effective-permissions",  permissionHandler.GetCurrentUserEffectivePermissions)

	gate := middleware.GroupGate(resolver, auth.ResAdminUser)
	authed.GET("/all",         gate(auth.ActionView),   handler.GetAll)
	authed.GET("/paginated",   gate(auth.ActionView),   handler.GetPaginated)
	authed.GET("/:id",         gate(auth.ActionView),   handler.GetByID)
	authed.POST("/",           gate(auth.ActionCreate), handler.Create)
	authed.PATCH("/:id",       gate(auth.ActionEdit),   handler.Update)
	authed.DELETE("/:id",      gate(auth.ActionDelete), handler.Delete)
}

func InitPermissionRoutes(router *gin.RouterGroup, handler handlers.IPermissionHandler, resolver auth.Resolver) {
	g := router.Group("/permission")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResAdminRoleGrant)
	g.GET("/all",                       gate(auth.ActionView),   handler.GetAll)
	g.GET("/role/name/:roleName",       gate(auth.ActionView),   handler.GetByRoleName)
	g.GET("/role/url/:resourceURL",     gate(auth.ActionView),   handler.GetByResourceURL)
	g.POST("/",                         gate(auth.ActionCreate), handler.Create)
	g.POST("/batch",                    gate(auth.ActionCreate), handler.CreateBatch)
	g.PATCH("/",                        gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",                    gate(auth.ActionDelete), handler.Delete)
}

func InitRoleRoutes(router *gin.RouterGroup, handler handlers.IRoleHandler, resolver auth.Resolver) {
	g := router.Group("/role")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResAdminRole)
	g.GET("/all",     gate(auth.ActionView),   handler.GetAll)
	g.POST("/",       gate(auth.ActionCreate), handler.Create)
	g.PATCH("/",      gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",  gate(auth.ActionDelete), handler.Delete)
}

func InitResourceRoutes(router *gin.RouterGroup, handler handlers.IResourceHandler, resolver auth.Resolver) {
	g := router.Group("/resource")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResAdminResourceType)
	g.GET("/", gate(auth.ActionView), handler.GetAll)
}

func InitOperationRoutes(router *gin.RouterGroup, handler handlers.IOperationHandler, resolver auth.Resolver) {
	g := router.Group("/operation")
	g.Use(middleware.Authentication())
	gate := middleware.GroupGate(resolver, auth.ResRefOperation)
	g.GET("/paginated",          gate(auth.ActionView),   handler.GetPaginated)
	g.GET("/all",                gate(auth.ActionView),   handler.GetAll)
	g.GET("/document/template",  gate(auth.ActionExport), handler.GetTemplateFile)
	g.POST("/",                  gate(auth.ActionCreate), handler.Create)
	g.POST("/document/import",   gate(auth.ActionImport), handler.Import)
	g.PATCH("/",                 gate(auth.ActionEdit),   handler.Update)
	g.DELETE("/:id",             gate(auth.ActionDelete), handler.Delete)
}
