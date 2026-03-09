'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { api, type Cart, type Address } from '@/lib/api';
import { useAuthStore } from '@/stores/auth';

export default function CheckoutPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading: authLoading } = useAuthStore();
  const [cart, setCart] = useState<Cart | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const [address, setAddress] = useState<Omit<Address, 'id'>>({
    street: '',
    city: '',
    state: '',
    zip: '',
    country: 'US',
  });

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/login?redirect=/checkout');
      return;
    }
    api.cart
      .get()
      .then(setCart)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [authLoading, isAuthenticated, router]);

  const subtotal = cart?.subtotal ?? 0;
  const shipping = subtotal >= 100 ? 0 : 9.99;
  const total = subtotal + shipping;

  function handleChange(e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) {
    setAddress((prev) => ({ ...prev, [e.target.name]: e.target.value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      const order = await api.orders.create({ shippingAddress: address });
      router.push(`/account/orders?success=${order.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to place order');
    } finally {
      setSubmitting(false);
    }
  }

  if (authLoading || loading) {
    return (
      <div className="pt-16">
        <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 py-12">
          <div className="animate-pulse space-y-6">
            <div className="h-8 bg-gray-200 rounded w-1/3" />
            <div className="h-40 bg-gray-200 rounded" />
            <div className="h-32 bg-gray-200 rounded" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="pt-16">
      <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 py-12">
        <h1 className="text-2xl font-light tracking-tight font-display mb-10">Checkout</h1>

        <form onSubmit={handleSubmit} className="space-y-10">
          {/* Shipping Address */}
          <section>
            <h2 className="text-sm font-semibold uppercase tracking-[0.15em] mb-6">
              Shipping Address
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="sm:col-span-2">
                <label className="block text-xs text-gray-600 mb-1">Street Address</label>
                <input
                  required
                  type="text"
                  name="street"
                  value={address.street}
                  onChange={handleChange}
                  className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors"
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">City</label>
                <input
                  required
                  type="text"
                  name="city"
                  value={address.city}
                  onChange={handleChange}
                  className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors"
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">State</label>
                <input
                  required
                  type="text"
                  name="state"
                  value={address.state}
                  onChange={handleChange}
                  className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors"
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">ZIP / Postal Code</label>
                <input
                  required
                  type="text"
                  name="zip"
                  value={address.zip}
                  onChange={handleChange}
                  className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors"
                />
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">Country</label>
                <select
                  name="country"
                  value={address.country}
                  onChange={handleChange}
                  className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors bg-white"
                >
                  <option value="US">United States</option>
                  <option value="CA">Canada</option>
                  <option value="GB">United Kingdom</option>
                  <option value="AU">Australia</option>
                  <option value="DE">Germany</option>
                  <option value="FR">France</option>
                </select>
              </div>
            </div>
          </section>

          {/* Order Summary */}
          <section>
            <h2 className="text-sm font-semibold uppercase tracking-[0.15em] mb-6">
              Order Summary
            </h2>
            <div className="bg-gray-50 rounded-lg p-6">
              {cart && cart.items.length > 0 && (
                <div className="space-y-3 mb-4 pb-4 border-b border-gray-200">
                  {cart.items.map((item) => (
                    <div key={item.id} className="flex justify-between text-sm">
                      <span className="text-gray-600">
                        {item.product.name} ({item.variant.size}/{item.variant.color}) &times;{' '}
                        {item.quantity}
                      </span>
                      <span>${(item.variant.price * item.quantity).toFixed(2)}</span>
                    </div>
                  ))}
                </div>
              )}
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Subtotal</span>
                  <span>${subtotal.toFixed(2)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Shipping</span>
                  <span>{shipping === 0 ? 'Free' : `$${shipping.toFixed(2)}`}</span>
                </div>
                <div className="border-t border-gray-200 pt-2 flex justify-between font-medium text-base">
                  <span>Total</span>
                  <span>${total.toFixed(2)}</span>
                </div>
              </div>
            </div>
          </section>

          {error && (
            <div className="rounded-md bg-red-50 border border-red-200 px-4 py-3">
              <p className="text-sm text-red-700">{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={submitting || !cart || cart.items.length === 0}
            className="w-full bg-black text-white py-3.5 text-sm uppercase tracking-[0.2em] font-medium hover:bg-gray-900 transition-colors disabled:opacity-50"
          >
            {submitting ? 'Placing Order...' : 'Place Order'}
          </button>
        </form>
      </div>
    </div>
  );
}
