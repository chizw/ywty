-- 回滚 0004_feedback_violation 迁移
DROP INDEX IF EXISTS idx_violations_status;
DROP INDEX IF EXISTS idx_violations_photo_id;
DROP TABLE IF EXISTS violations;
DROP INDEX IF EXISTS idx_feedbacks_created_at;
DROP INDEX IF EXISTS idx_feedbacks_type;
DROP TABLE IF EXISTS feedbacks;
