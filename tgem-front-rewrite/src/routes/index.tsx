import { LOGIN, PAGE_NOT_FOUND, PERMISSION_DENIED, HOME, INVOICE_INPUT, REFERENCE_BOOK_WORKER, REFERENCE_BOOK_OBJECTS, REFERENCE_BOOK_OPERATIONS, REFERENCE_BOOK_MATERIAL_COST, REFERENCE_BOOK_DISTRICT, INVOICE_OBJECT_USER, REPORT, ADMIN_USERS_PAGE, REFERENCE_BOOK_TEAM, REFERENCE_BOOK_MATERIAL, INVOICE_OBJECT_MUTATION_ADD, INVOICE_OBJECT_DETAILS, INVOICE_CORRECTION, REFERENCE_BOOK_KL04KV_OBJECT, REFERENCE_BOOK_MJD_OBJECT, REFERENCE_BOOK_SIP_OBJECT, REFERENCE_BOOK_STVT_OBJECT, REFERENCE_BOOK_TP_OBJECT, INVOICE_RETURN_TEAM, INVOICE_RETURN_OBJECT, REFERENCE_BOOK, REFERENCE_BOOK_SUBSTATION_OBJECT, INVOICE_OUTPUT_IN_PROJECT, INVOICE_OUTPUT_OUT_OF_PROJECT, ADMINISTRATOR_HOME_PAGE, ADMINISTRATOR_PROJECT, ADMINISTRATOR_USERS, ADMINISTRATOR_USER_ACTIONS, IMPORT, MATERIAL_LOCATION_LIVE, WRITEOFF_WAREHOUSE, LOSS_WAREHOUSE, LOSS_TEAM, LOSS_OBJECT, WRITEOFF_OBJECT, HR_ATTENDANCE, REFERENCE_BOOK_SUBSTATION_CELL_OBJECT, AUCTION_PUBLIC, AUCTION_PRIVATE, INVOICE_OBJECT_PAGINATED_PAGE, STATISTICS, ADMINISTRATOR_WORKERS, } from "./paths"
import Home from "@features/home/HomePage"
import Login from "@features/system/login/LoginPage"
import PermissionDenied from "@features/system/permission-denied/PermissionDeniedPage"
import ReferenceBooks from "@features/reference-books/menu/ReferenceBooksPage"
import ErrorPage from "@features/system/error/ErrorPage"
import InvoiceInput from "@features/invoice/input/InvoiceInputPage"
import InvoiceObjectUser from "@features/invoice/object/InvoiceObjectUserPage"
import District from "@features/reference-books/district/DistrictsPage"
import Materials from "@features/reference-books/material/MaterialsPage"
import MaterialsCosts from "@features/reference-books/material-cost/MaterialCostsPage"
import Objects from "@features/reference-books/object/ObjectsPage"
import Operatons from "@features/reference-books/operation/OperationsPage"
import Team from "@features/reference-books/team/TeamsPage"
import Worker from "@features/reference-books/worker/WorkersPage"
import Report from "@features/report/ReportPage"
import AdminUserPage from "@features/admin/users/AdminUsersPage"
import InvoiceObjectMutationAdd from "@features/invoice/object/InvoiceObjectMutationMobilePage"
import InvoiceObjectDetails from "@features/invoice/object/InvoiceObjectDetailPage"
import InvoiceCorrection from "@features/invoice/correction/InvoiceCorrectionPage"
import KL04KVObject from "@features/reference-books/object/KL04KVObjectPage"
import MJDObject from "@features/reference-books/object/MJDObjectPage"
import SIPObject from "@features/reference-books/object/SIPObjectPage"
import STVTObject from "@features/reference-books/object/STVTObjectPage"
import TPObject from "@features/reference-books/object/TPObjectPage"
import InvoiceReturnTeam from "@features/invoice/return-team/InvoiceReturnTeamPage"
import InvoiceReturnObject from "@features/invoice/return-object/InvoiceReturnObjectPage"
import SubstationObject from "@features/reference-books/object/SubstationObjectPage"
import InvoiceOutputInProject from "@features/invoice/output-in-project/InvoiceOutputInProjectPage"
import InvoiceOutputOutOfProject from "@features/invoice/output-out-of-project/InvoiceOutputOutOfProjectPage"
import AdministatorHome from "@features/admin/home/AdministratorHomePage"
import { AdministratorProject } from "@features/admin/projects/AdministratorProjectPage"
import Import from '@features/import-export/import/ImportPage'
import MaterialLocationLive from "@features/material-location/MaterialLocationLivePage"
import WriteOffWarehouse from "@features/writeoff/warehouse/WriteOffWarehousePage"
import LossWarehouse from "@features/writeoff/warehouse/LossWarehousePage"
import LossTeam from "@features/writeoff/team/LossTeamPage"
import LossObject from "@features/writeoff/object/LossObjectPage"
import WriteOffObject from "@features/writeoff/object/WriteOffObjectPage"
import Attendance from "@features/hr/attendance/AttendancePage"
import SubstationCellObject from "@features/reference-books/object/SubstationCellObjectPage"
import AuctionPublic from "@features/auction/public/AuctionPublicPage"
import AuctionPrivate from "@features/auction/private/AuctionPrivatePage"
import InvoiceObjectPaginatedPage from "@features/invoice/object/InvoiceObjectPaginatedPage"
import Statistics from "@features/statistics/StatisticsPage"
import AdminUserActionsPage from "@features/admin/user-actions/AdminUserActionsPage"

