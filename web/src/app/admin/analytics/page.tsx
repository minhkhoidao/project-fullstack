'use client';

import { useEffect, useState } from 'react';
import { Loader2, TrendingUp, Award } from 'lucide-react';
import {
  getRevenue,
  getTopProducts,
  type RevenueByDay,
  type TopProduct,
} from '@/lib/api';

const RANGES = [
  { label: '7 days', value: 7 },
  { label: '30 days', value: 30 },
  { label: '90 days', value: 90 },
] as const;

function currency(n: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(n);
}

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  });
}

export default function AnalyticsPage() {
  const [days, setDays] = useState<number>(30);
  const [revenue, setRevenue] = useState<RevenueByDay[]>([]);
  const [topProducts, setTopProducts] = useState<TopProduct[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([getRevenue(days), getTopProducts(10)])
      .then(([r, p]) => {
        setRevenue(r);
        setTopProducts(p);
      })
      .finally(() => setLoading(false));
  }, [days]);

  const maxRevenue = Math.max(...revenue.map((r) => r.revenue), 1);
  const totalRevenue = revenue.reduce((sum, r) => sum + r.revenue, 0);
  const totalOrders = revenue.reduce((sum, r) => sum + r.orderCount, 0);
  const avgDaily = revenue.length > 0 ? totalRevenue / revenue.length : 0;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Analytics</h2>

        {/* Date range selector */}
        <div className="flex rounded-lg border border-gray-300 bg-white shadow-sm">
          {RANGES.map(({ label, value }) => (
            <button
              key={value}
              onClick={() => setDays(value)}
              className={`px-4 py-2 text-sm font-medium transition-colors first:rounded-l-lg last:rounded-r-lg ${
                days === value
                  ? 'bg-gray-900 text-white'
                  : 'text-gray-600 hover:bg-gray-50'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="flex h-64 items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-gray-400" />
        </div>
      ) : (
        <>
          {/* Summary cards */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-3">
            <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
              <p className="text-sm text-gray-500">Period Revenue</p>
              <p className="mt-1 text-2xl font-semibold text-gray-900">
                {currency(totalRevenue)}
              </p>
            </div>
            <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
              <p className="text-sm text-gray-500">Total Orders</p>
              <p className="mt-1 text-2xl font-semibold text-gray-900">
                {totalOrders.toLocaleString()}
              </p>
            </div>
            <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
              <p className="text-sm text-gray-500">Avg Daily Revenue</p>
              <p className="mt-1 text-2xl font-semibold text-gray-900">
                {currency(avgDaily)}
              </p>
            </div>
          </div>

          {/* Revenue by day */}
          <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
            <div className="flex items-center gap-2 border-b border-gray-200 px-6 py-4">
              <TrendingUp className="h-5 w-5 text-gray-400" />
              <h3 className="text-lg font-semibold text-gray-900">
                Revenue by Day
              </h3>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-gray-100 text-xs uppercase text-gray-500">
                    <th className="px-6 py-3 font-medium">Date</th>
                    <th className="px-6 py-3 font-medium">Revenue</th>
                    <th className="px-6 py-3 font-medium text-right">Orders</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {revenue.map((day) => (
                    <tr
                      key={day.date}
                      className="transition-colors hover:bg-gray-50"
                    >
                      <td className="px-6 py-3 text-gray-700">
                        {fmtDate(day.date)}
                      </td>
                      <td className="px-6 py-3">
                        <div className="flex items-center gap-3">
                          <div className="h-4 flex-1 overflow-hidden rounded-full bg-gray-100">
                            <div
                              className="h-full rounded-full bg-brand-400 transition-all"
                              style={{
                                width: `${(day.revenue / maxRevenue) * 100}%`,
                              }}
                            />
                          </div>
                          <span className="w-20 text-right font-medium text-gray-900">
                            {currency(day.revenue)}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-3 text-right text-gray-600">
                        {day.orderCount}
                      </td>
                    </tr>
                  ))}
                  {revenue.length === 0 && (
                    <tr>
                      <td
                        colSpan={3}
                        className="px-6 py-8 text-center text-gray-400"
                      >
                        No revenue data for this period.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Top products */}
          <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
            <div className="flex items-center gap-2 border-b border-gray-200 px-6 py-4">
              <Award className="h-5 w-5 text-gray-400" />
              <h3 className="text-lg font-semibold text-gray-900">
                Top Products
              </h3>
            </div>
            <div className="divide-y divide-gray-100">
              {topProducts.map((product, i) => {
                const maxProductRevenue = topProducts[0]?.revenue ?? 1;
                return (
                  <div
                    key={product.id}
                    className="flex items-center gap-4 px-6 py-4 transition-colors hover:bg-gray-50"
                  >
                    <span
                      className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-bold ${
                        i === 0
                          ? 'bg-amber-100 text-amber-700'
                          : i === 1
                            ? 'bg-gray-200 text-gray-600'
                            : i === 2
                              ? 'bg-orange-100 text-orange-700'
                              : 'bg-gray-100 text-gray-500'
                      }`}
                    >
                      {i + 1}
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-medium text-gray-900">
                        {product.name}
                      </p>
                      <div className="mt-1 flex items-center gap-3">
                        <div className="h-2 flex-1 overflow-hidden rounded-full bg-gray-100">
                          <div
                            className="h-full rounded-full bg-brand-400 transition-all"
                            style={{
                              width: `${(product.revenue / maxProductRevenue) * 100}%`,
                            }}
                          />
                        </div>
                      </div>
                    </div>
                    <div className="shrink-0 text-right">
                      <p className="text-sm font-semibold text-gray-900">
                        {currency(product.revenue)}
                      </p>
                      <p className="text-xs text-gray-500">
                        {product.unitsSold} sold
                      </p>
                    </div>
                  </div>
                );
              })}
              {topProducts.length === 0 && (
                <div className="px-6 py-8 text-center text-gray-400">
                  No product data available.
                </div>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
