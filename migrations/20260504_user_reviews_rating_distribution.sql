-- User-reviews: set rating histogram + average from event payload
--   "data": { "rating": { "1": n1, "2": n2, "3": n3, "4": n4, "5": n5 } }
--
-- RabbitMQ topic name (exchange) for consumers: user-reviews
-- Event types: rating.updated (matches your metadata_rules example) and user.rating.v1 (matches sample JSON "type")
--
-- Apply: mysql ... < migrations/20260504_user_reviews_rating_distribution.sql
-- Re-run safe: deletes actions + rules with the IDs below, then re-inserts.

SET NAMES utf8mb4;

-- Optional: consume from topic exchange named user-reviews (skip error if topic_name already exists)
INSERT INTO event_sources (id, topic_name, enabled, created_at)
VALUES ('aa0e9b7a-2124-11f1-8b87-068146e4f871', 'user-reviews', 1, NOW())
ON DUPLICATE KEY UPDATE enabled = VALUES(enabled);

DELETE FROM metadata_rule_actions WHERE rule_id IN (
  '14e899b7-2124-11f1-8b87-068146e4f871',
  '14e899c7-2124-11f1-8b87-068146e4f871'
);
DELETE FROM metadata_rules WHERE id IN (
  '14e899b7-2124-11f1-8b87-068146e4f871',
  '14e899c7-2124-11f1-8b87-068146e4f871'
);

INSERT INTO metadata_rules (id, event_type, event_version, enabled, priority, description, created_at) VALUES
(
  '14e899b7-2124-11f1-8b87-068146e4f871',
  'rating.updated',
  '',
  1,
  0,
  'Set ratings_dist_* and ratings_avg from data.rating counts (user-reviews)',
  NOW()
),
(
  '14e899c7-2124-11f1-8b87-068146e4f871',
  'user.rating.v1',
  '',
  1,
  0,
  'Set ratings_dist_* and ratings_avg from data.rating counts (user-reviews, same payload as rating.updated)',
  NOW()
);

-- If another rule still matches user.rating.v1 (e.g. old increment-by-scalar), disable it or this rule will combine badly:
-- UPDATE metadata_rules SET enabled = 0 WHERE event_type = 'user.rating.v1' AND id <> '14e899c7-2124-11f1-8b87-068146e4f871';

-- rating.updated — actions
INSERT INTO metadata_rule_actions (id, rule_id, operation, metadata_key, value_source, value_template, condition_expression, execution_order) VALUES
('14e89a01-2124-11f1-8b87-068146e4f801', '14e899b7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_1', 'formula', 'event.data.rating.1', '', 0),
('14e89a02-2124-11f1-8b87-068146e4f802', '14e899b7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_2', 'formula', 'event.data.rating.2', '', 1),
('14e89a03-2124-11f1-8b87-068146e4f803', '14e899b7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_3', 'formula', 'event.data.rating.3', '', 2),
('14e89a04-2124-11f1-8b87-068146e4f804', '14e899b7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_4', 'formula', 'event.data.rating.4', '', 3),
('14e89a05-2124-11f1-8b87-068146e4f805', '14e899b7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_5', 'formula', 'event.data.rating.5', '', 4),
('14e89a06-2124-11f1-8b87-068146e4f806', '14e899b7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_avg', 'formula',
 '(1*metadata.ratings_dist_1+2*metadata.ratings_dist_2+3*metadata.ratings_dist_3+4*metadata.ratings_dist_4+5*metadata.ratings_dist_5)/(metadata.ratings_dist_1+metadata.ratings_dist_2+metadata.ratings_dist_3+metadata.ratings_dist_4+metadata.ratings_dist_5)',
 '', 5);

-- user.rating.v1 — same actions (different rule_id + action ids)
INSERT INTO metadata_rule_actions (id, rule_id, operation, metadata_key, value_source, value_template, condition_expression, execution_order) VALUES
('14e89b01-2124-11f1-8b87-068146e4f901', '14e899c7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_1', 'formula', 'event.data.rating.1', '', 0),
('14e89b02-2124-11f1-8b87-068146e4f902', '14e899c7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_2', 'formula', 'event.data.rating.2', '', 1),
('14e89b03-2124-11f1-8b87-068146e4f903', '14e899c7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_3', 'formula', 'event.data.rating.3', '', 2),
('14e89b04-2124-11f1-8b87-068146e4f904', '14e899c7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_4', 'formula', 'event.data.rating.4', '', 3),
('14e89b05-2124-11f1-8b87-068146e4f905', '14e899c7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_dist_5', 'formula', 'event.data.rating.5', '', 4),
('14e89b06-2124-11f1-8b87-068146e4f906', '14e899c7-2124-11f1-8b87-068146e4f871', 'set', 'ratings_avg', 'formula',
 '(1*metadata.ratings_dist_1+2*metadata.ratings_dist_2+3*metadata.ratings_dist_3+4*metadata.ratings_dist_4+5*metadata.ratings_dist_5)/(metadata.ratings_dist_1+metadata.ratings_dist_2+metadata.ratings_dist_3+metadata.ratings_dist_4+metadata.ratings_dist_5)',
 '', 5);
