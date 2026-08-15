ALTER TABLE users
  DROP COLUMN IF EXISTS avatar_mime_type,
  DROP COLUMN IF EXISTS avatar;
