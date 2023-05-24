package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"math"
	"os"
	"strings"
)

type config struct {
	min int
	max int
	muy float64

	from  int
	to    int
	ratio float64
}

func (c config) proportionAt(x float64, sigma float64) float64 {
	dx := x - c.muy
	return math.Exp(-dx * dx / (2.0 * sigma * sigma))
}

func (c config) proportionAtDerivation(x float64, sigma float64) float64 {
	dx := x - c.muy
	exp := math.Exp(-dx * dx / (2.0 * sigma * sigma))
	return exp * dx * dx / (sigma * sigma * sigma)
}

func (c config) computeRatio(sigma float64) float64 {
	sum := 0.0
	for x := c.min; x <= c.max; x++ {
		sum += c.proportionAt(float64(x), sigma)
	}

	a := 0.0
	for x := c.from; x <= c.to; x++ {
		k := c.proportionAt(float64(x), sigma)
		a += k
	}

	return a / sum
}

func (c config) computeRatioDerivation(sigma float64) float64 {
	sum := 0.0
	dsum := 0.0
	for x := c.min; x <= c.max; x++ {
		sum += c.proportionAt(float64(x), sigma)
		dsum += c.proportionAtDerivation(float64(x), sigma)
	}

	a := 0.0
	da := 0.0
	for x := c.from; x <= c.to; x++ {
		a += c.proportionAt(float64(x), sigma)
		da += c.proportionAtDerivation(float64(x), sigma)
	}

	return (da*sum - a*dsum) / (sum * sum)
}

func (c config) findSigma(initSigma float64) float64 {
	sigma := initSigma
	for i := 0; i < 50; i++ {
		sigma = sigma - (c.computeRatio(sigma)-c.ratio)/c.computeRatioDerivation(sigma)
	}
	return sigma
}

func middle(a, b float64) float64 {
	x := math.Log(a)
	y := math.Log(b)
	mid := math.Exp((x + y) / 2.0)
	return mid
}

func (c config) bisectSearch() float64 {
	a := 0.01
	b := 1000000.0

	getSign := func(x float64) bool {
		return math.Signbit(c.computeRatio(x) - c.ratio)
	}

	signA := getSign(a)
	signB := getSign(b)
	if signA == signB {
		fmt.Println("[ERROR] CAN NOT find the solution")
		os.Exit(1)
	}

	for i := 0; i < 20; i++ {
		mid := middle(a, b)
		signMiddle := getSign(mid)
		if signA == signMiddle {
			a = mid
		} else {
			b = mid
		}
	}
	return middle(a, b)
}

func newSimpleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "calculate",
		Run: runSimpleCommand,
	}
	cmd.Flags().Int("min", 10, "minimum discount")
	cmd.Flags().Int("max", 30, "maximum discount")
	cmd.Flags().Float64("mean_value", 10, "mean value")

	cmd.Flags().Int("from", 10, "from discount for the ratio")
	cmd.Flags().Int("to", 30, "to discount for the ratio")
	cmd.Flags().Float64("ratio", 60, "ratio between 'from' and 'to', in percentage")

	return cmd
}

func runSimpleCommand(cmd *cobra.Command, _ []string) {
	min, err := cmd.Flags().GetInt("min")
	if err != nil {
		panic(err)
	}

	max, err := cmd.Flags().GetInt("max")
	if err != nil {
		panic(err)
	}

	muy, err := cmd.Flags().GetFloat64("mean_value")
	if err != nil {
		panic(err)
	}

	from, err := cmd.Flags().GetInt("from")
	if err != nil {
		panic(err)
	}

	to, err := cmd.Flags().GetInt("to")
	if err != nil {
		panic(err)
	}

	ratio, err := cmd.Flags().GetFloat64("ratio")
	if err != nil {
		panic(err)
	}

	fmt.Println("min =", min)
	fmt.Println("max =", max)
	fmt.Println("mean_value =", muy)
	fmt.Println("from =", from)
	fmt.Println("to =", to)
	fmt.Printf("ratio = %.2f%%\n", ratio)

	c := config{
		min: min,
		max: max,
		muy: muy,

		from:  from,
		to:    to,
		ratio: ratio / 100,
	}

	initSigma := c.bisectSearch()
	optimal := c.findSigma(initSigma)
	fmt.Println("Deviation Value =", optimal)
	fmt.Println("==============================================")
	fmt.Println("mean_value (x1000) =", muy*1000)
	fmt.Println("Deviation Value (x1000 & Rounded) =", math.Round(optimal*1000))
}

func newSelectionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "calculate",
		Run: runSelectionCommand,
	}
	cmd.Flags().Int("min", 10, "minimum discount")
	cmd.Flags().Int("max", 30, "maximum discount")
	cmd.Flags().String(
		"mode", "",
		`mode of random distribution, is one of values: `+allModes(),
	)

	return cmd
}

func allModes() string {
	return strings.Join([]string{
		randModeSuperLow,
		randModeLow,
		randModeMedium,
		randModeHigh,
		randModeSuperHigh,
	}, ", ")
}

const (
	randModeSuperLow  = "super_low"
	randModeLow       = "low"
	randModeMedium    = "medium"
	randModeHigh      = "high"
	randModeSuperHigh = "super_high"
)

func exitWithError(msg string) {
	fmt.Println("[ERROR]", msg)
	os.Exit(1)
}

type computeResult struct {
	modeString string

	avg float64

	mean        float64
	meanFormula string

	from        int
	fromFormula string

	to        int
	toFormula string

	ratio float64
}

func computeMeanValue(randMode string, min int, max int) computeResult {
	avg := (float64(min) + float64(max)) / 2

	delta := (float64(max) - float64(min)) / 4.0
	deltaFormula := "[(max-min)/4]"

	switch randMode {
	case randModeSuperLow:
		return computeResult{
			modeString: "Super Low",
			avg:        avg,

			mean:        float64(min),
			meanFormula: "min",

			from:        min,
			fromFormula: "min",

			to:        int(float64(min) + delta),
			toFormula: "min + " + deltaFormula,

			ratio: 95,
		}

	case randModeLow:
		return computeResult{
			modeString: "Low",
			avg:        avg,

			mean:        avg - delta,
			meanFormula: "average - " + deltaFormula,

			from:        min,
			fromFormula: "min",

			to:        int(avg),
			toFormula: "average",

			ratio: 80,
		}

	case randModeMedium:
		return computeResult{
			modeString: "Medium",
			avg:        avg,

			mean:        avg,
			meanFormula: "average",

			from:        int(avg - delta),
			fromFormula: "average - " + deltaFormula,

			to:        int(avg + delta),
			toFormula: "average + " + deltaFormula,

			ratio: 70,
		}

	case randModeHigh:
		return computeResult{
			modeString: "High",
			avg:        avg,

			mean:        avg + delta,
			meanFormula: "average + " + deltaFormula,

			from:        int(avg),
			fromFormula: "average",

			to:        max,
			toFormula: "max",

			ratio: 80,
		}

	case randModeSuperHigh:
		return computeResult{
			modeString: "Super High",
			avg:        avg,

			mean:        float64(max),
			meanFormula: "max",

			from:        int(avg + delta),
			fromFormula: "average + " + deltaFormula,

			to:        max,
			toFormula: "max",

			ratio: 95,
		}

	default:
		exitWithError(`"mode" MUST BE one of values: ` + allModes())
		return computeResult{}
	}
}

func runSelectionCommand(cmd *cobra.Command, _ []string) {
	min, err := cmd.Flags().GetInt("min")
	if err != nil {
		panic(err)
	}

	max, err := cmd.Flags().GetInt("max")
	if err != nil {
		panic(err)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		panic(err)
	}

	result := computeMeanValue(mode, min, max)

	fmt.Println("min =", min)
	fmt.Println("max =", max)
	fmt.Println("mode =", result.modeString)
	fmt.Println("-----------------------------")
	fmt.Println("average = (max-min)/2 =", result.avg)
	fmt.Println("mean_value =", result.meanFormula, "=", result.mean)
	fmt.Println("from =", result.fromFormula, "=", result.from)
	fmt.Println("to =", result.toFormula, "=", result.to)
	fmt.Printf("ratio = %.2f%%\n", result.ratio)
	fmt.Println("-----------------------------")

	c := config{
		min: min,
		max: max,
		muy: result.mean,

		from:  result.from,
		to:    result.to,
		ratio: result.ratio / 100,
	}

	initSigma := c.bisectSearch()
	optimal := c.findSigma(initSigma)
	fmt.Println("Deviation Value =", optimal)
	fmt.Println("==============================================")
	fmt.Println("mean_value (x1000 & Rounded) =", result.mean*1000)
	fmt.Println("Deviation Value (x1000 & Rounded) =", math.Round(optimal*1000))
}

func main() {
	cmd := newSelectionCommand()
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
