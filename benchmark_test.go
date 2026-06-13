package main

import (
	"testing"
)

func BenchmarkInterpreter(b *testing.B) {
	program := `
		println(1 + 1);

		let y = 1;
		let sum = 0;

		while y < 100000 {
			let sum = sum + y;
			let y = y + 1;
		}

		println(sum);
	`
	verbose := false
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		run_program(program, &verbose)
	}
}