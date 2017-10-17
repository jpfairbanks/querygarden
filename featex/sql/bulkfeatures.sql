insert into mimic_v5.features
 (job_id, person_id, cohort_start_date, cohort_end_date, concept_id, value, type)
select ?, person_id, cohort_start_date, cohort_end_date, 'drugs'