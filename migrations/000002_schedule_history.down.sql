DROP INDEX IF EXISTS idx_schedule_history_actor_created_at;
DROP INDEX IF EXISTS idx_schedule_history_location_created_at;
DROP INDEX IF EXISTS idx_schedule_history_schedule_created_at;
DROP TABLE IF EXISTS schedule_history;
DROP TYPE IF EXISTS schedule_history_action_enum;
