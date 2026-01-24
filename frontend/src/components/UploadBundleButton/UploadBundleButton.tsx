import { useRef, useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { uploadBundle } from '../../api';
import styles from './UploadBundleButton.module.css';

interface UploadBundleButtonProps {
  onSuccess?: (bundleId: string) => void;
  onError?: (error: Error) => void;
}

export function UploadBundleButton({ onSuccess, onError }: UploadBundleButtonProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: uploadBundle,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['repros'] });
      queryClient.invalidateQueries({ queryKey: ['filters'] });
      onSuccess?.(data.bundle_id);
    },
    onError: (error: Error) => {
      onError?.(error);
    },
  });

  const handleFileSelect = (file: File | null) => {
    if (!file) return;

    // Validate file type
    if (!file.name.endsWith('.zip')) {
      onError?.(new Error('Please select a ZIP file containing the repro bundle'));
      return;
    }

    // Validate file size (max 500MB)
    const maxSize = 500 * 1024 * 1024;
    if (file.size > maxSize) {
      onError?.(new Error('File size exceeds 500MB limit'));
      return;
    }

    mutation.mutate(file);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] ?? null;
    handleFileSelect(file);
    // Reset input so the same file can be selected again
    e.target.value = '';
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);

    const file = e.dataTransfer.files[0] ?? null;
    handleFileSelect(file);
  };

  return (
    <div
      className={`${styles.container} ${isDragOver ? styles.dragOver : ''}`}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      <input
        ref={fileInputRef}
        type="file"
        accept=".zip"
        onChange={handleInputChange}
        className={styles.fileInput}
        disabled={mutation.isPending}
      />
      <button
        type="button"
        onClick={handleClick}
        disabled={mutation.isPending}
        className={styles.uploadBtn}
      >
        {mutation.isPending ? (
          <>
            <span className={styles.spinner} />
            Uploading...
          </>
        ) : (
          <>
            <svg 
              className={styles.uploadIcon} 
              viewBox="0 0 24 24" 
              fill="none" 
              stroke="currentColor" 
              strokeWidth="2"
            >
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="17 8 12 3 7 8" />
              <line x1="12" y1="3" x2="12" y2="15" />
            </svg>
            Upload Bundle
          </>
        )}
      </button>
    </div>
  );
}
