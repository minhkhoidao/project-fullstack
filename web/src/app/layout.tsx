import type { Metadata } from 'next';
import { AuthHydration } from '@/components/AuthHydration';
import { Navbar } from '@/components/Navbar';
import { Footer } from '@/components/Footer';
import './globals.css';

export const metadata: Metadata = {
  title: 'FASHION — Contemporary Style',
  description:
    'Discover the latest in contemporary fashion. Shop curated collections of premium clothing and accessories.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <AuthHydration />
        <Navbar />
        <main className="min-h-screen">{children}</main>
        <Footer />
      </body>
    </html>
  );
}
