'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Package, ChevronDown, ChevronUp } from 'lucide-react';
import { api, type Order } from '@/lib/api';
import { useAuthStore } from '@/stores/auth';

const STATUS_STYLES: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-800',
  paid: 'bg-blue-100 text-blue-800',
  shipped: 'bg-purple-100 text-purple-800',
  delivered: 'bg-green-100 text-green-800',
  cancelled: 'bg-red-100 text-red-800',
};

export default function OrdersPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, isLoading: authLoading } = useAuthStore();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(true);
  const [expandedOrder, setExpandedOrder] = useState<string | null>(null);

  const successOrderId = searchParams.get('success');

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/login?redirect=/account/orders');
      return;
    }
    if (isAuthenticated) {
      api.orders
        .list()
        .then(setOrders)
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [authLoading, isAuthenticated, router]);

  if (authLoading || loading) {
    return (
      <div className="pt-16">
        <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 py-12 animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/3" />
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-20 bg-gray-200 rounded" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="pt-16">
      <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 py-12">
        <div className="flex items-center justify-between mb-10">
          <h1 className="text-2xl font-light tracking-tight font-display">Order History</h1>
          <Link
            href="/account"
            className="text-xs text-gray-500 hover:text-black underline"
          >
            Back to Account
          </Link>
        </div>

        {successOrderId && (
          <div className="mb-8 rounded-md bg-green-50 border border-green-200 px-4 py-3">
            <p className="text-sm text-green-800">
              Order placed successfully! Your order is being processed.
            </p>
          </div>
        )}

        {orders.length === 0 ? (
          <div className="text-center py-16">
            <Package className="h-12 w-12 text-gray-300 mx-auto mb-4" strokeWidth={1} />
            <p className="text-gray-500 mb-4">No orders yet</p>
            <Link
              href="/products"
              className="inline-block px-8 py-3 bg-black text-white text-xs uppercase tracking-[0.2em] font-medium hover:bg-gray-900 transition-colors"
            >
              Start Shopping
            </Link>
          </div>
        ) : (
          <div className="space-y-4">
            {orders.map((order) => {
              const isExpanded = expandedOrder === order.id;
              const itemCount = order.items.reduce((sum, item) => sum + item.quantity, 0);

              return (
                <div
                  key={order.id}
                  className="border border-gray-200 rounded-lg overflow-hidden"
                >
                  <button
                    onClick={() => setExpandedOrder(isExpanded ? null : order.id)}
                    className="w-full flex items-center justify-between px-5 py-4 hover:bg-gray-50 transition-colors text-left"
                  >
                    <div className="flex items-center gap-4 flex-wrap">
                      <div>
                        <p className="text-sm font-medium">#{order.orderNumber}</p>
                        <p className="text-xs text-gray-500">
                          {new Date(order.createdAt).toLocaleDateString('en-US', {
                            year: 'numeric',
                            month: 'short',
                            day: 'numeric',
                          })}
                        </p>
                      </div>
                      <span
                        className={`inline-block px-2.5 py-0.5 text-xs font-medium rounded-full capitalize ${
                          STATUS_STYLES[order.status] ?? 'bg-gray-100 text-gray-800'
                        }`}
                      >
                        {order.status}
                      </span>
                    </div>
                    <div className="flex items-center gap-4">
                      <div className="text-right">
                        <p className="text-sm font-medium">${order.total.toFixed(2)}</p>
                        <p className="text-xs text-gray-500">
                          {itemCount} item{itemCount !== 1 ? 's' : ''}
                        </p>
                      </div>
                      {isExpanded ? (
                        <ChevronUp className="h-4 w-4 text-gray-400" />
                      ) : (
                        <ChevronDown className="h-4 w-4 text-gray-400" />
                      )}
                    </div>
                  </button>

                  {isExpanded && (
                    <div className="border-t border-gray-100 px-5 py-4 bg-gray-50/50 space-y-4">
                      {/* Items */}
                      <div className="space-y-3">
                        {order.items.map((item) => (
                          <div key={item.id} className="flex justify-between text-sm">
                            <div>
                              <p className="font-medium">{item.productName}</p>
                              <p className="text-xs text-gray-500">
                                {item.variantInfo} &times; {item.quantity}
                              </p>
                            </div>
                            <p>${(item.unitPrice * item.quantity).toFixed(2)}</p>
                          </div>
                        ))}
                      </div>

                      {/* Totals */}
                      <div className="border-t border-gray-200 pt-3 space-y-1 text-sm">
                        <div className="flex justify-between text-gray-600">
                          <span>Subtotal</span>
                          <span>${order.subtotal.toFixed(2)}</span>
                        </div>
                        <div className="flex justify-between text-gray-600">
                          <span>Shipping</span>
                          <span>
                            {order.shipping === 0 ? 'Free' : `$${order.shipping.toFixed(2)}`}
                          </span>
                        </div>
                        <div className="flex justify-between font-medium pt-1">
                          <span>Total</span>
                          <span>${order.total.toFixed(2)}</span>
                        </div>
                      </div>

                      {/* Shipping address */}
                      <div className="border-t border-gray-200 pt-3">
                        <p className="text-xs font-medium uppercase tracking-wider text-gray-500 mb-1">
                          Ships to
                        </p>
                        <p className="text-sm text-gray-700">
                          {order.shippingAddress.street}, {order.shippingAddress.city},{' '}
                          {order.shippingAddress.state} {order.shippingAddress.zip}
                        </p>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
