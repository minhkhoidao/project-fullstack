'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Truck, RotateCcw, Shield, Headphones } from 'lucide-react';
import { api, type Product, type Category } from '@/lib/api';
import { ProductCard } from '@/components/ProductCard';

const VALUE_PROPS = [
  {
    icon: Truck,
    title: 'Free Shipping',
    description: 'Complimentary shipping on all orders over $100',
  },
  {
    icon: RotateCcw,
    title: 'Easy Returns',
    description: '30-day return policy for a hassle-free experience',
  },
  {
    icon: Shield,
    title: 'Secure Payment',
    description: 'Your payment information is always protected',
  },
  {
    icon: Headphones,
    title: '24/7 Support',
    description: 'Our team is here to help whenever you need us',
  },
];

export default function HomePage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.products.getFeatured().catch(() => []),
      api.categories.list().catch(() => []),
    ]).then(([p, c]) => {
      setProducts(p.slice(0, 8));
      setCategories(c.slice(0, 4));
      setLoading(false);
    });
  }, []);

  return (
    <>
      {/* Hero */}
      <section className="relative h-[90vh] flex items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-black">
        <div className="absolute inset-0 bg-[url('/hero-pattern.svg')] opacity-5" />
        <div className="relative text-center text-white px-4 space-y-6">
          <p className="text-xs sm:text-sm uppercase tracking-[0.35em] text-gray-300">
            Spring / Summer 2026
          </p>
          <h1 className="text-4xl sm:text-6xl lg:text-7xl font-light tracking-tight font-display">
            New Season Arrivals
          </h1>
          <p className="text-base sm:text-lg text-gray-300 max-w-md mx-auto leading-relaxed">
            Discover the latest trends in contemporary fashion
          </p>
          <div className="pt-4">
            <Link
              href="/products"
              className="inline-block px-10 py-4 bg-white text-black text-xs uppercase tracking-[0.2em] font-medium hover:bg-gray-100 transition-colors"
            >
              Shop Now
            </Link>
          </div>
        </div>
      </section>

      {/* Featured Products */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <div className="text-center mb-12">
            <h2 className="text-2xl sm:text-3xl font-light tracking-tight font-display">
              Featured Collection
            </h2>
            <p className="mt-2 text-sm text-gray-500">Handpicked pieces for the season</p>
          </div>

          {loading ? (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="animate-pulse">
                  <div className="aspect-[3/4] bg-gray-200 rounded-lg" />
                  <div className="mt-3 h-4 bg-gray-200 rounded w-3/4" />
                  <div className="mt-2 h-3 bg-gray-200 rounded w-1/4" />
                </div>
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6">
              {products.map((product) => (
                <ProductCard key={product.id} product={product} />
              ))}
            </div>
          )}

          <div className="mt-12 text-center">
            <Link
              href="/products"
              className="inline-block px-8 py-3 border border-black text-xs uppercase tracking-[0.2em] font-medium hover:bg-black hover:text-white transition-colors"
            >
              View All Products
            </Link>
          </div>
        </div>
      </section>

      {/* Category Showcase */}
      {categories.length > 0 && (
        <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gray-50">
          <div className="mx-auto max-w-7xl">
            <div className="text-center mb-12">
              <h2 className="text-2xl sm:text-3xl font-light tracking-tight font-display">
                Shop by Category
              </h2>
            </div>
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
              {categories.map((cat) => (
                <Link
                  key={cat.id}
                  href={`/products?category=${cat.slug}`}
                  className="group relative aspect-[3/4] rounded-lg overflow-hidden bg-gray-200"
                >
                  {cat.image && (
                    <img
                      src={cat.image}
                      alt={cat.name}
                      className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                    />
                  )}
                  <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-black/10 to-transparent" />
                  <div className="absolute bottom-0 left-0 right-0 p-6">
                    <h3 className="text-lg font-light tracking-wide text-white">
                      {cat.name}
                    </h3>
                    <p className="text-xs text-gray-300 mt-1 uppercase tracking-widest group-hover:translate-x-1 transition-transform">
                      Shop Now &rarr;
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          </div>
        </section>
      )}

      {/* Why Shop With Us */}
      <section className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-7xl">
          <div className="text-center mb-12">
            <h2 className="text-2xl sm:text-3xl font-light tracking-tight font-display">
              Why Shop With Us
            </h2>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8 sm:gap-12">
            {VALUE_PROPS.map((item) => (
              <div key={item.title} className="text-center space-y-3">
                <div className="mx-auto flex h-12 w-12 items-center justify-center">
                  <item.icon className="h-6 w-6 text-gray-700" strokeWidth={1.5} />
                </div>
                <h3 className="text-sm font-medium tracking-wide">{item.title}</h3>
                <p className="text-xs text-gray-500 leading-relaxed">
                  {item.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>
    </>
  );
}
