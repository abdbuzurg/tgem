import { useEffect, useMemo, useState } from "react"
import { useInfiniteQuery, useQuery } from "@tanstack/react-query"
import { ENTRY_LIMIT } from "@shared/config/pagination"
import useScrollPaginated from "@shared/hooks/useScrollPaginated"
import {
  UserActionFilter,
  UserActionFilterUserOption,
  UserActionPaginated,
  UserActionView,
} from "@entities/user-action/types"
import { getPaginatedUserActions, getUserActionFilterUsers } from "@entities/user-action/api"

const ACTION_TYPES = ["create", "edit", "delete", "login", "import", "confirm", "report", "correct"]

const STATUS_OPTIONS: Array<{ label: string; value: "" | "true" | "false" }> = [
  { label: "Все", value: "" },
  { label: "Успешно", value: "true" },
  { label: "Ошибка", value: "false" },
]

interface FilterFormState {
  userQuery: string
  userID: number
  actionType: string
  status: "" | "true" | "false"
  dateFrom: string
  dateTo: string
}

const EMPTY_FILTER: FilterFormState = {
  userQuery: "",
  userID: 0,
  actionType: "",
  status: "",
  dateFrom: "",
  dateTo: "",
}

function toFilter(form: FilterFormState): UserActionFilter {
  const out: UserActionFilter = {}
  if (form.userID > 0) out.userID = form.userID
  if (form.actionType) out.actionType = form.actionType
  if (form.status === "true") out.status = true
  if (form.status === "false") out.status = false
  if (form.dateFrom) out.dateFrom = form.dateFrom
  if (form.dateTo) out.dateTo = form.dateTo
  return out
}

// formatUserOption is the visible label shown in the dropdown. We fold name
// and login into one searchable string so a native <datalist> can match
// either as the admin types.
function formatUserOption(opt: UserActionFilterUserOption): string {
  if (opt.workerName && opt.username) return `${opt.workerName} (${opt.username})`
  return opt.workerName || opt.username || `#${opt.id}`
}

// resolveUserID maps the free-text typed in the search box back to a
// concrete user id. It's permissive: it accepts the formatted label, the
// bare username, or the bare worker name (case-insensitive).
function resolveUserID(query: string, options: UserActionFilterUserOption[]): number {
  const trimmed = query.trim().toLowerCase()
  if (!trimmed) return 0
  for (const opt of options) {
    if (formatUserOption(opt).toLowerCase() === trimmed) return opt.id
    if (opt.username.toLowerCase() === trimmed) return opt.id
    if (opt.workerName.toLowerCase() === trimmed) return opt.id
  }
  return 0
}

function formatTimestamp(iso: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString("ru-RU", { hour12: false })
}