// RouteGate declares how a route should be guarded.
//   - { kind: "auth" }        → user must be logged in (menu pages, dashboards)
//   - { kind: "permission" }  → user needs the (action, resource) grant
// See docs/permissions-spec.md for resource codes & action vocabulary.
export type RouteGate =
  | { kind: "auth" }
  | { kind: "permission"; action: string; resource: string }

export interface AppRoute {
  path: string
  element: React.ReactElement
  gate: RouteGate
}

const auth = (): RouteGate => ({ kind: "auth" })
const perm = (action: string, resource: string): RouteGate => ({ kind: "permission", action, resource })

export const PAGES_WITHOUT_LAYOUT = [
  { path: LOGIN,             element: <Login /> },
  { path: PAGE_NOT_FOUND,    element: <ErrorPage /> },
  { path: PERMISSION_DENIED, element: <PermissionDenied /> },
  { path: AUCTION_PUBLIC,    element: <AuctionPublic /> },
  { path: AUCTION_PRIVATE,   element: <AuctionPrivate /> },
]

export const PAGES_WITH_LAYOUT: AppRoute[] = [
  { path: HOME,                                  element: <Home />,                       gate: auth() },
  { path: REFERENCE_BOOK,                        element: <ReferenceBooks />,             gate: auth() },
  { path: REPORT,                                element: <Report />,                     gate: auth() },

  { path: INVOICE_INPUT,                         element: <InvoiceInput />,               gate: perm("view", "invoice.input") },
  { path: INVOICE_RETURN_TEAM,                   element: <InvoiceReturnTeam />,          gate: perm("view", "invoice.return_team") },
  { path: INVOICE_RETURN_OBJECT,                 element: <InvoiceReturnObject />,        gate: perm("view", "invoice.return_object") },
  { path: INVOICE_OUTPUT_IN_PROJECT,             element: <InvoiceOutputInProject />,     gate: perm("view", "invoice.output") },
  { path: INVOICE_OUTPUT_OUT_OF_PROJECT,         element: <InvoiceOutputOutOfProject />,  gate: perm("view", "invoice.output_out_of_project") },
  { path: INVOICE_OBJECT_PAGINATED_PAGE,         element: <InvoiceObjectPaginatedPage />, gate: perm("view", "invoice.object") },
  { path: INVOICE_OBJECT_USER,                   element: <InvoiceObjectUser />,          gate: perm("view", "invoice.object") },
  { path: INVOICE_OBJECT_MUTATION_ADD,           element: <InvoiceObjectMutationAdd />,   gate: perm("create", "invoice.object") },
  { path: INVOICE_OBJECT_DETAILS,                element: <InvoiceObjectDetails />,       gate: perm("view", "invoice.object") },
  { path: INVOICE_CORRECTION,                    element: <InvoiceCorrection />,          gate: perm("view", "invoice.correction") },

  { path: REFERENCE_BOOK_WORKER,                 element: <Worker />,                     gate: perm("view", "reference.worker") },
  { path: REFERENCE_BOOK_MATERIAL,               element: <Materials />,                  gate: perm("view", "reference.material") },
  { path: REFERENCE_BOOK_TEAM,                   element: <Team />,                       gate: perm("view", "reference.team") },
  { path: REFERENCE_BOOK_OBJECTS,                element: <Objects />,                    gate: auth() },
  { path: REFERENCE_BOOK_KL04KV_OBJECT,          element: <KL04KVObject />,               gate: perm("view", "reference.object.kl04kv") },
  { path: REFERENCE_BOOK_MJD_OBJECT,             element: <MJDObject />,                  gate: perm("view", "reference.object.mjd") },
  { path: REFERENCE_BOOK_SIP_OBJECT,             element: <SIPObject />,                  gate: perm("view", "reference.object.sip") },
  { path: REFERENCE_BOOK_STVT_OBJECT,             element: <STVTObject />,                 gate: perm("view", "reference.object.stvt") },
  { path: REFERENCE_BOOK_TP_OBJECT,              element: <TPObject />,                   gate: perm("view", "reference.object.tp") },
  { path: REFERENCE_BOOK_SUBSTATION_OBJECT,      element: <SubstationObject />,           gate: perm("view", "reference.object.substation") },
  { path: REFERENCE_BOOK_SUBSTATION_CELL_OBJECT, element: <SubstationCellObject />,       gate: perm("view", "reference.object.substation_cell") },
  { path: REFERENCE_BOOK_OPERATIONS,             element: <Operatons />,                  gate: perm("view", "reference.operation") },
  { path: REFERENCE_BOOK_MATERIAL_COST,          element: <MaterialsCosts />,             gate: perm("view", "reference.material_cost") },
  { path: REFERENCE_BOOK_DISTRICT,               element: <District />,                   gate: perm("view", "reference.district") },

  { path: ADMIN_USERS_PAGE,                      element: <AdminUserPage />,              gate: perm("view", "admin.user") },
  { path: IMPORT,                                element: <Import />,                     gate: perm("import", "system.import") },
  { path: MATERIAL_LOCATION_LIVE,                element: <MaterialLocationLive />,       gate: perm("view", "system.material_location_live") },

  { path: WRITEOFF_WAREHOUSE,                    element: <WriteOffWarehouse />,          gate: perm("view", "invoice.writeoff") },
  { path: LOSS_WAREHOUSE,                        element: <LossWarehouse />,              gate: perm("view", "invoice.writeoff") },
  { path: LOSS_TEAM,                             element: <LossTeam />,                   gate: perm("view", "invoice.writeoff") },
  { path: LOSS_OBJECT,                           element: <LossObject />,                 gate: perm("view", "invoice.writeoff") },
  { path: WRITEOFF_OBJECT,                       element: <WriteOffObject />,             gate: perm("view", "invoice.writeoff") },

  { path: HR_ATTENDANCE,                         element: <Attendance />,                 gate: perm("view", "hr.attendance") },
  { path: STATISTICS,                            element: <Statistics />,                 gate: perm("view", "report.statistics") },
]

export const ADMIN_PAGES: AppRoute[] = [
  { path: ADMINISTRATOR_HOME_PAGE,               element: <AdministatorHome />,           gate: auth() },
  { path: ADMINISTRATOR_PROJECT,                 element: <AdministratorProject />,       gate: perm("view", "admin.project") },
  { path: ADMINISTRATOR_USERS,                   element: <AdminUserPage />,              gate: perm("view", "admin.user") },
  { path: ADMINISTRATOR_WORKERS,                 element: <Worker />,                     gate: perm("view", "reference.worker") },
  { path: ADMINISTRATOR_USER_ACTIONS,            element: <AdminUserActionsPage />,       gate: perm("view", "admin.user_action") },
]
