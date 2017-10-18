

-- AGE as binned by OHDSI/FeatureExtraction
select person_id,
        start_date,
        end_date,
        0 as concept_id, 
        age as value,
        'age' as type
  from  (
  SELECT
    person_id,
    cohort_start_date                                    AS start_date,
    cohort_end_date                                      AS end_date,
    date_part('year', cohort_start_date) - year_of_birth AS age
  FROM mimic_v5.person
    RIGHT JOIN results_mimic.cohort ON subject_id = person_id
  WHERE cohort_definition_id = $1
) as t


UNION ALL

-- gender
select person_id, cohort_start_date as start_date, cohort_end_date as end_date,
  gender_concept_id as concept_id, 1 as value, 'gender' as type
from mimic_v5.person
  RIGHT JOIN results_mimic.cohort on subject_id=person_id WHERE cohort_definition_id = $1




UNION ALL
-- race
select person_id, cohort_start_date as start_date, cohort_end_date as end_date,
                  race_concept_id as concept_id, 1 as value, 'race' as type
from mimic_v5.person
  RIGHT JOIN results_mimic.cohort on subject_id=person_id WHERE cohort_definition_id = $1