export default function AdminUserActionsPage() {
  const [pendingForm, setPendingForm] = useState<FilterFormState>(EMPTY_FILTER)
  const [appliedForm, setAppliedForm] = useState<FilterFormState>(EMPTY_FILTER)

  const userOptionsQuery = useQuery<UserActionFilterUserOption[], Error>({
    queryKey: ["user-action-filter-users"],
    queryFn: getUserActionFilterUsers,
    staleTime: 5 * 60 * 1000,
  })
  const userOptions = userOptionsQuery.data ?? []

  const filter = useMemo(() => toFilter(appliedForm), [appliedForm])

  const tableDataQuery = useInfiniteQuery<UserActionPaginated, Error>({
    queryKey: ["user-actions", filter],
    queryFn: ({ pageParam = 1 }) => getPaginatedUserActions(pageParam, filter),
    getNextPageParam: (lastPage) => {
      if (lastPage.page * ENTRY_LIMIT >= lastPage.count) return undefined
      return lastPage.page + 1
    },
  })

  const [tableData, setTableData] = useState<UserActionView[]>([])
  useEffect(() => {
    if (tableDataQuery.isSuccess && tableDataQuery.data) {
      const flat = tableDataQuery.data.pages.reduce<UserActionView[]>(
        (acc, page) => [...acc, ...page.data],
        [],
      )
      setTableData(flat)
    }
  }, [tableDataQuery.data, tableDataQuery.isSuccess])

  useScrollPaginated(tableDataQuery.fetchNextPage)

  const totalCount = tableDataQuery.data?.pages[0]?.count ?? 0

  const onApply = (e: React.FormEvent) => {
    e.preventDefault()
    // Resolve the typed name/login to an id at submit time so the user can
    // type freely without firing a query on every keystroke.
    const resolved = resolveUserID(pendingForm.userQuery, userOptions)
    const next: FilterFormState = { ...pendingForm, userID: resolved }
    setPendingForm(next)
    setAppliedForm(next)
  }

  const onReset = () => {
    setPendingForm(EMPTY_FILTER)
    setAppliedForm(EMPTY_FILTER)
  }

  return (
    <main>
      <div className="mt-2 px-2 flex justify-between items-center">
        <span className="text-3xl font-bold">Журнал действий пользователей</span>
        <span className="text-sm text-gray-600">Всего записей: {totalCount}</span>
      </div>

      <form onSubmit={onApply} className="px-2 mt-3 grid grid-cols-6 gap-3 items-end">
        <label className="flex flex-col text-sm">
          <span>Пользователь (имя или логин)</span>
          <input
            type="search"
            list="user-action-filter-users"
            value={pendingForm.userQuery}
            onChange={(e) => setPendingForm((p) => ({ ...p, userQuery: e.target.value, userID: 0 }))}
            placeholder={userOptionsQuery.isLoading ? "Загрузка..." : "Начните вводить..."}
            className="border rounded px-2 py-1"
            autoComplete="off"
          />
          <datalist id="user-action-filter-users">
            {userOptions.map((opt) => (
              <option key={opt.id} value={formatUserOption(opt)} />
            ))}
          </datalist>
        </label>
        <label className="flex flex-col text-sm">
          <span>Тип действия</span>
          <select
            value={pendingForm.actionType}
            onChange={(e) => setPendingForm((p) => ({ ...p, actionType: e.target.value }))}
            className="border rounded px-2 py-1"
          >
            <option value="">Все</option>
            {ACTION_TYPES.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
        </label>
        <label className="flex flex-col text-sm">
          <span>Статус</span>
          <select
            value={pendingForm.status}
            onChange={(e) => setPendingForm((p) => ({ ...p, status: e.target.value as FilterFormState["status"] }))}
            className="border rounded px-2 py-1"
          >
            {STATUS_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>{o.label}</option>
            ))}
          </select>
        </label>
        <label className="flex flex-col text-sm">
          <span>С даты</span>
          <input
            type="date"
            value={pendingForm.dateFrom}
            onChange={(e) => setPendingForm((p) => ({ ...p, dateFrom: e.target.value }))}
            className="border rounded px-2 py-1"
          />
        </label>
        <label className="flex flex-col text-sm">
          <span>По дату</span>
          <input
            type="date"
            value={pendingForm.dateTo}
            onChange={(e) => setPendingForm((p) => ({ ...p, dateTo: e.target.value }))}
            className="border rounded px-2 py-1"
          />
        </label>
        <div className="flex gap-2">
          <button
            type="submit"
            className="bg-gray-800 text-white px-3 py-1 rounded hover:bg-gray-900"
          >
            Применить
          </button>
          <button
            type="button"
            onClick={onReset}
            className="border border-gray-400 px-3 py-1 rounded hover:bg-gray-100"
          >
            Сброс
          </button>
        </div>
      </form>

      <table className="table-auto text-sm text-left mt-3 w-full border-box">
        <thead className="shadow-md border-t-2">
          <tr>
            <th className="px-3 py-2">Дата и время</th>
            <th className="px-3 py-2">Пользователь</th>
            <th className="px-3 py-2">Действие</th>
            <th className="px-3 py-2">URL</th>
            <th className="px-3 py-2">ID объекта</th>
            <th className="px-3 py-2">Статус</th>
            <th className="px-3 py-2">Сообщение</th>
          </tr>
        </thead>
        <tbody>
          {tableData.map((row) => (
            <tr key={row.id} className="border-b">
              <td className="px-3 py-2 whitespace-nowrap">{formatTimestamp(row.dateOfAction)}</td>
              <td className="px-3 py-2">
                {row.username || `#${row.userID}`}
              </td>
              <td className="px-3 py-2">
                <span className="font-mono text-xs px-1 py-0.5 bg-gray-100 rounded mr-1">{row.httpMethod}</span>
                {row.actionType}
              </td>
              <td className="px-3 py-2 font-mono text-xs">{row.actionURL}</td>
              <td className="px-3 py-2">{row.actionID || ""}</td>
              <td className="px-3 py-2">
                {row.actionStatus ? (
                  <span className="text-green-700">Успех</span>
                ) : (
                  <span className="text-red-700">Ошибка</span>
                )}
              </td>
              <td className="px-3 py-2 max-w-md truncate" title={row.actionStatusMessage}>
                {row.actionStatusMessage}
              </td>
            </tr>
          ))}
          {tableData.length === 0 && tableDataQuery.isSuccess && (
            <tr>
              <td colSpan={7} className="px-3 py-6 text-center text-gray-500">
                Записей не найдено
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </main>
  )
}
