'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Image from 'next/image';
import { Minus, Plus, Star, ShoppingBag } from 'lucide-react';
import { api, type Product, type Review, type ReviewSummary } from '@/lib/api';

export default function ProductDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [reviewSummary, setReviewSummary] = useState<ReviewSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [selectedImage, setSelectedImage] = useState(0);
  const [selectedSize, setSelectedSize] = useState('');
  const [selectedColor, setSelectedColor] = useState('');
  const [quantity, setQuantity] = useState(1);
  const [adding, setAdding] = useState(false);
  const [added, setAdded] = useState(false);

  useEffect(() => {
    if (!slug) return;
    setLoading(true);
    api.products
      .getBySlug(slug)
      .then((p) => {
        setProduct(p);
        if (p.variants.length > 0) {
          setSelectedSize(p.variants[0].size);
          setSelectedColor(p.variants[0].color);
        }
        return Promise.all([
          api.reviews.list(p.id).catch(() => []),
          api.reviews.getSummary(p.id).catch(() => null),
        ]);
      })
      .then(([r, s]) => {
        setReviews(r);
        setReviewSummary(s);
      })
      .finally(() => setLoading(false));
  }, [slug]);

  const sizes = [...new Set(product?.variants.map((v) => v.size) ?? [])];
  const colors = [...new Map(
    (product?.variants ?? []).map((v) => [v.color, { name: v.color, hex: v.colorHex }]),
  ).values()];

  const selectedVariant = product?.variants.find(
    (v) => v.size === selectedSize && v.color === selectedColor,
  );

  async function handleAddToCart() {
    if (!product || !selectedVariant) return;
    setAdding(true);
    try {
      await api.cart.addItem(product.id, selectedVariant.id, quantity);
      setAdded(true);
      setTimeout(() => setAdded(false), 2000);
    } catch {
      /* show error in real app */
    } finally {
      setAdding(false);
    }
  }

  if (loading) {
    return (
      <div className="pt-16">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
          <div className="grid lg:grid-cols-2 gap-12 animate-pulse">
            <div className="aspect-[3/4] bg-gray-200 rounded-lg" />
            <div className="space-y-4 py-4">
              <div className="h-8 bg-gray-200 rounded w-3/4" />
              <div className="h-6 bg-gray-200 rounded w-1/4" />
              <div className="h-4 bg-gray-200 rounded w-full mt-8" />
              <div className="h-4 bg-gray-200 rounded w-5/6" />
              <div className="h-4 bg-gray-200 rounded w-2/3" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!product) {
    return (
      <div className="pt-16 flex items-center justify-center min-h-[50vh]">
        <p className="text-gray-500">Product not found</p>
      </div>
    );
  }

  return (
    <div className="pt-16">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 py-12">
        <div className="grid lg:grid-cols-2 gap-8 lg:gap-12">
          {/* Image Gallery */}
          <div className="space-y-4">
            <div className="relative aspect-[3/4] bg-gray-100 rounded-lg overflow-hidden">
              {product.images[selectedImage] ? (
                <Image
                  src={product.images[selectedImage].url}
                  alt={product.images[selectedImage].alt || product.name}
                  fill
                  className="object-cover"
                  priority
                />
              ) : (
                <div className="absolute inset-0 bg-gradient-to-br from-gray-200 to-gray-300" />
              )}
            </div>
            {product.images.length > 1 && (
              <div className="flex gap-3 overflow-x-auto pb-1">
                {product.images.map((img, idx) => (
                  <button
                    key={img.id}
                    onClick={() => setSelectedImage(idx)}
                    className={`relative w-20 h-20 flex-shrink-0 rounded-md overflow-hidden border-2 transition-colors ${
                      idx === selectedImage ? 'border-black' : 'border-transparent hover:border-gray-300'
                    }`}
                  >
                    <Image src={img.url} alt={img.alt || ''} fill className="object-cover" />
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Product Info */}
          <div className="py-2 lg:py-4">
            {product.categoryName && (
              <p className="text-xs uppercase tracking-[0.2em] text-gray-500 mb-2">
                {product.categoryName}
              </p>
            )}
            <h1 className="text-2xl sm:text-3xl font-light tracking-tight font-display">
              {product.name}
            </h1>
            <p className="mt-2 text-xl">
              ${(selectedVariant?.price ?? product.basePrice).toFixed(2)}
            </p>

            {reviewSummary && reviewSummary.totalReviews > 0 && (
              <div className="mt-3 flex items-center gap-2">
                <div className="flex">
                  {Array.from({ length: 5 }).map((_, i) => (
                    <Star
                      key={i}
                      className={`h-4 w-4 ${
                        i < Math.round(reviewSummary.averageRating)
                          ? 'fill-yellow-400 text-yellow-400'
                          : 'text-gray-300'
                      }`}
                    />
                  ))}
                </div>
                <span className="text-sm text-gray-500">
                  ({reviewSummary.totalReviews} review{reviewSummary.totalReviews !== 1 ? 's' : ''})
                </span>
              </div>
            )}

            <p className="mt-6 text-sm text-gray-600 leading-relaxed">{product.description}</p>

            {/* Size Selector */}
            {sizes.length > 0 && (
              <div className="mt-8">
                <h3 className="text-xs font-semibold uppercase tracking-[0.15em] mb-3">Size</h3>
                <div className="flex flex-wrap gap-2">
                  {sizes.map((size) => (
                    <button
                      key={size}
                      onClick={() => setSelectedSize(size)}
                      className={`px-4 py-2 text-sm border rounded transition-colors ${
                        selectedSize === size
                          ? 'border-black bg-black text-white'
                          : 'border-gray-300 text-gray-700 hover:border-black'
                      }`}
                    >
                      {size}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Color Selector */}
            {colors.length > 0 && (
              <div className="mt-6">
                <h3 className="text-xs font-semibold uppercase tracking-[0.15em] mb-3">
                  Color — <span className="font-normal text-gray-500">{selectedColor}</span>
                </h3>
                <div className="flex gap-2">
                  {colors.map((color) => (
                    <button
                      key={color.name}
                      onClick={() => setSelectedColor(color.name)}
                      title={color.name}
                      className={`w-8 h-8 rounded-full border-2 transition-all ${
                        selectedColor === color.name
                          ? 'border-black scale-110'
                          : 'border-gray-200 hover:border-gray-400'
                      }`}
                      style={{ backgroundColor: color.hex }}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Quantity & Add to Cart */}
            <div className="mt-8 flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
              <div className="flex items-center border border-gray-300 rounded">
                <button
                  onClick={() => setQuantity(Math.max(1, quantity - 1))}
                  className="px-3 py-2.5 text-gray-600 hover:text-black transition-colors"
                  aria-label="Decrease quantity"
                >
                  <Minus className="h-4 w-4" />
                </button>
                <span className="w-12 text-center text-sm tabular-nums">{quantity}</span>
                <button
                  onClick={() => setQuantity(quantity + 1)}
                  className="px-3 py-2.5 text-gray-600 hover:text-black transition-colors"
                  aria-label="Increase quantity"
                >
                  <Plus className="h-4 w-4" />
                </button>
              </div>
              <button
                onClick={handleAddToCart}
                disabled={adding || !selectedVariant}
                className="flex-1 flex items-center justify-center gap-2 bg-black text-white py-3 px-8 text-sm uppercase tracking-[0.15em] font-medium hover:bg-gray-900 transition-colors disabled:opacity-50"
              >
                <ShoppingBag className="h-4 w-4" />
                {added ? 'Added!' : adding ? 'Adding...' : 'Add to Cart'}
              </button>
            </div>

            {selectedVariant && selectedVariant.stock <= 5 && selectedVariant.stock > 0 && (
              <p className="mt-3 text-xs text-amber-600">
                Only {selectedVariant.stock} left in stock
              </p>
            )}
            {selectedVariant && selectedVariant.stock === 0 && (
              <p className="mt-3 text-xs text-red-600">Out of stock</p>
            )}
          </div>
        </div>

        {/* Reviews Section */}
        <section className="mt-20 border-t border-gray-200 pt-12">
          <h2 className="text-xl font-light tracking-tight font-display mb-8">
            Customer Reviews
            {reviewSummary && (
              <span className="ml-2 text-base text-gray-400">
                ({reviewSummary.totalReviews})
              </span>
            )}
          </h2>

          {reviewSummary && reviewSummary.totalReviews > 0 && (
            <div className="mb-10 flex items-start gap-10">
              <div className="text-center">
                <p className="text-4xl font-light">{reviewSummary.averageRating.toFixed(1)}</p>
                <div className="flex mt-1">
                  {Array.from({ length: 5 }).map((_, i) => (
                    <Star
                      key={i}
                      className={`h-4 w-4 ${
                        i < Math.round(reviewSummary.averageRating)
                          ? 'fill-yellow-400 text-yellow-400'
                          : 'text-gray-300'
                      }`}
                    />
                  ))}
                </div>
                <p className="mt-1 text-xs text-gray-500">
                  {reviewSummary.totalReviews} review{reviewSummary.totalReviews !== 1 ? 's' : ''}
                </p>
              </div>
              <div className="flex-1 space-y-1.5">
                {[5, 4, 3, 2, 1].map((star) => {
                  const count = reviewSummary.distribution[star] ?? 0;
                  const pct =
                    reviewSummary.totalReviews > 0
                      ? (count / reviewSummary.totalReviews) * 100
                      : 0;
                  return (
                    <div key={star} className="flex items-center gap-2 text-sm">
                      <span className="w-3 text-right text-gray-500">{star}</span>
                      <Star className="h-3 w-3 text-gray-300" />
                      <div className="flex-1 h-2 bg-gray-100 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-yellow-400 rounded-full"
                          style={{ width: `${pct}%` }}
                        />
                      </div>
                      <span className="w-8 text-xs text-gray-400">{count}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {reviews.length === 0 ? (
            <p className="text-sm text-gray-500">No reviews yet. Be the first to review!</p>
          ) : (
            <div className="space-y-8">
              {reviews.map((review) => (
                <div key={review.id} className="border-b border-gray-100 pb-8">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium">{review.userName}</p>
                      <div className="flex mt-1">
                        {Array.from({ length: 5 }).map((_, i) => (
                          <Star
                            key={i}
                            className={`h-3.5 w-3.5 ${
                              i < review.rating
                                ? 'fill-yellow-400 text-yellow-400'
                                : 'text-gray-300'
                            }`}
                          />
                        ))}
                      </div>
                    </div>
                    <time className="text-xs text-gray-400">
                      {new Date(review.createdAt).toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric',
                      })}
                    </time>
                  </div>
                  <p className="mt-3 text-sm text-gray-600 leading-relaxed">{review.comment}</p>
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
