package sydney

type Plugin struct {
	Name        string
	OptionsSets []string
	ArgumentPlugin
}

var PluginList = []Plugin{
	{
		Name:        "Suno",
		OptionsSets: []string{"014CB21D"},
		ArgumentPlugin: ArgumentPlugin{
			Id:       "c310c353-b9f0-4d76-ab0d-1dd5e979cf68",
			Category: 1,
		},
	},
}
