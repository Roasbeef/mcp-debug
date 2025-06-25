package main

import (
	"fmt"
	"math"
	"time"
)

// calculateFactorial computes the factorial of n
func calculateFactorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

// fibonacci generates the nth Fibonacci number
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// processNumbers performs various calculations on a slice of numbers
func processNumbers(numbers []int) {
	fmt.Println("Processing numbers:", numbers)
	
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	fmt.Printf("Sum: %d\n", sum)
	
	// Calculate average
	if len(numbers) > 0 {
		avg := float64(sum) / float64(len(numbers))
		fmt.Printf("Average: %.2f\n", avg)
	}
	
	// Find min and max
	if len(numbers) > 0 {
		min, max := numbers[0], numbers[0]
		for _, num := range numbers[1:] {
			if num < min {
				min = num
			}
			if num > max {
				max = num
			}
		}
		fmt.Printf("Min: %d, Max: %d\n", min, max)
	}
}

func main() {
	fmt.Println("=== Simple Go Program for Debugging ===")
	
	// Set up some test data
	testNumbers := []int{5, 2, 8, 1, 9, 3}
	
	fmt.Println("\n--- Testing factorial function ---")
	for i := 1; i <= 5; i++ {
		factorial := calculateFactorial(i)
		fmt.Printf("Factorial of %d: %d\n", i, factorial)
	}
	
	fmt.Println("\n--- Testing fibonacci function ---")
	for i := 0; i < 8; i++ {
		fib := fibonacci(i)
		fmt.Printf("Fibonacci(%d): %d\n", i, fib)
	}
	
	fmt.Println("\n--- Processing numbers ---")
	processNumbers(testNumbers)
	
	fmt.Println("\n--- Mathematical operations ---")
	x, y := 15.5, 7.2
	fmt.Printf("x = %.1f, y = %.1f\n", x, y)
	fmt.Printf("x + y = %.1f\n", x+y)
	fmt.Printf("x - y = %.1f\n", x-y)
	fmt.Printf("x * y = %.1f\n", x*y)
	fmt.Printf("x / y = %.2f\n", x/y)
	fmt.Printf("sqrt(x) = %.2f\n", math.Sqrt(x))
	
	fmt.Println("\n--- Simulating some work ---")
	for i := 1; i <= 3; i++ {
		fmt.Printf("Step %d: Working...\n", i)
		time.Sleep(500 * time.Millisecond)
	}
	
	fmt.Println("\n--- Program completed successfully! ---")
}