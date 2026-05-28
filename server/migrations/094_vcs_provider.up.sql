-- +migrate Up
-- Add provider column to github_installation and github_pull_request tables
-- to support multiple VCS providers (github, gitee)

ALTER TABLE github_installation ADD COLUMN provider TEXT NOT NULL DEFAULT 'github';
ALTER TABLE github_pull_request ADD COLUMN provider TEXT NOT NULL DEFAULT 'github';

-- +migrate Down
ALTER TABLE github_pull_request DROP COLUMN provider;
ALTER TABLE github_installation DROP COLUMN provider;