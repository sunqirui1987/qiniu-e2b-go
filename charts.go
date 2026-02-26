package e2b

// ChartType represents the type of a chart
type ChartType string

const (
	// ChartTypeLine is a line chart
	ChartTypeLine ChartType = "line"
	// ChartTypeScatter is a scatter chart
	ChartTypeScatter ChartType = "scatter"
	// ChartTypeBar is a bar chart
	ChartTypeBar ChartType = "bar"
	// ChartTypePie is a pie chart
	ChartTypePie ChartType = "pie"
	// ChartTypeBoxAndWhisker is a box and whisker chart
	ChartTypeBoxAndWhisker ChartType = "box_and_whisker"
	// ChartTypeSuperchart is a superchart (contains multiple charts)
	ChartTypeSuperchart ChartType = "superchart"
	// ChartTypeUnknown is an unknown chart type
	ChartTypeUnknown ChartType = "unknown"
)

// ScaleType represents the scale type for chart axes
type ScaleType string

const (
	// ScaleTypeLinear is a linear scale
	ScaleTypeLinear ScaleType = "linear"
	// ScaleTypeDatetime is a datetime scale
	ScaleTypeDatetime ScaleType = "datetime"
	// ScaleTypeCategorical is a categorical scale
	ScaleTypeCategorical ScaleType = "categorical"
	// ScaleTypeLog is a log scale
	ScaleTypeLog ScaleType = "log"
)

// Chart represents a chart
type Chart struct {
	Type     ChartType `json:"type"`
	Title    string    `json:"title"`
	Elements []any     `json:"elements"`
}

// Chart2D represents a 2D chart
type Chart2D struct {
	Chart
	XLabel string `json:"x_label,omitempty"`
	YLabel string `json:"y_label,omitempty"`
	XUnit  string `json:"x_unit,omitempty"`
	YUnit  string `json:"y_unit,omitempty"`
}

// PointData represents point data for charts
type PointData struct {
	Label  string                `json:"label"`
	Points [][2]any              `json:"points"` // [x, y] pairs
}

// LineChart represents a line chart
type LineChart struct {
	Chart2D
	Type      ChartType   `json:"type"`
	XTicks    []any       `json:"x_ticks"`
	XScale    ScaleType   `json:"x_scale"`
	XTickLabels []string  `json:"x_tick_labels"`
	YTicks    []any       `json:"y_ticks"`
	YScale    ScaleType   `json:"y_scale"`
	YTickLabels []string  `json:"y_tick_labels"`
	Elements  []PointData `json:"elements"`
}

// ScatterChart represents a scatter chart
type ScatterChart struct {
	Chart2D
	Type      ChartType   `json:"type"`
	XTicks    []any       `json:"x_ticks"`
	XScale    ScaleType   `json:"x_scale"`
	XTickLabels []string  `json:"x_tick_labels"`
	YTicks    []any       `json:"y_ticks"`
	YScale    ScaleType   `json:"y_scale"`
	YTickLabels []string  `json:"y_tick_labels"`
	Elements  []PointData `json:"elements"`
}

// BarData represents bar data
type BarData struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Group string `json:"group"`
}

// BarChart represents a bar chart
type BarChart struct {
	Chart2D
	Type     ChartType `json:"type"`
	Elements []BarData `json:"elements"`
}

// PieData represents pie chart data
type PieData struct {
	Label  string  `json:"label"`
	Angle  float64 `json:"angle"`
	Radius float64 `json:"radius"`
}

// PieChart represents a pie chart
type PieChart struct {
	Chart
	Type     ChartType `json:"type"`
	Elements []PieData `json:"elements"`
}

// BoxAndWhiskerData represents box and whisker data
type BoxAndWhiskerData struct {
	Label         string  `json:"label"`
	Min           float64 `json:"min"`
	FirstQuartile float64 `json:"first_quartile"`
	Median        float64 `json:"median"`
	ThirdQuartile float64 `json:"third_quartile"`
	Max           float64 `json:"max"`
	Outliers      []float64 `json:"outliers"`
}

// BoxAndWhiskerChart represents a box and whisker chart
type BoxAndWhiskerChart struct {
	Chart2D
	Type     ChartType           `json:"type"`
	Elements []BoxAndWhiskerData `json:"elements"`
}

// SuperChart represents a chart containing multiple charts
type SuperChart struct {
	Chart
	Type     ChartType `json:"type"`
	Elements []Chart   `json:"elements"`
}

// ChartTypes is a union type for all chart types
type ChartTypes interface {
	*LineChart | *ScatterChart | *BarChart | *PieChart | *BoxAndWhiskerChart | *SuperChart
}
