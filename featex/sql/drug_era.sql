SELECT
 person_id,
 drug_era_start_date AS start_date,
 drug_era_end_date   AS end_date,
 drug_concept_id        concept_id,
 'drug_era'          AS type,
 1                   AS value
FROM mimic_v5.drug_era
WHERE person_id = $1
ORDER BY person_id, type, start_date, end_date