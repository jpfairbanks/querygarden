SELECT
 person_id,
 
 -- only a single date field so special case
  procedure_date as start_date,
  procedure_date as end_date,
 -- typical case has two date fields
 
 procedure_concept_id             concept_id,
 'procedure'          AS type,
 1                        AS value
FROM mimic_v5.procedure_occurrence

   -- cohort table : results_mimic.cohort
  RIGHT JOIN results_mimic.cohort on results_mimic.cohort.subject_id = person_id
  -- end

WHERE cohort_definition_id= $1
ORDER BY person_id, type, start_date, end_date