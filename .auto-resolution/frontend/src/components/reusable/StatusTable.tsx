import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "#components/ui/table"
import { EmptyState } from "./EmptyState"
import { cn } from "#lib/utils"
import type { LucideIcon } from "lucide-react"

export interface StatusTableColumn<T> {
  key: string
  header: string
  className?: string
  cell: (row: T, index: number) => React.ReactNode
  align?: "left" | "right" | "center"
  width?: string
}

interface StatusTableProps<T> {
  columns: StatusTableColumn<T>[]
  rows: T[]
  getRowId: (row: T, index: number) => string
  onRowClick?: (row: T, index: number) => void
  emptyIcon?: LucideIcon
  emptyTitle: string
  emptyDescription?: string
  className?: string
  rowClassName?: string
  getRowClassName?: (row: T, index: number) => string
  density?: "comfortable" | "compact"
}

const ALIGN_CLASS = {
  left: "text-left",
  right: "text-right",
  center: "text-center",
} as const

/**
 * StatusTable — shadcn Table wrapper that pairs with StatusPill / MethodLabel.
 * Use for history, audit, runner runs — anywhere a tabular ledger is needed
 * with a status ramp and tabular-nums code column.
 */
export function StatusTable<T>({
  columns,
  rows,
  getRowId,
  onRowClick,
  emptyIcon,
  emptyTitle,
  emptyDescription,
  className,
  rowClassName,
  getRowClassName,
  density = "comfortable",
}: StatusTableProps<T>) {
  const cellY = density === "compact" ? "py-1.5" : "py-2"

  if (rows.length === 0) {
    return (
      <EmptyState
        icon={emptyIcon}
        title={emptyTitle}
        description={emptyDescription}
        className="border border-dashed border-border/60"
      />
    )
  }

  return (
    <Table className={cn("w-full", className)}>
      <TableHeader>
        <TableRow>
          {columns.map((c) => (
            <TableHead
              key={c.key}
              className={cn(
                "font-mono text-[10px] uppercase tracking-wider text-muted-foreground",
                c.align ? ALIGN_CLASS[c.align] : ALIGN_CLASS.left,
                c.className,
              )}
              style={c.width ? { width: c.width } : undefined}
            >
              {c.header}
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((row, i) => {
          const id = getRowId(row, i)
          const extraCls = getRowClassName ? getRowClassName(row, i) : rowClassName
          return (
            <TableRow
              key={id}
              onClick={onRowClick ? () => onRowClick(row, i) : undefined}
              className={cn(onRowClick && "cursor-pointer", extraCls)}
              tabIndex={onRowClick ? 0 : undefined}
              onKeyDown={
                onRowClick
                  ? (e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault()
                        onRowClick(row, i)
                      }
                    }
                  : undefined
              }
            >
              {columns.map((c) => (
                <TableCell
                  key={c.key}
                  className={cn(
                    "font-mono text-xs",
                    cellY,
                    c.align ? ALIGN_CLASS[c.align] : ALIGN_CLASS.left,
                    c.className,
                  )}
                >
                  {c.cell(row, i)}
                </TableCell>
              ))}
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
