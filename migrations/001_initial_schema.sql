-- Initial database schema for Document Processing System
-- Migration: 001_initial_schema.sql

-- Create processes table
CREATE TABLE IF NOT EXISTS processes (
    process_id VARCHAR(36) PRIMARY KEY,
    folder_path TEXT NOT NULL,
    batch_size INTEGER NOT NULL,
    state VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    start_time TIMESTAMP NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP NULL,
    last_error TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create analysis_results table
CREATE TABLE IF NOT EXISTS analysis_results (
    id SERIAL PRIMARY KEY,
    process_id VARCHAR(36) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    word_count INTEGER NOT NULL DEFAULT 0,
    line_count INTEGER NOT NULL DEFAULT 0,
    char_count INTEGER NOT NULL DEFAULT 0,
    frequent_words JSONB NOT NULL DEFAULT '{}',
    summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (process_id) REFERENCES processes(process_id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_processes_state ON processes(state);
CREATE INDEX IF NOT EXISTS idx_processes_start_time ON processes(start_time);
CREATE INDEX IF NOT EXISTS idx_analysis_results_process_id ON analysis_results(process_id);
CREATE INDEX IF NOT EXISTS idx_analysis_results_file_name ON analysis_results(file_name);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for processes table
CREATE TRIGGER update_processes_updated_at 
    BEFORE UPDATE ON processes 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column(); 