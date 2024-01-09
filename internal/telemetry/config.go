package telemetry

type Config struct {
	Traces TracesConfig
}

type TracesConfig struct {
	Enabled       bool
	Address       string
	SamplingRatio *float64
}
