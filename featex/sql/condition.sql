SELECT
  person_id,
  condition_start_date AS start_date,
  condition_end_date   AS end_date,
  condition_concept_id    concept_id,
  'condition'          AS type,
  1                    AS value
FROM mimic_v5.condition_occurrence
WHERE person_id = $1
ORDER BY person_id, type, start_date, end_date