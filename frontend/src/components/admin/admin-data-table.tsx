import { Loader2 } from "lucide-react";
import { SurfacePanel } from "../surface-panel";

export type Column<T> = {
  header: string;
  cell: (row: T) => React.ReactNode;
  headerClassName?: string;
  cellClassName?: string;
};

type DataTableProps<T> = {
  data: T[];
  columns: Column<T>[];
  getRowKey: (row: T) => string;
  isLoading: boolean;
  isFetching: boolean;
  loadingLabel: string;
  emptyLabel: string;
  onRowClick?: (row: T) => void;
};

export function DataTable<T>({
  data,
  columns,
  getRowKey,
  isLoading,
  isFetching,
  loadingLabel,
  emptyLabel,
  onRowClick,
}: DataTableProps<T>) {
  if (isLoading) {
    return (
      <SurfacePanel className="flex min-h-112.5 flex-1 flex-col">
        <div className="flex flex-1 items-center justify-center gap-3">
          <Loader2 aria-hidden="true" className="size-5 animate-spin" />
          <span>{loadingLabel}</span>
        </div>
      </SurfacePanel>
    );
  }

  if (data.length === 0) {
    return (
      <SurfacePanel className="flex min-h-112.5 flex-1 flex-col">
        <div className="flex flex-1 items-center justify-center px-6 py-12 text-brand-text-muted">
          <span>{emptyLabel}</span>
        </div>
      </SurfacePanel>
    );
  }

  return (
    <SurfacePanel className="flex min-h-112.5 flex-1 flex-col overflow-hidden">
      <div
        className={`min-h-112.5 flex-1 overflow-auto ${isFetching ? "opacity-70 transition-opacity" : ""}`}
      >
        <table className="min-w-full border-collapse text-left text-sm">
          <thead>
            <tr className="border-b border-brand-border bg-white/2">
              {columns.map((column) => (
                <th
                  key={column.header}
                  scope="col"
                  className="whitespace-nowrap px-4 py-3 text-xs font-semibold uppercase tracking-wide text-brand-text-subtle"
                >
                  {column.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.map((row) => (
              <tr
                key={getRowKey(row)}
                onClick={
                  onRowClick
                    ? (event) => {
                        const target = event.target;
                        if (
                          target instanceof Element &&
                          target.closest("a, button, input, select, textarea")
                        ) {
                          return;
                        }

                        onRowClick(row);
                      }
                    : undefined
                }
                className={`border-b border-brand-border/70 last:border-b-0 ${
                  onRowClick
                    ? "cursor-pointer transition-colors hover:bg-white/4s-within:bg-white/[0.04]"
                    : ""
                }`}
              >
                {columns.map((column) => (
                  <td
                    key={column.header}
                    className={
                      column.cellClassName ?? "whitespace-nowrap px-4 py-3 text-brand-text-muted"
                    }
                  >
                    {column.cell(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </SurfacePanel>
  );
}
