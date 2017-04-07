// omop.go: this file contains code that is omop specific.
// OMOP is a specific database common data model that is used in the medical domain
// the feature extraction package is not domain specific, but there
// is some data that only applies to the OMOP CDM and that content goes here.
package featex

var omopqueries map[string]string = make(map[string]string)

// OMOPQueries: the set of all queries that we want to use for feature extraction
// when performing patient level prediciton on an OMOP database.
// These queries apply only in the medical domain.
type OMOPQueries struct {
	demographic string
	drug        string
	procedure   string
	condition   string
}

// OMOPQ is an instance of all the query types for patient level prediction.
// these need to be modified to include the query parameters and cohort joins.
var OMOPQ OMOPQueries = OMOPQueries{demographic: `select
    person.person_id, gender_concept_id as concept_id, 1 as value, 'gender' as type,
  obs.observation_period_start_date as start_date,
  obs.observation_period_end_date as end_date
 from lite_synpuf2.person
  LEFT JOIN lite_synpuf2.observation_period as obs on person.person_id = obs.person_id
  UNION
    select person.person_id, race_concept_id as concept_id, 1 as value, 'race' as type,
  obs.observation_period_start_date as start_date,
  obs.observation_period_end_date as end_date
 from lite_synpuf2.person
  LEFT JOIN lite_synpuf2.observation_period as obs on person.person_id = obs.person_id
  WHERE person.year_of_birth > 1900
  ORDER BY person_id ASC, type ASC, concept_id ASC, value ASC;`,
	drug: `
select person.person_id,
  obs.observation_period_start_date as start_date,
  obs.observation_period_end_date as end_date,
  de.drug_concept_id as concept_id,
  1 as value,
  'drug' as type
from lite_synpuf2.person as person
LEFT JOIN lite_synpuf2.drug_exposure as de on de.person_id = person.person_id
  LEFT JOIN lite_synpuf2.observation_period as obs on person.person_id = obs.person_id
  WHERE person.year_of_birth > 1900
  ORDER BY person_id ASC, type ASC, concept_id ASC, value ASC
`,
	procedure: `
select person.person_id,
  obs.observation_period_start_date as start_date,
  obs.observation_period_end_date as end_date,
  pr.procedure_concept_id as concept_id,
  1 as value,
  'procedure' as type
from lite_synpuf2.person as person
LEFT JOIN lite_synpuf2.procedure_occurrence as pr on pr.person_id = person.person_id
  LEFT JOIN lite_synpuf2.observation_period as obs on person.person_id = obs.person_id
  WHERE person.year_of_birth > 1900
  ORDER BY person_id ASC, type ASC, concept_id ASC, value ASC
`,
	condition: `
select person.person_id,
  obs.observation_period_start_date as start_date,
  obs.observation_period_end_date as end_date,
  pr.condition_concept_id as concept_id,
  1 as value,
  'condition' as type
from lite_synpuf2.person as person
LEFT JOIN lite_synpuf2.condition_occurrence as pr on pr.person_id = person.person_id
  LEFT JOIN lite_synpuf2.observation_period as obs on person.person_id = obs.person_id
  WHERE person.year_of_birth > 1900
  ORDER BY person_id ASC, type ASC, concept_id ASC, value ASC
`}
