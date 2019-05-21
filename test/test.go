package main

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	p, _ := plot.New()

	p.Title.Text = "Hello Price"
	p.X.Label.Text = "Quantity Demand"
	p.Y.Label.Text = "Price"

	points := plotter.XYs{
		{2.0, 60000.0},
		{4.0, 40000.0},
		{6.0, 30000.0},
		{8.0, 25000.0},
		{10.0, 23000.0},
	}

	plotutil.AddLinePoints(p, points)

	p.Save(4*vg.Inch, 4*vg.Inch, "price.png")
}