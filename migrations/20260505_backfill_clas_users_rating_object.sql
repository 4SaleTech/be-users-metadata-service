-- Backfill clas_users.meta_data:
-- Move legacy top-level rating keys into nested object meta_data.rating.*
--
-- From:
-- {"ratings_avg": 1.8, "ratings_dist_1": 3, ...}
-- To:
-- {"rating": {"ratings_avg": 1.8, "ratings_dist_1": 3, ...}}
--
-- Apply:
--   mysql ... < migrations/20260505_backfill_clas_users_rating_object.sql

SET NAMES utf8mb4;

UPDATE clas_users
SET meta_data = JSON_REMOVE(
    JSON_SET(
      COALESCE(meta_data, JSON_OBJECT()),
      '$.rating.ratings_avg',
      COALESCE(
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.ratings_avg'),
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.rating.ratings_avg')
      ),
      '$.rating.ratings_dist_1',
      COALESCE(
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.ratings_dist_1'),
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.rating.ratings_dist_1')
      ),
      '$.rating.ratings_dist_2',
      COALESCE(
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.ratings_dist_2'),
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.rating.ratings_dist_2')
      ),
      '$.rating.ratings_dist_3',
      COALESCE(
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.ratings_dist_3'),
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.rating.ratings_dist_3')
      ),
      '$.rating.ratings_dist_4',
      COALESCE(
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.ratings_dist_4'),
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.rating.ratings_dist_4')
      ),
      '$.rating.ratings_dist_5',
      COALESCE(
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.ratings_dist_5'),
        JSON_EXTRACT(COALESCE(meta_data, JSON_OBJECT()), '$.rating.ratings_dist_5')
      )
    ),
    '$.ratings_avg',
    '$.ratings_dist_1',
    '$.ratings_dist_2',
    '$.ratings_dist_3',
    '$.ratings_dist_4',
    '$.ratings_dist_5'
  )
WHERE JSON_CONTAINS_PATH(
    COALESCE(meta_data, JSON_OBJECT()),
    'one',
    '$.ratings_avg',
    '$.ratings_dist_1',
    '$.ratings_dist_2',
    '$.ratings_dist_3',
    '$.ratings_dist_4',
    '$.ratings_dist_5'
  );
