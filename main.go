package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"math"
	"os"
)

type config struct {
	min int
	max int
	muy float64

	from  int
	to    int
	ratio float64
}

func (c config) propotionAt(x float64, sigma float64) float64 {
	dx := x - c.muy
	return math.Exp(-dx * dx / (2.0 * sigma * sigma))
}

func (c config) propotionAtDerivation(x float64, sigma float64) float64 {
	dx := x - c.muy
	exp := math.Exp(-dx * dx / (2.0 * sigma * sigma))
	return exp * dx * dx / (sigma * sigma * sigma)
}

func (c config) computeRatio(sigma float64) float64 {
	sum := 0.0
	for x := c.min; x <= c.max; x++ {
		sum += c.propotionAt(float64(x), sigma)
	}

	a := 0.0
	for x := c.from; x <= c.to; x++ {
		k := c.propotionAt(float64(x), sigma)
		a += k
	}

	return a / sum
}

func (c config) computeRatioDerivation(sigma float64) float64 {
	sum := 0.0
	dsum := 0.0
	for x := c.min; x <= c.max; x++ {
		sum += c.propotionAt(float64(x), sigma)
		dsum += c.propotionAtDerivation(float64(x), sigma)
	}

	a := 0.0
	da := 0.0
	for x := c.from; x <= c.to; x++ {
		a += c.propotionAt(float64(x), sigma)
		da += c.propotionAtDerivation(float64(x), sigma)
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

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "calculate",
		Run: runCommand,
	}
	cmd.Flags().Int("min", 10, "minimum discount")
	cmd.Flags().Int("max", 30, "maximum discount")
	cmd.Flags().Float64("mean_value", 10, "mean value")

	cmd.Flags().Int("from", 10, "from discount for the ratio")
	cmd.Flags().Int("to", 30, "to discount for the ratio")
	cmd.Flags().Float64("ratio", 60, "ratio between 'from' and 'to', in percentage")

	return cmd
}

func runCommand(cmd *cobra.Command, _ []string) {
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

func main() {
	cmd := newCommand()
	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
