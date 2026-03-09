'use client';

import { useEffect, useState, useCallback } from 'react';
import { useSearchParams } from 'next/navigation';
import { SlidersHorizontal, X } from 'lucide-react';
import { api, type Product, type Category, type ProductFilters } from '@/lib/api';
import { ProductCard } from '@/components/ProductCard';

const SIZES = ['XS', 'S', 'M', 'L', 'XL', 'XXL'];
const PRICE_RANGES = [
  { label: 'Under $50', min: 0, max: 50 },
  { label: '$50 – $100', min: 50, max: 100 },
  { label: '$100 – $200', min: 100, max: 200 },
  { label: '$200 – $500', min: 200, max: 500 },
  { label: '$500+', min: 500, max: undefined },
];
const COLORS = [
  { name: 'Black', hex: '#000000' },
  { name: 'White', hex: '#FFFFFF' },
  { name: 'Navy', hex: '#1B2A4A' },
  { name: 'Gray', hex: '#9CA3AF' },
  { name: 'Beige', hex: '#D4C5A9' },
  { name: 'Red', hex: '#DC2626' },
  { name: 'Green', hex: '#16A34A' },
  { name: 'Blue', hex: '#2563EB' },
];

export default function ProductsPage() {
  const searchParams = useSearchParams();
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [nextCursor, setNextCursor] = useState<string | undefined>();
  const [filtersOpen, setFiltersOpen] = useState(false);

  const [selectedCategory, setSelectedCategory] = useState(searchParams.get('category') || '');
  const [selectedSize, setSelectedSize] = useState('');
  const [selectedColor, setSelectedColor] = useState('');
  const [selectedPrice, setSelectedPrice] = useState<{ min?: number; max?: number } | null>(null);

  const buildFilters = useCallback((): ProductFilters => {
    const filters: ProductFilters = { limit: 12 };
    if (selectedCategory) filters.category = selectedCategory;
    if (selectedSize) filters.size = selectedSize;
    if (selectedColor) filters.color = selectedColor;
    if (selectedPrice) {
      filters.minPrice = selectedPrice.min;
      filters.maxPrice = selectedPrice.max;
    }
    return filters;
  }, [selectedCategory, selectedSize, selectedColor, selectedPrice]);

  const fetchProducts = useCallback(async () => {
    setLoading(true);
    try {
      const res = await api.products.list(buildFilters());
      setProducts(res.products);
      setNextCursor(res.nextCursor);
    } catch {
      setProducts([]);
    } finally {
      setLoading(false);
    }
  }, [buildFilters]);

  useEffect(() => {
    fetchProducts();
  }, [fetchProducts]);

  useEffect(() => {
    api.categories.list().then(setCategories).catch(() => {});
  }, []);

  async function loadMore() {
    if (!nextCursor || loadingMore) return;
    setLoadingMore(true);
    try {
      const res = await api.products.list({ ...buildFilters(), cursor: nextCursor });
      setProducts((prev) => [...prev, ...res.products]);
      setNextCursor(res.nextCursor);
    } catch {
      /* ignore */
    } finally {
      setLoadingMore(false);
    }
  }

  function clearFilters() {
    setSelectedCategory('');
    setSelectedSize('');
    setSelectedColor('');
    setSelectedPrice(null);
  }

  const hasActiveFilters = selectedCategory || selectedSize || selectedColor || selectedPrice;

  return (
    <div className="pt-16">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <h1 className="text-2xl font-light tracking-tight font-display">All Products</h1>
          <button
            onClick={() => setFiltersOpen(!filtersOpen)}
            className="lg:hidden flex items-center gap-2 text-sm text-gray-700 border border-gray-300 px-4 py-2 rounded-md hover:border-black transition-colors"
          >
            <SlidersHorizontal className="h-4 w-4" />
            Filters
          </button>
        </div>

        <div className="flex gap-8">
          {/* Sidebar Filters */}
          <aside
            className={`
              fixed inset-0 z-40 bg-white p-6 overflow-y-auto transition-transform lg:transition-none
              lg:static lg:z-auto lg:w-56 lg:shrink-0 lg:p-0 lg:bg-transparent
              ${filtersOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
            `}
          >
            <div className="flex items-center justify-between mb-6 lg:hidden">
              <h2 className="font-medium">Filters</h2>
              <button onClick={() => setFiltersOpen(false)}>
                <X className="h-5 w-5" />
              </button>
            </div>

            {hasActiveFilters && (
              <button
                onClick={clearFilters}
                className="mb-6 text-xs uppercase tracking-wider text-gray-500 hover:text-black underline"
              >
                Clear all filters
              </button>
            )}

            {/* Category */}
            <div className="mb-8">
              <h3 className="text-xs font-semibold uppercase tracking-[0.15em] text-gray-900 mb-3">
                Category
              </h3>
              <div className="space-y-2">
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    onClick={() => {
                      setSelectedCategory(selectedCategory === cat.slug ? '' : cat.slug);
                      setFiltersOpen(false);
                    }}
                    className={`block text-sm transition-colors ${
                      selectedCategory === cat.slug
                        ? 'text-black font-medium'
                        : 'text-gray-500 hover:text-black'
                    }`}
                  >
                    {cat.name}
                  </button>
                ))}
              </div>
            </div>

            {/* Price Range */}
            <div className="mb-8">
              <h3 className="text-xs font-semibold uppercase tracking-[0.15em] text-gray-900 mb-3">
                Price
              </h3>
              <div className="space-y-2">
                {PRICE_RANGES.map((range) => {
                  const active =
                    selectedPrice?.min === range.min && selectedPrice?.max === range.max;
                  return (
                    <button
                      key={range.label}
                      onClick={() => {
                        setSelectedPrice(active ? null : { min: range.min, max: range.max });
                        setFiltersOpen(false);
                      }}
                      className={`block text-sm transition-colors ${
                        active ? 'text-black font-medium' : 'text-gray-500 hover:text-black'
                      }`}
                    >
                      {range.label}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Size */}
            <div className="mb-8">
              <h3 className="text-xs font-semibold uppercase tracking-[0.15em] text-gray-900 mb-3">
                Size
              </h3>
              <div className="flex flex-wrap gap-2">
                {SIZES.map((size) => (
                  <button
                    key={size}
                    onClick={() => {
                      setSelectedSize(selectedSize === size ? '' : size);
                      setFiltersOpen(false);
                    }}
                    className={`px-3 py-1.5 text-xs border rounded transition-colors ${
                      selectedSize === size
                        ? 'border-black bg-black text-white'
                        : 'border-gray-300 text-gray-600 hover:border-black'
                    }`}
                  >
                    {size}
                  </button>
                ))}
              </div>
            </div>

            {/* Color */}
            <div className="mb-8">
              <h3 className="text-xs font-semibold uppercase tracking-[0.15em] text-gray-900 mb-3">
                Color
              </h3>
              <div className="flex flex-wrap gap-2">
                {COLORS.map((color) => (
                  <button
                    key={color.name}
                    onClick={() => {
                      setSelectedColor(selectedColor === color.name ? '' : color.name);
                      setFiltersOpen(false);
                    }}
                    title={color.name}
                    className={`w-7 h-7 rounded-full border-2 transition-all ${
                      selectedColor === color.name
                        ? 'border-black scale-110'
                        : 'border-gray-200 hover:border-gray-400'
                    }`}
                    style={{ backgroundColor: color.hex }}
                  />
                ))}
              </div>
            </div>
          </aside>

          {/* Overlay for mobile filter */}
          {filtersOpen && (
            <div
              className="fixed inset-0 z-30 bg-black/30 lg:hidden"
              onClick={() => setFiltersOpen(false)}
            />
          )}

          {/* Product Grid */}
          <div className="flex-1">
            {loading ? (
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6">
                {Array.from({ length: 12 }).map((_, i) => (
                  <div key={i} className="animate-pulse">
                    <div className="aspect-[3/4] bg-gray-200 rounded-lg" />
                    <div className="mt-3 h-4 bg-gray-200 rounded w-3/4" />
                    <div className="mt-2 h-3 bg-gray-200 rounded w-1/4" />
                  </div>
                ))}
              </div>
            ) : products.length === 0 ? (
              <div className="text-center py-20">
                <p className="text-gray-500">No products found</p>
                {hasActiveFilters && (
                  <button
                    onClick={clearFilters}
                    className="mt-4 text-sm underline text-gray-700 hover:text-black"
                  >
                    Clear filters
                  </button>
                )}
              </div>
            ) : (
              <>
                <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 sm:gap-6">
                  {products.map((product) => (
                    <ProductCard key={product.id} product={product} />
                  ))}
                </div>

                {nextCursor && (
                  <div className="mt-12 text-center">
                    <button
                      onClick={loadMore}
                      disabled={loadingMore}
                      className="px-10 py-3 border border-black text-xs uppercase tracking-[0.2em] font-medium hover:bg-black hover:text-white transition-colors disabled:opacity-50"
                    >
                      {loadingMore ? 'Loading...' : 'Load More'}
                    </button>
                  </div>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
