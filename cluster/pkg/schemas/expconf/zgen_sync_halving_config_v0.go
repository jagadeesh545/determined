// Code generated by gen.py. DO NOT EDIT.

package expconf

import (
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/determined-ai/determined/cluster/pkg/schemas"
)

func (s SyncHalvingConfigV0) NumRungs() int {
	if s.RawNumRungs == nil {
		panic("You must call WithDefaults on SyncHalvingConfigV0 before .NumRungs")
	}
	return *s.RawNumRungs
}

func (s *SyncHalvingConfigV0) SetNumRungs(val int) {
	s.RawNumRungs = &val
}

func (s SyncHalvingConfigV0) MaxLength() LengthV0 {
	if s.RawMaxLength == nil {
		panic("You must call WithDefaults on SyncHalvingConfigV0 before .MaxLength")
	}
	return *s.RawMaxLength
}

func (s *SyncHalvingConfigV0) SetMaxLength(val LengthV0) {
	s.RawMaxLength = &val
}

func (s SyncHalvingConfigV0) Budget() LengthV0 {
	if s.RawBudget == nil {
		panic("You must call WithDefaults on SyncHalvingConfigV0 before .Budget")
	}
	return *s.RawBudget
}

func (s *SyncHalvingConfigV0) SetBudget(val LengthV0) {
	s.RawBudget = &val
}

func (s SyncHalvingConfigV0) Divisor() float64 {
	if s.RawDivisor == nil {
		panic("You must call WithDefaults on SyncHalvingConfigV0 before .Divisor")
	}
	return *s.RawDivisor
}

func (s *SyncHalvingConfigV0) SetDivisor(val float64) {
	s.RawDivisor = &val
}

func (s SyncHalvingConfigV0) TrainStragglers() bool {
	if s.RawTrainStragglers == nil {
		panic("You must call WithDefaults on SyncHalvingConfigV0 before .TrainStragglers")
	}
	return *s.RawTrainStragglers
}

func (s *SyncHalvingConfigV0) SetTrainStragglers(val bool) {
	s.RawTrainStragglers = &val
}

func (s SyncHalvingConfigV0) ParsedSchema() interface{} {
	return schemas.ParsedSyncHalvingConfigV0()
}

func (s SyncHalvingConfigV0) SanityValidator() *jsonschema.Schema {
	return schemas.GetSanityValidator("http://determined.ai/schemas/expconf/v0/searcher-sync-halving.json")
}

func (s SyncHalvingConfigV0) CompletenessValidator() *jsonschema.Schema {
	return schemas.GetCompletenessValidator("http://determined.ai/schemas/expconf/v0/searcher-sync-halving.json")
}
