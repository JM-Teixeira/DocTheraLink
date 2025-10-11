/**
 * App Component - Main Application Entry Point for Klyniva
 * 
 * This is the root component of the Klyniva healthcare management system.
 * It establishes the complete application infrastructure including:
 * 
 * - React Query for server state management with healthcare-optimized caching
 * - Authentication system with role-based access control for healthcare workers
 * - Plugin architecture for modular healthcare applications
 * - Global error handling and user notifications
 * - Feature flags for controlled feature rollouts
 * - Analytics and monitoring integrations
 * - Development tools for debugging
 * 
 * The component uses a layered provider pattern where each provider adds
 * specific functionality to the application context, creating a robust
 * foundation for healthcare data management.
 */

import {
  MutationCache,
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import React, { Suspense } from "react";

import { Toaster } from "@/components/ui/toaster";

import Loading from "@/components/Common/Loading";

import Integrations from "@/Integrations";
import PluginEngine from "@/PluginEngine";
import AuthUserProvider from "@/Providers/AuthUserProvider";
import HistoryAPIProvider from "@/Providers/HistoryAPIProvider";
import Routers from "@/Routers";
import { FeatureFlagsProvider } from "@/Utils/featureFlags";
import { handleHttpError } from "@/Utils/request/errorHandler";

import { PubSubProvider } from "./Utils/pubsubContext";
import { createSampleData } from "./Utils/sampleData";

// Configure React Query client optimized for healthcare data management
// Healthcare data requires careful caching strategies due to:
// - Patient privacy concerns (data should not persist too long)
// - Real-time updates needed for critical patient information
// - Network reliability in healthcare environments
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 2,                    // Retry failed requests twice (network issues common in hospitals)
      refetchOnWindowFocus: false, // Don't refetch when switching tabs (reduces server load)
      staleTime: 5 * 60 * 1000,   // Consider data fresh for 5 minutes (balance between freshness and performance)
    },
  },
  // Global error handling for all API queries
  // Ensures consistent error handling across the application
  queryCache: new QueryCache({
    onError: handleHttpError, // Centralized error handling for queries
  }),
  // Global error handling for all API mutations (create, update, delete operations)
  mutationCache: new MutationCache({
    onError: handleHttpError, // Centralized error handling for mutations
  }),
});

/**
 * Main App Component
 * 
 * Renders the complete DocTheraLink application with layered providers:
 * 
 * 1. QueryClientProvider - Server state management and API caching
 * 2. Suspense - Loading fallback for code-split components
 * 3. PubSubProvider - Event system for component communication
 * 4. PluginEngine - Modular plugin system for healthcare applications
 * 5. HistoryAPIProvider - Browser history and navigation management
 * 6. AuthUserProvider - Authentication and user session management
 * 7. FeatureFlagsProvider - Feature toggle system for controlled rollouts
 * 8. Routers - Main application routing system
 * 
 * Additional components:
 * - Toaster - Global notification system
 * - Integrations - External service integrations (Sentry, Plausible)
 * - ReactQueryDevtools - Development debugging tools
 */
const App = () => {
  // Initialize sample data for demo purposes
  React.useEffect(() => {
    if (import.meta.env.DEV) {
      createSampleData().then(() => {
        console.log('Klyniva sample data initialized');
      });
    }
  }, []);

  return (
    <QueryClientProvider client={queryClient}>
      {/* Suspense boundary for lazy-loaded components with Klyniva-themed loading */}
      <Suspense fallback={<Loading />}>
        {/* PubSub system for decoupled component communication */}
        <PubSubProvider>
          {/* Plugin engine enables modular healthcare applications */}
          <PluginEngine>
            {/* Browser history management for SPA navigation */}
            <HistoryAPIProvider>
              {/* Authentication provider with role-based access control
                  Shows SessionRouter (login page) when user is not authenticated */}
              <AuthUserProvider unauthorized={<Routers.SessionRouter />}>
                {/* Feature flags for controlled feature rollouts in healthcare environment */}
                <FeatureFlagsProvider>
                  {/* Main application router - handles all authenticated routes */}
                  <Routers.AppRouter />
                </FeatureFlagsProvider>
              </AuthUserProvider>

              {/* External service integrations for monitoring and analytics */}
              <Integrations.Sentry disabled={!import.meta.env.PROD} />
              <Integrations.Plausible />
            </HistoryAPIProvider>
            
            {/* Global toast notification system for user feedback */}
            <Toaster />
          </PluginEngine>
        </PubSubProvider>
      </Suspense>

      {/* Development tools - automatically excluded from production builds */}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};

export default App;
