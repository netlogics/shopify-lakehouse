import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Shopify Webhook Service",
  description: "Webhook receiver and data explorer for Shopify lakehouse",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50 text-gray-900 antialiased">
        <nav className="border-b border-gray-200 bg-white px-6 py-4">
          <div className="mx-auto flex max-w-7xl items-center justify-between">
            <a href="/" className="text-lg font-semibold text-gray-900">
              Shopify Webhook Service
            </a>
            <div className="flex gap-6 text-sm">
              <a href="/" className="text-gray-600 hover:text-gray-900">
                Dashboard
              </a>
              <a href="/explorer" className="text-gray-600 hover:text-gray-900">
                Data Explorer
              </a>
            </div>
          </div>
        </nav>
        <main className="mx-auto max-w-7xl px-6 py-8">{children}</main>
      </body>
    </html>
  );
}
