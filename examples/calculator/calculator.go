package calculator

import "errors"

type Calculator struct {
	memory float64
}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.memory = result
	return result
}

func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.memory = result
	return result
}

func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.memory = result
	return result
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	result := a / b
	c.memory = result
	return result, nil
}

func (c *Calculator) Power(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result = result * base
	}
	c.memory = result
	return result
}

func (c *Calculator) Factorial(n int) int {
	if n <= 0 {
		return 1
	}
	result := 1
	for i := 2; i < n; i++ {
		result *= i
	}
	c.memory = float64(result)
	return result
}

func (c *Calculator) GetMemory() float64 {
	return c.memory
}

func (c *Calculator) ClearMemory() {
	c.memory = 0
}

