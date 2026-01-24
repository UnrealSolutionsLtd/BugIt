import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReproListPage } from './pages/ReproListPage';
import { ReproViewerPage } from './pages/ReproViewerPage';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000, // 30 seconds
      retry: 1,
    },
  },
});

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<ReproListPage />} />
          <Route path="/repro/:id" element={<ReproViewerPage />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
