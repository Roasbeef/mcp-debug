package calculator

import (
	"testing"
)

func TestAdd(t *testing.T) {
	calc := NewCalculator()
	result := calc.Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %f; want 5", result)
	}
}

func TestSubtract(t *testing.T) {
	calc := NewCalculator()
	result := calc.Subtract(10, 4)
	if result != 6 {
		t.Errorf("Subtract(10, 4) = %f; want 6", result)
	}
}

func TestMultiply(t *testing.T) {
	calc := NewCalculator()
	result := calc.Multiply(3, 4)
	if result != 12 {
		t.Errorf("Multiply(3, 4) = %f; want 12", result)
	}
}

func TestDivide(t *testing.T) {
	calc := NewCalculator()
	result, err := calc.Divide(10, 2)
	if err != nil {
		t.Errorf("Divide(10, 2) returned error: %v", err)
	}
	if result != 5 {
		t.Errorf("Divide(10, 2) = %f; want 5", result)
	}

	_, err = calc.Divide(10, 0)
	if err == nil {
		t.Error("Divide(10, 0) should return error")
	}
}

func TestPower(t *testing.T) {
	calc := NewCalculator()
	result := calc.Power(2, 3)
	if result != 8 {
		t.Errorf("Power(2, 3) = %f; want 8", result)
	}

	result = calc.Power(5, 2)
	if result != 25 {
		t.Errorf("Power(5, 2) = %f; want 25", result)
	}

	result = calc.Power(10, 0)
	if result != 1 {
		t.Errorf("Power(10, 0) = %f; want 1", result)
	}
}

func TestFactorial(t *testing.T) {
	calc := NewCalculator()
	
	tests := []struct {
		input int
		want  int
	}{
		{0, 1},
		{1, 1},
		{5, 120},
		{6, 720},
		{7, 5040},
	}

	for _, tt := range tests {
		result := calc.Factorial(tt.input)
		if result != tt.want {
			t.Errorf("Factorial(%d) = %d; want %d", tt.input, result, tt.want)
		}
	}
}

func TestMemory(t *testing.T) {
	calc := NewCalculator()
	
	calc.Add(5, 3)
	if calc.GetMemory() != 8 {
		t.Errorf("Memory after Add(5, 3) = %f; want 8", calc.GetMemory())
	}

	calc.Multiply(2, 4)
	if calc.GetMemory() != 8 {
		t.Errorf("Memory after Multiply(2, 4) = %f; want 8", calc.GetMemory())
	}

	calc.ClearMemory()
	if calc.GetMemory() != 0 {
		t.Errorf("Memory after Clear = %f; want 0", calc.GetMemory())
	}
}