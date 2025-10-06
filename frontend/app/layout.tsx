import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "StoryForge",
  description: "Collaborative data storytelling workbench",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-slate-50 text-slate-900">
        <div className="mx-auto flex min-h-screen w-full max-w-6xl flex-col px-6 py-8">
          <header className="mb-6 flex flex-col gap-2 border-b border-slate-200 pb-4 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight text-slate-900">StoryForge</h1>
              <p className="text-sm text-slate-600">Coordinate notebooks, narration, and realtime feedback.</p>
            </div>
            <div className="flex items-center gap-2 text-sm text-slate-500">
              <span className="rounded-full bg-primary-100 px-3 py-1 font-medium text-primary-700">MVP</span>
              <span>prototype</span>
            </div>
          </header>
          <main className="flex-1">{children}</main>
        </div>
      </body>
    </html>
  );
}
