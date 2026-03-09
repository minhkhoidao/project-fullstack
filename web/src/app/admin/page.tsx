'use client';

import { useEffect, useState } from 'react';
import { Users, Package, ShoppingCart, DollarSign, Loader2 } from 'lucide-react';
import {
  getDashboard,
  getRevenue,
  getTopProducts,
  type DashboardStats,
  type RevenueByDay,
  type TopProduct,
} from '@/lib/api';

function currency(n: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(n);
}

function shortDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  });
}

const STATUS_COLORS: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800',
  paid: 'bg-blue-100 text-blue-800',
  shipped: 'bg-purple-100 text-purple-800',
  delivered: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
};

export default function AdminDashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [revenue, setRevenue] = useState<RevenueByDay[]>([]);
  const [topProducts, setTopProducts] = useState<TopProduct[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getDashboard(), getRevenue(30), getTopProducts(5)])
      .then(([s, r, p]) => {
        setStats(s);
        setRevenue(r);
        setTopProducts(p);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
      </div>
    );
  }

  if (!stats) return null;

  const statCards = [
    { label: 'Total Users', value: stats.totalUsers.toLocaleString(), icon: Users, color: 'text-blue-600 bg-blue-50' },
    { label: 'Total Products', value: stats.totalProducts.toLocaleString(), icon: Package, color: 'text-purple-600 bg-purple-50' },
    { label: 'Total Orders', value: stats.totalOrders.toLocaleString(), icon: ShoppingCart, color: 'text-amber-600 bg-amber-50' },
    { label: 'Total Revenue', value: currency(stats.totalRevenue), icon: DollarSign, color: 'text-green-600 bg-green-50' },
  ];

  const maxRevenue = Math.max(...revenue.map((r) => r.revenue), 1);

  return (
    <div className="space-y-8">
      <h2 className="text-2xl font-bold text-gray-900">Dashboard</h2>

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 xl:grid-cols-4">
        {statCards.map(({ label, value, icon: Icon, color }) => (
          <div
            key={label}
            className="flex items-center gap-4 rounded-xl border border-gray-200 bg-white p-6 shadow-sm"
          >
            <div className={`flex h-12 w-12 items-center justify-center rounded-lg ${color}`}>
              <Icon className="h-6 w-6" />
            </div>
            <div>
              <p className="text-sm text-gray-500">{label}</p>
              <p className="text-2xl font-semibold text-gray-900">{value}</p>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-6 xl:grid-cols-3">
        {/* Revenue chart (bar representation) */}
        <div className="col-span-2 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-semibold text-gray-900">
            Revenue — Last 30 Days
          </h3>
          {revenue.length === 0 ? (
            <p className="text-sm text-gray-500">No revenue data available.</p>
          ) : (
            <div className="space-y-2">
              {revenue.slice(-14).map((day) => (
                <div key={day.date} className="flex items-center gap-3 text-sm">
                  <span className="w-16 shrink-0 text-gray-500">
                    {shortDate(day.date)}
                  </span>
                  <div className="flex-1">
                    <div
                      className="h-5 rounded bg-brand-400 transition-all"
                      style={{
                        width: `${(day.revenue / maxRevenue) * 100}%`,
                        minWidth: day.revenue > 0 ? '4px' : '0px',
                      }}
                    />
                  </div>
                  <span className="w-20 text-right font-medium text-gray-700">
                    {currency(day.revenue)}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Top selling products */}
        <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h3 className="mb-4 text-lg font-semibold text-gray-900">
            Top Selling Products
          </h3>
          <ul className="space-y-4">
            {topProducts.map((product, i) => (
              <li key={product.id} className="flex items-center gap-3">
                <span className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100 text-sm font-semibold text-gray-600">
                  {i + 1}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="truncate text-sm font-medium text-gray-900">
                    {product.name}
                  </p>
                  <p className="text-xs text-gray-500">
                    {product.unitsSold} sold · {currency(product.revenue)}
                  </p>
                </div>
              </li>
            ))}
            {topProducts.length === 0 && (
              <p className="text-sm text-gray-500">No product data.</p>
            )}
          </ul>
        </div>
      </div>

      {/* Recent orders */}
      <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
        <div className="border-b border-gray-200 px-6 py-4">
          <h3 className="text-lg font-semibold text-gray-900">Recent Orders</h3>
        </div>
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
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {stats.recentOrders.slice(0, 5).map((order) => (
                <tr key={order.id} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-3 font-mono text-xs text-gray-600">
                    {order.id.slice(0, 8)}…
                  </td>
                  <td className="px-6 py-3 text-gray-900">{order.customerEmail}</td>
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
                    {shortDate(order.createdAt)}
                  </td>
                </tr>
              ))}
              {stats.recentOrders.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-8 text-center text-gray-400">
                    No orders yet.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
