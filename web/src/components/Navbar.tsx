'use client';

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Search, ShoppingBag, User, Menu, X } from 'lucide-react';
import { useAuthStore } from '@/stores/auth';
import { api } from '@/lib/api';

export function Navbar() {
  const { user, isAuthenticated, logout } = useAuthStore();
  const router = useRouter();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [cartCount, setCartCount] = useState(0);
  const userMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    api.cart
      .get()
      .then((cart) => setCartCount(cart.items.length))
      .catch(() => {});
  }, []);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setUserMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (searchQuery.trim()) {
      router.push(`/products?q=${encodeURIComponent(searchQuery.trim())}`);
      setSearchOpen(false);
      setSearchQuery('');
    }
  }

  function handleLogout() {
    logout();
    setUserMenuOpen(false);
    router.push('/');
  }

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-white shadow-sm">
      <nav className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          {/* Mobile hamburger */}
          <button
            className="lg:hidden p-2 -ml-2 text-gray-700 hover:text-black"
            onClick={() => setMobileOpen(!mobileOpen)}
            aria-label="Toggle menu"
          >
            {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>

          {/* Logo */}
          <Link href="/" className="text-xl tracking-[0.25em] font-semibold text-black">
            FASHION
          </Link>

          {/* Desktop nav links */}
          <div className="hidden lg:flex items-center gap-8 ml-12">
            <Link href="/products" className="text-sm tracking-wide text-gray-700 hover:text-black transition-colors">
              Shop
            </Link>
            <Link href="/products?view=categories" className="text-sm tracking-wide text-gray-700 hover:text-black transition-colors">
              Categories
            </Link>
          </div>

          {/* Spacer */}
          <div className="flex-1" />

          {/* Desktop search */}
          <div className="hidden lg:block relative mr-6">
            {searchOpen ? (
              <form onSubmit={handleSearch} className="flex items-center">
                <input
                  autoFocus
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search..."
                  className="w-60 border-b border-gray-300 bg-transparent py-1 text-sm outline-none focus:border-black transition-colors"
                />
                <button
                  type="button"
                  onClick={() => { setSearchOpen(false); setSearchQuery(''); }}
                  className="ml-2 text-gray-400 hover:text-black"
                >
                  <X className="h-4 w-4" />
                </button>
              </form>
            ) : (
              <button
                onClick={() => setSearchOpen(true)}
                className="text-gray-700 hover:text-black transition-colors"
                aria-label="Search"
              >
                <Search className="h-5 w-5" />
              </button>
            )}
          </div>

          {/* Cart */}
          <Link href="/cart" className="relative p-2 text-gray-700 hover:text-black transition-colors">
            <ShoppingBag className="h-5 w-5" />
            {cartCount > 0 && (
              <span className="absolute -top-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-black text-[10px] font-medium text-white">
                {cartCount}
              </span>
            )}
          </Link>

          {/* User menu */}
          <div className="relative ml-2" ref={userMenuRef}>
            <button
              onClick={() => setUserMenuOpen(!userMenuOpen)}
              className="p-2 text-gray-700 hover:text-black transition-colors"
              aria-label="User menu"
            >
              <User className="h-5 w-5" />
            </button>

            {userMenuOpen && (
              <div className="absolute right-0 mt-2 w-48 rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5">
                {isAuthenticated ? (
                  <>
                    <div className="px-4 py-2 border-b border-gray-100">
                      <p className="text-sm font-medium truncate">
                        {user?.firstName} {user?.lastName}
                      </p>
                      <p className="text-xs text-gray-500 truncate">{user?.email}</p>
                    </div>
                    <Link
                      href="/account"
                      onClick={() => setUserMenuOpen(false)}
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      Profile
                    </Link>
                    <Link
                      href="/account/orders"
                      onClick={() => setUserMenuOpen(false)}
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      Orders
                    </Link>
                    <button
                      onClick={handleLogout}
                      className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      Logout
                    </button>
                  </>
                ) : (
                  <>
                    <Link
                      href="/login"
                      onClick={() => setUserMenuOpen(false)}
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      Login
                    </Link>
                    <Link
                      href="/register"
                      onClick={() => setUserMenuOpen(false)}
                      className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                    >
                      Register
                    </Link>
                  </>
                )}
              </div>
            )}
          </div>
        </div>
      </nav>

      {/* Mobile menu */}
      {mobileOpen && (
        <div className="lg:hidden border-t border-gray-100 bg-white">
          <div className="px-4 py-4 space-y-4">
            <form onSubmit={handleSearch}>
              <div className="flex items-center border-b border-gray-200 pb-2">
                <Search className="h-4 w-4 text-gray-400 mr-2" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search..."
                  className="flex-1 bg-transparent text-sm outline-none"
                />
              </div>
            </form>
            <Link
              href="/products"
              onClick={() => setMobileOpen(false)}
              className="block text-sm tracking-wide text-gray-700"
            >
              Shop
            </Link>
            <Link
              href="/products?view=categories"
              onClick={() => setMobileOpen(false)}
              className="block text-sm tracking-wide text-gray-700"
            >
              Categories
            </Link>
            {!isAuthenticated && (
              <>
                <Link href="/login" onClick={() => setMobileOpen(false)} className="block text-sm text-gray-700">
                  Login
                </Link>
                <Link href="/register" onClick={() => setMobileOpen(false)} className="block text-sm text-gray-700">
                  Register
                </Link>
              </>
            )}
            {isAuthenticated && (
              <>
                <Link href="/account" onClick={() => setMobileOpen(false)} className="block text-sm text-gray-700">
                  Profile
                </Link>
                <Link href="/account/orders" onClick={() => setMobileOpen(false)} className="block text-sm text-gray-700">
                  Orders
                </Link>
                <button onClick={handleLogout} className="block text-sm text-gray-700">
                  Logout
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </header>
  );
}
