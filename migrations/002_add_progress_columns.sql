-- Add progress tracking columns to processes table
-- Migration: 002_add_progress_columns.sql

-- Add total_files and processed_files columns to processes table
ALTER TABLE processes 
ADD COLUMN IF NOT EXISTS total_files INTEGER NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS processed_files INTEGER NOT NULL DEFAULT 0; 