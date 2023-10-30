// Code generated by gen.py. DO NOT EDIT.

package expconf

import (
	"github.com/santhosh-tekuri/jsonschema/v2"

	"github.com/determined-ai/determined/cluster/pkg/schemas"
)

func (o OptimizationsConfigV0) AggregationFrequency() int {
	if o.RawAggregationFrequency == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .AggregationFrequency")
	}
	return *o.RawAggregationFrequency
}

func (o *OptimizationsConfigV0) SetAggregationFrequency(val int) {
	o.RawAggregationFrequency = &val
}

func (o OptimizationsConfigV0) AverageAggregatedGradients() bool {
	if o.RawAverageAggregatedGradients == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .AverageAggregatedGradients")
	}
	return *o.RawAverageAggregatedGradients
}

func (o *OptimizationsConfigV0) SetAverageAggregatedGradients(val bool) {
	o.RawAverageAggregatedGradients = &val
}

func (o OptimizationsConfigV0) AverageTrainingMetrics() bool {
	if o.RawAverageTrainingMetrics == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .AverageTrainingMetrics")
	}
	return *o.RawAverageTrainingMetrics
}

func (o *OptimizationsConfigV0) SetAverageTrainingMetrics(val bool) {
	o.RawAverageTrainingMetrics = &val
}

func (o OptimizationsConfigV0) GradientCompression() bool {
	if o.RawGradientCompression == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .GradientCompression")
	}
	return *o.RawGradientCompression
}

func (o *OptimizationsConfigV0) SetGradientCompression(val bool) {
	o.RawGradientCompression = &val
}

func (o OptimizationsConfigV0) GradUpdateSizeFile() *string {
	return o.RawGradUpdateSizeFile
}

func (o *OptimizationsConfigV0) SetGradUpdateSizeFile(val *string) {
	o.RawGradUpdateSizeFile = val
}

func (o OptimizationsConfigV0) MixedPrecision() string {
	if o.RawMixedPrecision == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .MixedPrecision")
	}
	return *o.RawMixedPrecision
}

func (o *OptimizationsConfigV0) SetMixedPrecision(val string) {
	o.RawMixedPrecision = &val
}

func (o OptimizationsConfigV0) TensorFusionThreshold() int {
	if o.RawTensorFusionThreshold == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .TensorFusionThreshold")
	}
	return *o.RawTensorFusionThreshold
}

func (o *OptimizationsConfigV0) SetTensorFusionThreshold(val int) {
	o.RawTensorFusionThreshold = &val
}

func (o OptimizationsConfigV0) TensorFusionCycleTime() int {
	if o.RawTensorFusionCycleTime == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .TensorFusionCycleTime")
	}
	return *o.RawTensorFusionCycleTime
}

func (o *OptimizationsConfigV0) SetTensorFusionCycleTime(val int) {
	o.RawTensorFusionCycleTime = &val
}

func (o OptimizationsConfigV0) AutoTuneTensorFusion() bool {
	if o.RawAutoTuneTensorFusion == nil {
		panic("You must call WithDefaults on OptimizationsConfigV0 before .AutoTuneTensorFusion")
	}
	return *o.RawAutoTuneTensorFusion
}

func (o *OptimizationsConfigV0) SetAutoTuneTensorFusion(val bool) {
	o.RawAutoTuneTensorFusion = &val
}

func (o OptimizationsConfigV0) ParsedSchema() interface{} {
	return schemas.ParsedOptimizationsConfigV0()
}

func (o OptimizationsConfigV0) SanityValidator() *jsonschema.Schema {
	return schemas.GetSanityValidator("http://determined.ai/schemas/expconf/v0/optimizations.json")
}

func (o OptimizationsConfigV0) CompletenessValidator() *jsonschema.Schema {
	return schemas.GetCompletenessValidator("http://determined.ai/schemas/expconf/v0/optimizations.json")
}
