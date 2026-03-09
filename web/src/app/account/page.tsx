'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Trash2, Plus } from 'lucide-react';
import { api, type Address } from '@/lib/api';
import { useAuthStore } from '@/stores/auth';

export default function AccountPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading } = useAuthStore();
  const [addresses, setAddresses] = useState<Address[]>([]);
  const [editing, setEditing] = useState(false);
  const [saving, setSaving] = useState(false);
  const [showAddressForm, setShowAddressForm] = useState(false);

  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [newAddress, setNewAddress] = useState<Omit<Address, 'id'>>({
    street: '',
    city: '',
    state: '',
    zip: '',
    country: 'US',
  });

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login?redirect=/account');
    }
  }, [isLoading, isAuthenticated, router]);

  useEffect(() => {
    if (user) {
      setFirstName(user.firstName);
      setLastName(user.lastName);
    }
  }, [user]);

  useEffect(() => {
    if (isAuthenticated) {
      api.addresses.list().then(setAddresses).catch(() => {});
    }
  }, [isAuthenticated]);

  async function handleSaveProfile(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    try {
      await api.profile.update({ firstName, lastName });
      setEditing(false);
    } catch {
      /* ignore */
    } finally {
      setSaving(false);
    }
  }

  async function handleAddAddress(e: React.FormEvent) {
    e.preventDefault();
    try {
      const created = await api.addresses.create(newAddress);
      setAddresses((prev) => [...prev, created]);
      setNewAddress({ street: '', city: '', state: '', zip: '', country: 'US' });
      setShowAddressForm(false);
    } catch {
      /* ignore */
    }
  }

  async function handleDeleteAddress(id: string) {
    try {
      await api.addresses.delete(id);
      setAddresses((prev) => prev.filter((a) => a.id !== id));
    } catch {
      /* ignore */
    }
  }

  if (isLoading || !user) {
    return (
      <div className="pt-16">
        <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 py-12 animate-pulse space-y-6">
          <div className="h-8 bg-gray-200 rounded w-1/3" />
          <div className="h-32 bg-gray-200 rounded" />
        </div>
      </div>
    );
  }

  return (
    <div className="pt-16">
      <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8 py-12">
        <h1 className="text-2xl font-light tracking-tight font-display mb-10">My Account</h1>

        {/* Profile */}
        <section className="mb-12">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-sm font-semibold uppercase tracking-[0.15em]">Profile</h2>
            {!editing && (
              <button
                onClick={() => setEditing(true)}
                className="text-xs text-gray-500 hover:text-black underline"
              >
                Edit
              </button>
            )}
          </div>

          {editing ? (
            <form onSubmit={handleSaveProfile} className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs text-gray-600 mb-1">First Name</label>
                  <input
                    type="text"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                    className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs text-gray-600 mb-1">Last Name</label>
                  <input
                    type="text"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                    className="w-full border border-gray-300 rounded px-3 py-2.5 text-sm outline-none focus:border-black transition-colors"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs text-gray-600 mb-1">Email</label>
                <input
                  type="email"
                  value={user.email}
                  disabled
                  className="w-full border border-gray-200 rounded px-3 py-2.5 text-sm bg-gray-50 text-gray-500"
                />
              </div>
              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={saving}
                  className="px-6 py-2 bg-black text-white text-xs uppercase tracking-wider font-medium hover:bg-gray-900 transition-colors disabled:opacity-50"
                >
                  {saving ? 'Saving...' : 'Save'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setEditing(false);
                    setFirstName(user.firstName);
                    setLastName(user.lastName);
                  }}
                  className="px-6 py-2 border border-gray-300 text-xs uppercase tracking-wider font-medium hover:border-black transition-colors"
                >
                  Cancel
                </button>
              </div>
            </form>
          ) : (
            <div className="space-y-2 text-sm">
              <p>
                <span className="text-gray-500 w-20 inline-block">Name</span>
                {user.firstName} {user.lastName}
              </p>
              <p>
                <span className="text-gray-500 w-20 inline-block">Email</span>
                {user.email}
              </p>
            </div>
          )}
        </section>

        {/* Addresses */}
        <section className="mb-12">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-sm font-semibold uppercase tracking-[0.15em]">Addresses</h2>
            <button
              onClick={() => setShowAddressForm(!showAddressForm)}
              className="flex items-center gap-1 text-xs text-gray-500 hover:text-black"
            >
              <Plus className="h-3 w-3" /> Add
            </button>
          </div>

          {showAddressForm && (
            <form onSubmit={handleAddAddress} className="mb-6 p-4 border border-gray-200 rounded-lg space-y-3">
              <input
                required
                type="text"
                placeholder="Street Address"
                value={newAddress.street}
                onChange={(e) => setNewAddress((p) => ({ ...p, street: e.target.value }))}
                className="w-full border border-gray-300 rounded px-3 py-2 text-sm outline-none focus:border-black transition-colors"
              />
              <div className="grid grid-cols-2 gap-3">
                <input
                  required
                  type="text"
                  placeholder="City"
                  value={newAddress.city}
                  onChange={(e) => setNewAddress((p) => ({ ...p, city: e.target.value }))}
                  className="border border-gray-300 rounded px-3 py-2 text-sm outline-none focus:border-black transition-colors"
                />
                <input
                  required
                  type="text"
                  placeholder="State"
                  value={newAddress.state}
                  onChange={(e) => setNewAddress((p) => ({ ...p, state: e.target.value }))}
                  className="border border-gray-300 rounded px-3 py-2 text-sm outline-none focus:border-black transition-colors"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <input
                  required
                  type="text"
                  placeholder="ZIP Code"
                  value={newAddress.zip}
                  onChange={(e) => setNewAddress((p) => ({ ...p, zip: e.target.value }))}
                  className="border border-gray-300 rounded px-3 py-2 text-sm outline-none focus:border-black transition-colors"
                />
                <input
                  type="text"
                  placeholder="Country"
                  value={newAddress.country}
                  onChange={(e) => setNewAddress((p) => ({ ...p, country: e.target.value }))}
                  className="border border-gray-300 rounded px-3 py-2 text-sm outline-none focus:border-black transition-colors"
                />
              </div>
              <div className="flex gap-2">
                <button
                  type="submit"
                  className="px-5 py-2 bg-black text-white text-xs uppercase tracking-wider font-medium hover:bg-gray-900 transition-colors"
                >
                  Save Address
                </button>
                <button
                  type="button"
                  onClick={() => setShowAddressForm(false)}
                  className="px-5 py-2 border border-gray-300 text-xs uppercase tracking-wider font-medium hover:border-black transition-colors"
                >
                  Cancel
                </button>
              </div>
            </form>
          )}

          {addresses.length === 0 ? (
            <p className="text-sm text-gray-500">No saved addresses</p>
          ) : (
            <div className="space-y-3">
              {addresses.map((addr) => (
                <div
                  key={addr.id}
                  className="flex items-start justify-between border border-gray-200 rounded-lg p-4"
                >
                  <div className="text-sm text-gray-700">
                    <p>{addr.street}</p>
                    <p>
                      {addr.city}, {addr.state} {addr.zip}
                    </p>
                    <p>{addr.country}</p>
                  </div>
                  <button
                    onClick={() => addr.id && handleDeleteAddress(addr.id)}
                    className="text-gray-400 hover:text-red-500 transition-colors p-1"
                    aria-label="Delete address"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>

        {/* Order History Link */}
        <section>
          <Link
            href="/account/orders"
            className="inline-block px-8 py-3 border border-black text-xs uppercase tracking-[0.2em] font-medium hover:bg-black hover:text-white transition-colors"
          >
            View Order History
          </Link>
        </section>
      </div>
    </div>
  );
}
