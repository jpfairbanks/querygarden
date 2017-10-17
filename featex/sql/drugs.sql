SELECT
 person_id,
 drug_exposure_start_date AS start_date,
 drug_exposure_end_date   AS end_date,
 drug_concept_id             concept_id,
 'drug_exposure'          AS type,
 1                        AS value
FROM mimic_v5.drug_exposure
WHERE person_id = $1
ORDER BY person_id, type, start_date, end_date