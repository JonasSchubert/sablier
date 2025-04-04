package config

import "time"

type DynamicStrategy struct {
	CustomThemesPath        string        `mapstructure:"CUSTOM_THEMES_PATH" yaml:"customThemesPath"`
	ShowDetailsByDefault    bool          `mapstructure:"SHOW_DETAILS_BY_DEFAULT" yaml:"showDetailsByDefault"`
	DefaultTheme            string        `mapstructure:"DEFAULT_THEME" yaml:"defaultTheme" default:"hacker-terminal"`
	DefaultRefreshFrequency time.Duration `mapstructure:"DEFAULT_REFRESH_FREQUENCY" yaml:"defaultRefreshFrequency" default:"5s"`
}

type BlockingStrategy struct {
	DefaultTimeout          time.Duration `mapstructure:"DEFAULT_TIMEOUT" yaml:"defaultTimeout" default:"1m"`
	DefaultRefreshFrequency time.Duration `mapstructure:"DEFAULT_REFRESH_FREQUENCY" yaml:"defaultRefreshFrequency" default:"5s"`
}

type Strategy struct {
	Dynamic  DynamicStrategy
	Blocking BlockingStrategy
}

func NewStrategyConfig() Strategy {
	return Strategy{
		Dynamic:  newDynamicStrategy(),
		Blocking: newBlockingStrategy(),
	}
}

func newDynamicStrategy() DynamicStrategy {
	return DynamicStrategy{
		DefaultTheme:            "hacker-terminal",
		ShowDetailsByDefault:    true,
		DefaultRefreshFrequency: 5 * time.Second,
	}
}

func newBlockingStrategy() BlockingStrategy {
	return BlockingStrategy{
		DefaultTimeout:          1 * time.Minute,
		DefaultRefreshFrequency: 5 * time.Second,
	}
}
