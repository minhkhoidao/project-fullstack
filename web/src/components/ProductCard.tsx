import Link from 'next/link';
import Image from 'next/image';
import type { Product } from '@/lib/api';

export function ProductCard({ product }: { product: Product }) {
  const primaryImage = product.images?.[0];

  return (
    <Link href={`/products/${product.slug}`} className="group">
      <div className="rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow duration-300 bg-white">
        <div className="relative aspect-[3/4] bg-gray-100 overflow-hidden">
          {primaryImage ? (
            <Image
              src={primaryImage.url}
              alt={primaryImage.alt || product.name}
              fill
              sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 25vw"
              className="object-cover group-hover:scale-105 transition-transform duration-500"
            />
          ) : (
            <div className="absolute inset-0 bg-gradient-to-br from-gray-200 to-gray-300" />
          )}
          <div className="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors duration-300 flex items-center justify-center">
            <span className="opacity-0 group-hover:opacity-100 transition-opacity duration-300 bg-white/95 backdrop-blur-sm px-6 py-2.5 text-xs uppercase tracking-[0.15em] font-medium">
              Quick View
            </span>
          </div>
        </div>
        <div className="p-4">
          <h3 className="text-sm font-medium text-gray-900 truncate">{product.name}</h3>
          <p className="mt-1 text-sm text-gray-600">${product.basePrice.toFixed(2)}</p>
        </div>
      </div>
    </Link>
  );
}
