'use client';

import { useEffect, useState, useCallback } from 'react';
import { Loader2, ChevronDown } from 'lucide-react';
import {
  listAllOrders,
  updateOrderStatus,
  type OrderSummary,
  type OrderStatus,
} from '@/lib/api';

const ALL_STATUSES: (OrderStatus | 'all')[] = [
  'all',
  'pending',
  'paid',
  'shipped',
  'delivered',
  'cancelled',
];

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800',
  paid: 'bg-blue-100 text-blue-800',
  shipped: 'bg-purple-100 text-purple-800',
  delivered: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
};

const NEXT_STATUSES: Record<OrderStatus, OrderStatus[]> = {
  pending: ['paid', 'cancelled'],
  paid: ['shipped', 'cancelled'],
  shipped: ['delivered', 'cancelled'],
  delivered: [],
  cancelled: [],
};

function currency(n: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(n);
}

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

export default function OrdersPage() {
  const [orders, setOrders] = useState<OrderSummary[]>([]);
  const [cursor, setCursor] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<OrderStatus | 'all'>('all');
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [updatingId, setUpdatingId] = useState<string | null>(null);

  const fetchOrders = useCallback(
    async (reset = false) => {
      const setter = reset ? setLoading : setLoadingMore;
      setter(true);
      try {
        const res = await listAllOrders({
          status: statusFilter,
          cursor: reset ? undefined : cursor ?? undefined,
          limit: 20,
        });
        setOrders((prev) => (reset ? res.orders : [...prev, ...res.orders]));
        setCursor(res.nextCursor);
      } finally {
        setter(false);
      }
    },
    [statusFilter, cursor],
  );

  useEffect(() => {
    setCursor(null);
    setOrders([]);
    setLoading(true);
    listAllOrders({ status: statusFilter, limit: 20 })
      .then((res) => {
        setOrders(res.orders);
        setCursor(res.nextCursor);
      })
      .finally(() => setLoading(false));
  }, [statusFilter]);

  async function handleStatusChange(id: string, newStatus: OrderStatus) {
    setUpdatingId(id);
    try {
      await updateOrderStatus(id, newStatus);
      setOrders((prev) =>
        prev.map((o) => (o.id === id ? { ...o, status: newStatus } : o)),
      );
    } finally {
      setUpdatingId(null);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Orders</h2>
      </div>

      {/* Filter bar */}
      <div className="flex items-center gap-3">
        <label htmlFor="status-filter" className="text-sm font-medium text-gray-700">Status:</label>
        <select
          id="status-filter"
          value={statusFilter}
          onChange={(e) =>
            setStatusFilter(e.target.value as OrderStatus | 'all')
          }
          className="rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
        >
          {ALL_STATUSES.map((s) => (
            <option key={s} value={s}>
              {s === 'all' ? 'All Statuses' : s.charAt(0).toUpperCase() + s.slice(1)}
            </option>
          ))}
        </select>
      </div>

      {/* Table */}
      <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead>
              <tr className="border-b border-gray-100 text-xs uppercase text-gray-500">
                <th className="px-6 py-3 font-medium">Order ID</th>
                <th className="px-6 py-3 font-medium">Customer</th>
                <th className="px-6 py-3 font-medium">Status</th>
                <th className="px-6 py-3 font-medium text-right">Total</th>
                <th className="px-6 py-3 font-medium text-right">Items</th>
                <th className="px-6 py-3 font-medium">Date</th>
                <th className="px-6 py-3 font-medium">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {loading ? (
                <tr>
                  <td colSpan={7} className="px-6 py-12 text-center">
                    <Loader2 className="mx-auto h-6 w-6 animate-spin text-gray-400" />
                  </td>
                </tr>
              ) : orders.length === 0 ? (
                <tr>
                  <td
                    colSpan={7}
                    className="px-6 py-12 text-center text-gray-400"
                  >
                    No orders found.
                  </td>
                </tr>
              ) : (
                orders.map((order) => {
                  const nextStatuses = NEXT_STATUSES[order.status];
                  return (
                    <tr
                      key={order.id}
                      className="transition-colors hover:bg-gray-50"
                    >
                      <td className="px-6 py-3 font-mono text-xs text-gray-600">
                        {order.id.slice(0, 8)}…
                      </td>
                      <td className="px-6 py-3 text-gray-900">
                        {order.customerEmail}
                      </td>
                      <td className="px-6 py-3">
                        <span
                          className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium capitalize ${STATUS_COLORS[order.status] ?? 'bg-gray-100 text-gray-800'}`}
                        >
                          {order.status}
                        </span>
                      </td>
                      <td className="px-6 py-3 text-right font-medium text-gray-900">
                        {currency(order.total)}
                      </td>
                      <td className="px-6 py-3 text-right text-gray-600">
                        {order.itemCount}
                      </td>
                      <td className="px-6 py-3 text-gray-500">
                        {fmtDate(order.createdAt)}
                      </td>
                      <td className="px-6 py-3">
                        {nextStatuses.length > 0 ? (
                          <div className="relative inline-block">
                            <select
                              disabled={updatingId === order.id}
                              value=""
                              onChange={(e) => {
                                if (e.target.value) {
                                  handleStatusChange(
                                    order.id,
                                    e.target.value as OrderStatus,
                                  );
                                }
                              }}
                              className="rounded-md border border-gray-300 bg-white py-1 pl-2 pr-7 text-xs shadow-sm focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400 disabled:opacity-50"
                            >
                              <option value="">Change…</option>
                              {nextStatuses.map((s) => (
                                <option key={s} value={s}>
                                  {s.charAt(0).toUpperCase() + s.slice(1)}
                                </option>
                              ))}
                            </select>
                            {updatingId === order.id && (
                              <Loader2 className="absolute -right-6 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin text-gray-400" />
                            )}
                          </div>
                        ) : (
                          <span className="text-xs text-gray-400">—</span>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        {/* Load more */}
        {cursor && !loading && (
          <div className="border-t border-gray-100 px-6 py-4 text-center">
            <button
              onClick={() => fetchOrders(false)}
              disabled={loadingMore}
              className="inline-flex items-center gap-2 rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm transition-colors hover:bg-gray-50 disabled:opacity-50"
            >
              {loadingMore ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
              Load More
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
