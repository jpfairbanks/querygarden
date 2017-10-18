SELECT
 person_id,
 
  drug_era_start_date AS start_date,
  drug_era_end_date   AS end_date,
 
 drug_concept_id             concept_id,
 'drug'          AS type,
 1                        AS value
FROM mimic_v5.drug_era

   -- cohort table : results_mimic.cohort
  RIGHT JOIN results_mimic.cohort on results_mimic.cohort.subject_id = person_id
  -- end

WHERE cohort_definition_id= $1 and ( cohort_end_date > drug_era_start_date ) and not (drug_era_end_date < cohort_start_date)
ORDER BY person_id, type, start_date, end_date