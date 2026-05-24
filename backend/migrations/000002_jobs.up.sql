ALTER TABLE analysis_jobs ADD COLUMN IF NOT EXISTS progress_percent INT DEFAULT 0;
ALTER TABLE analysis_jobs ADD COLUMN IF NOT EXISTS progress_stage TEXT;
ALTER TABLE analysis_jobs ADD COLUMN IF NOT EXISTS output_files JSONB;
