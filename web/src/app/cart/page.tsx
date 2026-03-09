'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { Minus, Plus, Trash2, ShoppingBag } from 'lucide-react';
import { api, type Cart } from '@/lib/api';

export default function CartPage() {
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [updatingItem, setUpdatingItem] = useState<string | null>(null);

  useEffect(() => {
    api.cart
      .get()
      .then(setCart)
      .catch(() => setCart(null))
      .finally(() => setLoading(false));
  }, []);

  async function updateQuantity(itemId: string, quantity: number) {
    setUpdatingItem(itemId);
    try {
      const updated = await api.cart.updateItem(itemId, quantity);
      setCart(updated);
    } catch {
      /* ignore */
    } finally {
      setUpdatingItem(null);
    }
  }

  async function removeItem(itemId: string) {
    setUpdatingItem(itemId);
    try {
      const updated = await api.cart.removeItem(itemId);
      setCart(updated);
    } catch {
      /* ignore */
    } finally {
      setUpdatingItem(null);
    }
  }

  const subtotal = cart?.subtotal ?? 0;
  const shipping = subtotal >= 100 ? 0 : 9.99;
  const total = subtotal + shipping;

  if (loading) {
    return (
      <div className="pt-16">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
          <h1 className="text-2xl font-light tracking-tight font-display mb-8">Shopping Cart</h1>
          <div className="animate-pulse space-y-6">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="flex gap-4">
                <div className="w-24 h-32 bg-gray-200 rounded" />
                <div className="flex-1 space-y-3">
                  <div className="h-4 bg-gray-200 rounded w-1/2" />
                  <div className="h-3 bg-gray-200 rounded w-1/4" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (!cart || cart.items.length === 0) {
    return (
      <div className="pt-16 flex flex-col items-center justify-center min-h-[60vh] px-4">
        <ShoppingBag className="h-16 w-16 text-gray-300 mb-6" strokeWidth={1} />
        <h1 className="text-2xl font-light tracking-tight font-display mb-2">
          Your cart is empty
        </h1>
        <p className="text-sm text-gray-500 mb-8">Looks like you haven&apos;t added anything yet</p>
        <Link
          href="/products"
          className="px-8 py-3 bg-black text-white text-xs uppercase tracking-[0.2em] font-medium hover:bg-gray-900 transition-colors"
        >
          Continue Shopping
        </Link>
      </div>
    );
  }

  return (
    <div className="pt-16">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
        <h1 className="text-2xl font-light tracking-tight font-display mb-8">Shopping Cart</h1>

        <div className="lg:grid lg:grid-cols-3 lg:gap-12">
          {/* Cart Items */}
          <div className="lg:col-span-2 divide-y divide-gray-100">
            {cart.items.map((item) => {
              const image = item.product.images?.[0];
              const lineTotal = item.variant.price * item.quantity;
              return (
                <div
                  key={item.id}
                  className={`flex gap-4 sm:gap-6 py-6 first:pt-0 transition-opacity ${
                    updatingItem === item.id ? 'opacity-50' : ''
                  }`}
                >
                  {/* Image */}
                  <div className="relative w-24 sm:w-28 aspect-[3/4] bg-gray-100 rounded-md overflow-hidden flex-shrink-0">
                    {image ? (
                      <Image
                        src={image.url}
                        alt={image.alt || item.product.name}
                        fill
                        className="object-cover"
                      />
                    ) : (
                      <div className="absolute inset-0 bg-gradient-to-br from-gray-200 to-gray-300" />
                    )}
                  </div>

                  {/* Details */}
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between gap-4">
                      <div>
                        <Link
                          href={`/products/${item.product.slug}`}
                          className="text-sm font-medium hover:underline"
                        >
                          {item.product.name}
                        </Link>
                        <p className="mt-0.5 text-xs text-gray-500">
                          {item.variant.size} / {item.variant.color}
                        </p>
                      </div>
                      <p className="text-sm font-medium whitespace-nowrap">
                        ${lineTotal.toFixed(2)}
                      </p>
                    </div>

                    <div className="mt-4 flex items-center justify-between">
                      <div className="flex items-center border border-gray-300 rounded">
                        <button
                          onClick={() => updateQuantity(item.id, Math.max(1, item.quantity - 1))}
                          disabled={updatingItem === item.id}
                          className="px-2.5 py-1.5 text-gray-500 hover:text-black transition-colors"
                          aria-label="Decrease"
                        >
                          <Minus className="h-3 w-3" />
                        </button>
                        <span className="w-8 text-center text-xs tabular-nums">
                          {item.quantity}
                        </span>
                        <button
                          onClick={() => updateQuantity(item.id, item.quantity + 1)}
                          disabled={updatingItem === item.id}
                          className="px-2.5 py-1.5 text-gray-500 hover:text-black transition-colors"
                          aria-label="Increase"
                        >
                          <Plus className="h-3 w-3" />
                        </button>
                      </div>
                      <button
                        onClick={() => removeItem(item.id)}
                        disabled={updatingItem === item.id}
                        className="text-gray-400 hover:text-red-500 transition-colors p-1"
                        aria-label="Remove item"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>

                    <p className="mt-2 text-xs text-gray-500">
                      ${item.variant.price.toFixed(2)} each
                    </p>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Order Summary */}
          <div className="mt-10 lg:mt-0">
            <div className="bg-gray-50 rounded-lg p-6 sticky top-24">
              <h2 className="text-sm font-semibold uppercase tracking-[0.15em] mb-6">
                Order Summary
              </h2>
              <div className="space-y-3 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Subtotal</span>
                  <span>${subtotal.toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Shipping</span>
                  <span>{shipping === 0 ? 'Free' : `$${shipping.toFixed(2)}`}</span>
                </div>
                {subtotal < 100 && (
                  <p className="text-xs text-gray-500">
                    Free shipping on orders over $100
                  </p>
                )}
                <div className="border-t border-gray-200 pt-3 flex justify-between font-medium">
                  <span>Total</span>
                  <span>${total.toFixed(2)}</span>
                </div>
              </div>
              <Link
                href="/checkout"
                className="mt-6 block w-full bg-black text-white text-center py-3 text-xs uppercase tracking-[0.2em] font-medium hover:bg-gray-900 transition-colors"
              >
                Proceed to Checkout
              </Link>
              <Link
                href="/products"
                className="mt-3 block text-center text-xs text-gray-500 hover:text-black transition-colors"
              >
                Continue Shopping
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
