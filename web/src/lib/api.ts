// ─── Auth Types ──────────────────────────────────────────────────────────────

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role?: string;
}

export interface ApiError {
  message: string;
  status?: number;
}

// ─── Product Types ───────────────────────────────────────────────────────────

export interface Product {
  id: string;
  name: string;
  slug: string;
  description: string;
  basePrice: number;
  images: ProductImage[];
  variants: ProductVariant[];
  categoryId: string;
  categoryName: string;
}

export interface ProductImage {
  id: string;
  url: string;
  alt: string;
  position: number;
}

export interface ProductVariant {
  id: string;
  sku: string;
  size: string;
  color: string;
  colorHex: string;
  price: number;
  stock: number;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  image?: string;
}

export interface ProductFilters {
  category?: string;
  minPrice?: number;
  maxPrice?: number;
  size?: string;
  color?: string;
  cursor?: string;
  limit?: number;
}

export interface ProductListResponse {
  products: Product[];
  nextCursor?: string;
}

// ─── Cart Types ──────────────────────────────────────────────────────────────

export interface Cart {
  id: string;
  items: CartItem[];
  subtotal: number;
}

export interface CartItem {
  id: string;
  productId: string;
  variantId: string;
  product: Product;
  variant: ProductVariant;
  quantity: number;
}

// ─── Order Types ─────────────────────────────────────────────────────────────

export type OrderStatus = 'pending' | 'paid' | 'shipped' | 'delivered' | 'cancelled';

export interface Order {
  id: string;
  orderNumber: string;
  status: OrderStatus;
  items: OrderItem[];
  subtotal: number;
  shipping: number;
  total: number;
  shippingAddress: Address;
  createdAt: string;
}

export interface OrderItem {
  id: string;
  productName: string;
  variantInfo: string;
  quantity: number;
  unitPrice: number;
  imageUrl: string;
}

export interface OrderSummary {
  id: string;
  orderNumber: string;
  status: OrderStatus;
  total: number;
  itemCount: number;
  customerEmail: string;
  createdAt: string;
}

// ─── Address Types ───────────────────────────────────────────────────────────

export interface Address {
  id?: string;
  street: string;
  city: string;
  state: string;
  zip: string;
  country: string;
}

// ─── Review Types ────────────────────────────────────────────────────────────

export interface Review {
  id: string;
  userId: string;
  userName: string;
  rating: number;
  comment: string;
  createdAt: string;
}

export interface ReviewSummary {
  averageRating: number;
  totalReviews: number;
  distribution: Record<number, number>;
}

// ─── Admin Types ─────────────────────────────────────────────────────────────

export interface DashboardStats {
  totalRevenue: number;
  totalOrders: number;
  totalProducts: number;
  totalCustomers: number;
  recentOrders: OrderSummary[];
}

export interface RevenueByDay {
  date: string;
  revenue: number;
  orders: number;
}

export interface TopProduct {
  id: string;
  name: string;
  revenue: number;
  unitsSold: number;
}

// ─── HTTP Client ─────────────────────────────────────────────────────────────

const BASE_URL = process.env.NEXT_PUBLIC_API_URL || '/api/v1';

function getHeaders(): HeadersInit {
  const headers: HeadersInit = { 'Content-Type': 'application/json' };
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('access_token');
    if (token) headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: { ...getHeaders(), ...options?.headers },
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.message || `Request failed (${res.status})`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

function toQuery(params?: object): string {
  if (!params) return '';
  const entries = Object.entries(params).filter(([, v]) => v != null && v !== '');
  return new URLSearchParams(entries.map(([k, v]) => [k, String(v)])).toString();
}

// ─── Auth API (consumed by lib/auth.ts) ──────────────────────────────────────

interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: User;
}

export const auth = {
  login: (data: LoginRequest) =>
    request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    }).then((data) => ({ data })),

  register: (data: RegisterRequest) =>
    request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    }).then((data) => ({ data })),

  getProfile: () =>
    request<User>('/auth/me').then((data) => ({ data })),
};

// ─── Storefront API ──────────────────────────────────────────────────────────

export const api = {
  products: {
    list: (params?: ProductFilters) =>
      request<ProductListResponse>(`/products?${toQuery(params)}`),
    getBySlug: (slug: string) => request<Product>(`/products/${slug}`),
    getFeatured: () => request<Product[]>('/products/featured'),
  },

  categories: {
    list: () => request<Category[]>('/categories'),
  },

  cart: {
    get: () => request<Cart>('/cart'),
    addItem: (productId: string, variantId: string, quantity: number) =>
      request<Cart>('/cart/items', {
        method: 'POST',
        body: JSON.stringify({ productId, variantId, quantity }),
      }),
    updateItem: (itemId: string, quantity: number) =>
      request<Cart>(`/cart/items/${itemId}`, {
        method: 'PATCH',
        body: JSON.stringify({ quantity }),
      }),
    removeItem: (itemId: string) =>
      request<Cart>(`/cart/items/${itemId}`, { method: 'DELETE' }),
  },

  orders: {
    create: (data: { shippingAddress: Address }) =>
      request<Order>('/orders', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    list: () => request<Order[]>('/orders'),
    get: (id: string) => request<Order>(`/orders/${id}`),
  },

  reviews: {
    list: (productId: string) =>
      request<Review[]>(`/products/${productId}/reviews`),
    getSummary: (productId: string) =>
      request<ReviewSummary>(`/products/${productId}/reviews/summary`),
  },

  addresses: {
    list: () => request<Address[]>('/addresses'),
    create: (data: Omit<Address, 'id'>) =>
      request<Address>('/addresses', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      request<void>(`/addresses/${id}`, { method: 'DELETE' }),
  },

  profile: {
    get: () => request<User>('/auth/me'),
    update: (data: Partial<User>) =>
      request<User>('/auth/me', {
        method: 'PATCH',
        body: JSON.stringify(data),
      }),
  },
};

// ─── Admin API ───────────────────────────────────────────────────────────────

export function getDashboard() {
  return request<DashboardStats>('/admin/dashboard');
}

export function getRevenue(days = 30) {
  return request<RevenueByDay[]>(`/admin/analytics/revenue?days=${days}`);
}

export function getTopProducts(days = 30) {
  return request<TopProduct[]>(`/admin/analytics/top-products?days=${days}`);
}

export function listAllOrders(params?: {
  status?: string;
  cursor?: string;
  limit?: number;
}) {
  return request<{ orders: OrderSummary[]; nextCursor?: string }>(
    `/admin/orders?${toQuery(params)}`,
  );
}

export function updateOrderStatus(id: string, status: OrderStatus) {
  return request<void>(`/admin/orders/${id}/status`, {
    method: 'PATCH',
    body: JSON.stringify({ status }),
  });
}

export function listProducts(params?: {
  cursor?: string;
  limit?: number;
  search?: string;
}) {
  return request<ProductListResponse>(`/admin/products?${toQuery(params)}`);
}

export function deleteProduct(id: string) {
  return request<void>(`/admin/products/${id}`, { method: 'DELETE' });
}

export function listAllUsers(params?: {
  cursor?: string;
  limit?: number;
  search?: string;
}) {
  return request<{ users: User[]; nextCursor?: string }>(
    `/admin/users?${toQuery(params)}`,
  );
}

export function updateUserRole(id: string, role: string) {
  return request<void>(`/admin/users/${id}/role`, {
    method: 'PATCH',
    body: JSON.stringify({ role }),
  });
}
