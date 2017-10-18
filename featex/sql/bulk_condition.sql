SELECT
 person_id,
 
  condition_era_start_date AS start_date,
  condition_era_end_date   AS end_date,
 
 condition_concept_id             concept_id,
 'condition'          AS type,
 1                        AS value
FROM mimic_v5.condition_era

   -- cohort table : results_mimic.cohort
  RIGHT JOIN results_mimic.cohort on results_mimic.cohort.subject_id = person_id
  -- end

WHERE cohort_definition_id= $1 and ( cohort_end_date > condition_era_start_date ) -- and not (condition_era_end_date < cohort_start_date)
ORDER BY person_id, type, start_date, end_date