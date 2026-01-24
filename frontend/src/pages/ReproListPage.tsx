import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getRepros, getFilters } from '../api';
import { ReproCard, UploadBundleButton } from '../components';
import type { ReproFilters, Platform } from '../types';
import styles from './ReproListPage.module.css';

interface Toast {
  id: number;
  type: 'success' | 'error';
  message: string;
}

export function ReproListPage() {
  const [filters, setFilters] = useState<ReproFilters>({
    page: 1,
    limit: 20,
  });
  const [searchInput, setSearchInput] = useState('');
  const [toasts, setToasts] = useState<Toast[]>([]);

  // Fetch filter options
  const { data: filterOptions } = useQuery({
    queryKey: ['filters'],
    queryFn: getFilters,
  });

  // Fetch repros
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['repros', filters],
    queryFn: () => getRepros(filters),
  });

  const handleFilterChange = (key: keyof ReproFilters, value: string | undefined) => {
    setFilters(prev => ({
      ...prev,
      [key]: value || undefined,
      page: 1, // Reset to first page on filter change
    }));
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setFilters(prev => ({
      ...prev,
      search: searchInput || undefined,
      page: 1,
    }));
  };

  const totalPages = useMemo(() => {
    if (!data) return 1;
    return Math.ceil(data.total / (filters.limit || 20));
  }, [data, filters.limit]);

  const showToast = (type: 'success' | 'error', message: string) => {
    const id = Date.now();
    setToasts(prev => [...prev, { id, type, message }]);
    // Auto-remove after 5 seconds
    setTimeout(() => {
      setToasts(prev => prev.filter(t => t.id !== id));
    }, 5000);
  };

  const handleUploadSuccess = (bundleId: string) => {
    showToast('success', `Bundle uploaded successfully: ${bundleId}`);
  };

  const handleUploadError = (error: Error) => {
    showToast('error', error.message || 'Failed to upload bundle');
  };

  const dismissToast = (id: number) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };

  return (
    <div className={styles.container}>
      <header className={styles.header}>
        <h1 className={styles.title}>
          <span className={styles.logoIcon}>🐛</span>
          <span className={styles.logoText}>BugIt</span>
        </h1>
        <div className={styles.headerActions}>
          <UploadBundleButton
            onSuccess={handleUploadSuccess}
            onError={handleUploadError}
          />
          <button className={styles.refreshBtn} onClick={() => refetch()}>
            Refresh
          </button>
        </div>
      </header>

      <div className={styles.filters}>
        <select
          value={filters.build || ''}
          onChange={(e) => handleFilterChange('build', e.target.value)}
          className={styles.select}
        >
          <option value="">All Builds</option>
          {filterOptions?.builds.map(build => (
            <option key={build} value={build}>{build}</option>
          ))}
        </select>

        <select
          value={filters.platform || ''}
          onChange={(e) => handleFilterChange('platform', e.target.value as Platform)}
          className={styles.select}
        >
          <option value="">All Platforms</option>
          {filterOptions?.platforms.map(platform => (
            <option key={platform} value={platform}>{platform}</option>
          ))}
        </select>

        <select
          value={filters.map || ''}
          onChange={(e) => handleFilterChange('map', e.target.value)}
          className={styles.select}
        >
          <option value="">All Maps</option>
          {filterOptions?.maps.map(map => (
            <option key={map} value={map}>{map}</option>
          ))}
        </select>

        <form onSubmit={handleSearch} className={styles.searchForm}>
          <input
            type="text"
            placeholder="Search tags, notes..."
            value={searchInput}
            onChange={(e) => setSearchInput(e.target.value)}
            className={styles.searchInput}
          />
          <button type="submit" className={styles.searchBtn}>
            Search
          </button>
        </form>
      </div>

      <main className={styles.main}>
        {isLoading && (
          <div className={styles.loading}>Loading repros...</div>
        )}

        {error && (
          <div className={styles.error}>
            Failed to load repros. Is the backend running?
          </div>
        )}

        {data && data.repros.length === 0 && (
          <div className={styles.empty}>
            No repros found matching your filters.
          </div>
        )}

        {data && data.repros.length > 0 && (
          <>
            <div className={styles.reproList}>
              {data.repros.map(repro => (
                <ReproCard key={repro.bundle_id} repro={repro} />
              ))}
            </div>

            <div className={styles.pagination}>
              <span className={styles.pageInfo}>
                Showing {((filters.page || 1) - 1) * (filters.limit || 20) + 1}–
                {Math.min((filters.page || 1) * (filters.limit || 20), data.total)} of {data.total}
              </span>
              
              <div className={styles.pageButtons}>
                <button
                  disabled={(filters.page || 1) <= 1}
                  onClick={() => setFilters(prev => ({ ...prev, page: (prev.page || 1) - 1 }))}
                  className={styles.pageBtn}
                >
                  Previous
                </button>
                <span className={styles.pageNumber}>
                  Page {filters.page || 1} of {totalPages}
                </span>
                <button
                  disabled={(filters.page || 1) >= totalPages}
                  onClick={() => setFilters(prev => ({ ...prev, page: (prev.page || 1) + 1 }))}
                  className={styles.pageBtn}
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </main>

      {/* Toast notifications */}
      {toasts.length > 0 && (
        <div className={styles.toastContainer}>
          {toasts.map(toast => (
            <div
              key={toast.id}
              className={`${styles.toast} ${styles[toast.type]}`}
              onClick={() => dismissToast(toast.id)}
            >
              <span className={styles.toastMessage}>{toast.message}</span>
              <button className={styles.toastDismiss} aria-label="Dismiss">
                &times;
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
