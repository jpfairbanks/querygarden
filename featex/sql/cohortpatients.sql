select subject_id as person_id, cohort_start_date as start_date, cohort_end_date as end_date, cohort_definition_id as concept_id, 'cohort_id' as type, 1 as value from results_mimic.cohort
where cohort_definition_id = $1
