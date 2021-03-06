select
t.person_id,
start_date,
end_date,
concept_id,
type
from (SELECT
                 person_id,
                 condition_start_date AS start_date,
                 condition_end_date   AS end_date,
                 condition_concept_id    concept_id,
                 'condition'          AS type
               FROM mimic_v5.condition_occurrence
               UNION ALL
               SELECT
                 person_id,
                 drug_exposure_start_date AS start_date,
                 drug_exposure_end_date   AS end_date,
                 drug_concept_id             concept_id,
                 'drug_exposure'          AS type
               FROM mimic_v5.drug_exposure
               UNION ALL
               SELECT
                 person_id,
                 drug_era_start_date AS start_date,
                 drug_era_end_date   AS end_date,
                 drug_concept_id        concept_id,
                 'drug_era'          AS type
               FROM mimic_v5.drug_era
) as t
 INNER JOIN
 mimic_v5.person as pt
 on t.person_id = pt.person_id

-- WHERE
WHERE PT.year_of_birth > 1980
ORDER BY person_id, type, start_date, end_date, type, concept_id